package web

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/hobeone/tv2go/db"
	"github.com/hobeone/tv2go/naming"
	"github.com/hobeone/tv2go/storage"
	"github.com/hobeone/tv2go/types"
)

// http://wiki.sabnzbd.org/user-scripts
type postprocessReq struct {
	Path       string `json:"path" form:"path" binding:"required"`
	SourceName string `json:"source_name" form:"source_name"`
}

func writeAndFlush(c *gin.Context, str string, args ...interface{}) {
	fmt.Fprintf(c.Writer, str+"\n", args...)
	c.Writer.Flush()
}

func (server *Server) getShowFromName(c *gin.Context, name string) (*db.Show, error) {
	cleanedName := naming.CleanSeriesName(name)
	dbshow, err := server.dbHandle.GetShowByName(cleanedName)
	if err == nil {
		writeAndFlush(c, "Found show with name %s in database.")
		return dbshow, nil
	}
	writeAndFlush(c, "Couldn't find show with name in db: '%s'.", cleanedName)

	sceneName := naming.FullSanitizeSceneName(cleanedName)
	writeAndFlush(c, "Trying to find match in Name Exceptions for %s", sceneName)
	dbshow, err = server.dbHandle.GetShowFromNameException(sceneName)
	if err == nil {
		writeAndFlush(c, "Matched %s to database show %s", sceneName, dbshow.Name)
		return dbshow, nil
	}

	writeAndFlush(c, "Couldn't find show in exception list")

	writeAndFlush(c, "Searching indexers for %s", sceneName)
	for name, idxer := range server.indexers {
		writeAndFlush(c, "Search %s for %s", name, sceneName)
		shows, err := idxer.Search(sceneName)
		if err != nil {
			writeAndFlush(c, "Error searching indexer for %s: %s", sceneName, err)
			continue
		}
		writeAndFlush(c, "Got %d results from indexer %s", len(shows), name)
		for _, show := range shows {
			dbshow, err = server.dbHandle.GetShowByName(show.Name)
			if err != nil {
				writeAndFlush(c, "Error searching DB for show with name %s: %s", show.Name, err)
				continue
			}
			writeAndFlush(c, "Found show in DB with name %s", dbshow.Name)
			// Check that indexerid's match
			//if dbshow.IndexerID != show.IndexerID {
			//	writeAndFlush(c, "Indexer id for dbshow and indexer result differ (%d != %d), skipping.", dbshow.IndexerID, show.IndexerID)
			//}
			return dbshow, nil
		}
	}
	return nil, fmt.Errorf("Couldn't match %s with any known show", name)
}

// Postprocess takes the given path, tries to match it with a show and episode
// and then moves it to where it should go.
func (server *Server) Postprocess(c *gin.Context) {
	var reqJSON postprocessReq

	if !c.Bind(&reqJSON) {
		genError(c, http.StatusBadRequest, c.Errors.String())
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	// Does directory exist and is it readable
	// Get media files from dir
	// for each media file
	//   map to show
	//   map to season/episode
	//   get show & episode
	//   move file
	//   update episode

	if !server.Broker.Readable(reqJSON.Path) {
		c.String(http.StatusBadRequest, "Can't open path '%s'\n", reqJSON.Path)
		return
	}
	mediaFiles, err := storage.MediaFilesInDir(reqJSON.Path)
	if err != nil {
		c.String(http.StatusInternalServerError, "Couldn't get files from %s: %s", reqJSON.Path, err)
		return
	}
	if len(mediaFiles) == 0 {
		writeAndFlush(c, "Couldn't find any media files in '%s'", reqJSON.Path)
		return
	}

	goodresults := []naming.ParseResult{}

	np := naming.NewNameParser(naming.AllRegexes)
	for _, file := range mediaFiles {
		writeAndFlush(c, "Trying to parse %s", file)
		nameres := np.ParseFile(file)
		if nameres.SeriesName == "" {
			writeAndFlush(c, "Couldn't parse series name from %s: skipping", file)
			continue
		}
		if len(nameres.AbsoluteEpisodeNumbers) == 0 && len(nameres.EpisodeNumbers) == 0 {
			writeAndFlush(c, "Couldn't parse episode numbers from '%s'", reqJSON.Path)
			continue
		}
		writeAndFlush(c, "Parsed to %+v", nameres)
		goodresults = append(goodresults, nameres)
	}
	if len(goodresults) == 0 {
		writeAndFlush(c, "Couldn't parse any files in '%s'", reqJSON.Path)
		return
	}

	for _, res := range goodresults {
		dbshow, err := server.getShowFromName(c, res.SeriesName)
		if err != nil {
			writeAndFlush(c, "Couldn't find show with name: '%s'.", res.SeriesName)
			continue
		}

		epnum := res.FirstEpisode()
		dbep, err := server.dbHandle.GetEpisodeByShowSeasonAndNumber(
			dbshow.ID, res.SeasonNumber, epnum)

		if err != nil {
			writeAndFlush(c, "Couldn't find an season/episode for %d, %v, %v", dbshow.ID, res.SeasonNumber, epnum)
			continue
		}

		ext := filepath.Ext(res.OriginalName)
		//loc, err := dbep.GetLocation()
		//Season 01/Show Name-S01E01-Ep Title.ext"
		loc := "Season %02d/%s - S%02dE%02d - %s%s"
		expandedLoc := fmt.Sprintf(loc, dbep.Season, dbshow.Name, dbep.Season, dbep.Episode, dbep.Name, ext)

		expandedLoc = filepath.Join(dbshow.Location, expandedLoc)
		// TODO: import if forced, or quality is equal or better
		err = server.Broker.FileReadable(expandedLoc)
		if err == nil {
			writeAndFlush(c, "File already exists at '%s'", expandedLoc)
			continue
		}

		writeAndFlush(c, "Moving file %s to %s", res.OriginalName, expandedLoc)

		// Change this to provide progress over a channel so we can write progress
		// to the client.

		err = server.Broker.MoveFile(res.OriginalName, expandedLoc)
		if err != nil {
			writeAndFlush(c, "Error moving file to location: %s", err)
			continue
		}

		dbep.Location = expandedLoc
		dbep.Status = types.DOWNLOADED
		err = server.dbHandle.SaveEpisode(dbep)
		if err != nil {
			writeAndFlush(c, "Error saving new episode location: %s", err)
			continue
		}
	}
	c.String(200, "Added episodes.")
}
