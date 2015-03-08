package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/hobeone/tv2go/config"
	"github.com/hobeone/tv2go/db"
	"github.com/hobeone/tv2go/indexers"
	"github.com/hobeone/tv2go/naming"
	"github.com/hobeone/tv2go/providers"
	"github.com/hobeone/tv2go/quality"
	"github.com/hobeone/tv2go/storage"
	"github.com/hobeone/tv2go/types"
)

type genericResult struct {
	Message string `json:"message"`
	Result  string `json:"result"`
}

func genError(c *gin.Context, status int, msg string) {
	glog.Errorf("Error serving %s: %s", c.Request.URL.String(), msg)
	c.JSON(status, genericResult{
		Message: msg,
		Result:  "failure",
	})
}

type jsonShowCache struct {
	Banner int
	Poster int
}
type jsonShow struct {
	ID            int64         `json:"id"`
	AirByDate     bool          `json:"air_by_date"`
	Airs          string        `json:"airs"`
	Cache         jsonShowCache `json:"cache"`
	Anime         bool          `json:"anime"`
	IndexerID     int64         `json:"indexerid"`
	Language      string        `json:"language"`
	Network       string        `json:"network"`
	NextEpAirdate *time.Time    `json:"next_ep_airdate,omitempty"`
	Paused        bool          `json:"paused"`
	QualityGroup  string        `json:"quality_group"`
	Name          string        `json:"name"`
	Sports        bool          `json:"sports"`
	Status        string        `json:"status"`
	Subtitles     bool          `json:"subtitles"`
	TVDBID        int64         `json:"tvdbid"`
	TVRageID      int64         `json:"tvrage_id"`
	TVRageName    string        `json:"tvrage_name"`
	Location      string        `json:"location"`
}

func (server *Server) showToResponse(dbshow *db.Show) jsonShow {
	nextAirDate := server.dbHandle.NextAirdateForShow(dbshow)
	return jsonShow{
		ID:        dbshow.ID,
		AirByDate: dbshow.AirByDate,
		Airs:      dbshow.Airs,
		//Cache
		Anime:         dbshow.Anime,
		IndexerID:     dbshow.IndexerID,
		Language:      dbshow.Language,
		Network:       dbshow.Network,
		NextEpAirdate: nextAirDate,
		Paused:        dbshow.Paused,
		QualityGroup:  dbshow.QualityGroup.Name,
		Name:          dbshow.Name,
		Sports:        dbshow.Sports,
		Status:        dbshow.Status,
		Subtitles:     dbshow.Subtitles,
		TVDBID:        dbshow.IndexerID,
		Location:      dbshow.Location,
		//TVdbid, rageid + name
	}
}

// Shows returns all the shows
func (server *Server) Shows(c *gin.Context) {
	h := server.dbHandle
	shows, err := h.GetAllShows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, "")
	}
	jsonshows := make([]jsonShow, len(shows))
	for i, s := range shows {
		jsonshows[i] = server.showToResponse(&s)
	}
	c.JSON(200, jsonshows)
}

// Show returns just one show
func (server *Server) Show(c *gin.Context) {
	h := server.dbHandle
	id := c.Params.ByName("showid")
	tvdbid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		genError(c, http.StatusInternalServerError, "invalid show id")
		return
	}
	s, err := h.GetShowByID(tvdbid)
	if err != nil {
		genError(c, http.StatusNotFound, "Show not found")
		return
	}
	response := server.showToResponse(s)
	c.JSON(http.StatusOK, response)
}

// UpdateShow will update the POSTed show's information
func (server *Server) UpdateShow(c *gin.Context) {
	var showUpdate jsonShow
	if !c.Bind(&showUpdate) {
		genError(c, http.StatusBadRequest, c.Errors.String())
		return
	}
	dbshow, err := server.dbHandle.GetShowByID(showUpdate.ID)
	if err != nil {
		genError(c, http.StatusBadRequest, fmt.Sprintf("Couldn't find Show %d: %s", showUpdate.ID, err.Error()))
		return
	}

	dbshow.Location = showUpdate.Location
	dbshow.Anime = showUpdate.Anime
	dbshow.Paused = showUpdate.Paused
	dbshow.AirByDate = showUpdate.AirByDate
	server.dbHandle.SaveShow(dbshow)

	c.JSON(200, server.showToResponse(dbshow))
}

// ShowUpdateFromDisk scans the show's location to find the episodes that exist.  It then tries to match these to the episodes in the database.
func (server *Server) ShowUpdateFromDisk(c *gin.Context) {
	id := c.Params.ByName("showid")
	showid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		genError(c, http.StatusInternalServerError, "invalid show id")
		return
	}

	dbshow, err := server.dbHandle.GetShowByID(showid)
	if err != nil {
		genError(c, http.StatusNotFound, "Show not found")
		return
	}

	_, err = os.Stat(dbshow.Location)
	if os.IsNotExist(err) {
		_, err = server.Broker.CreateDir(dbshow.Location)
		if err != nil {
			genError(c, http.StatusInternalServerError, fmt.Sprintf("Show directory didn't exist and got error trying to create it: %s", err.Error()))
			return
		}
	}
	if err != nil {
		genError(c, http.StatusInternalServerError, fmt.Sprintf("Error stating show directory: %s", err.Error()))
		return
	}

	parseRes, err := storage.LoadEpisodesFromDisk(dbshow.Location)
	if err != nil {
		genError(c, http.StatusInternalServerError, fmt.Sprintf("Error loading information from disk: %s", err))
		return
	}

	dbeps := []*db.Episode{}
	for _, pr := range parseRes {
		if len(pr.EpisodeNumbers) == 0 {
			glog.Errorf("Didn't get episode number from parse result for %s", pr.OriginalName)
			continue
		}
		dbep, err := server.dbHandle.GetEpisodeByShowSeasonAndNumber(showid, pr.SeasonNumber, pr.EpisodeNumbers[0])
		if err != nil {
			glog.Errorf("Couldn't find episode by show, season, number: %d, %d, %d", showid, pr.SeasonNumber, pr.EpisodeNumbers[0])
			continue
		}
		dbep.Quality = pr.Quality
		dbep.Location = pr.OriginalName
		dbep.Status = types.DOWNLOADED
		dbeps = append(dbeps, dbep)
	}
	err = server.dbHandle.SaveEpisodes(dbeps)
	if err != nil {
		genError(c, http.StatusInternalServerError, fmt.Sprintf("Error saving episodes: %s", err))
		return
	}
	c.JSON(200, server.showToResponse(dbshow))
}

// ShowUpdateFromIndexer updates show information from the indexer
func (server *Server) ShowUpdateFromIndexer(c *gin.Context) {
	h := server.dbHandle
	id := c.Params.ByName("showid")
	tvdbid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		genError(c, http.StatusInternalServerError, "invalid show id")
		return
	}
	dbshow, err := h.GetShowByID(tvdbid)
	if err != nil {
		genError(c, http.StatusNotFound, "Show not found")
		return
	}

	if _, ok := server.indexers[dbshow.Indexer]; !ok {
		genError(c, http.StatusInternalServerError, fmt.Sprintf("Unknown indexer '%s'", dbshow.Indexer))
		return
	}

	err = server.indexers[dbshow.Indexer].UpdateShow(dbshow)
	if err != nil {
		genError(c, http.StatusInternalServerError, fmt.Sprintf("Error updating show: %s", err.Error()))
		return
	}
	showDir, err := server.createShowDirectory(dbshow)
	if err != nil {
		c.JSON(500, fmt.Sprintf("Error creating show directory: %s", err.Error()))
		return
	}
	dbshow.Location = showDir

	h.SaveShow(dbshow)

	c.JSON(200, server.showToResponse(dbshow))
}

// ShowEpisodes returns all of a shows episodes
func (server *Server) ShowEpisodes(c *gin.Context) {
	h := server.dbHandle
	id := c.Params.ByName("showid")
	showid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		genError(c, http.StatusInternalServerError, "invalid show id")
		return
	}
	show, err := h.GetShowByID(showid)
	if err != nil {
		genError(c, http.StatusNotFound, "Show not found")
		return
	}
	eps, err := h.GetShowEpisodes(show)
	if err != nil {
		genError(c, http.StatusInternalServerError, "Couldnt get episodes for show")
	}

	resp := episodesToResponse(eps)

	c.JSON(200, resp)
}

func episodeToResponse(ep *db.Episode) episodeResponse {
	return episodeResponse{
		ID:          ep.ID,
		ShowID:      ep.ShowId,
		AirDate:     ep.AirDateString(),
		Description: ep.Description,
		FileSize:    ep.FileSize,
		Location:    ep.Location,
		Name:        ep.Name,
		Quality:     ep.Quality.String(),
		ReleaseName: ep.ReleaseName,
		Status:      ep.Status.String(),
		Season:      ep.Season,
		Episode:     ep.Episode,
	}
}

func episodesToResponse(eps []db.Episode) []episodeResponse {
	resp := make([]episodeResponse, len(eps))
	for i, ep := range eps {
		resp[i] = episodeToResponse(&ep)
	}
	return resp
}

type episodeResponse struct {
	ID            int64  `json:"id" form:"id" binding:"required"`
	ShowID        int64  `json:"showid" form:"showid" binding:"required"`
	Name          string `json:"name" form:"name" binding:"required"`
	Season        int64  `json:"season" form:"season"`
	Episode       int64  `json:"episode" form:"episode"`
	AirDate       string `json:"airdate" form:"airdate"`
	Description   string `json:"description" form:"description"`
	FileSize      int64  `json:"file_size" form:"file_size"`
	FileSizeHuman string `json:"file_size_human" form:"file_size_human"`
	Location      string `json:"location" form:"location"`
	Quality       string `json:"quality" form:"quality"`
	ReleaseName   string `json:"release_name" form:"release_name"`
	Status        string `json:"status" form:"status"`
}

// Episode returns just one episode
func (server *Server) Episode(c *gin.Context) {
	h := server.dbHandle
	episodeid, err := strconv.ParseInt(c.Params.ByName("episodeid"), 10, 64)

	if err != nil {
		c.JSON(http.StatusNotFound, genericResult{
			Message: fmt.Sprintf("Invalid episodeid: %v", c.Params.ByName("episodeid")),
			Result:  "failure",
		})
		return
	}

	ep, err := h.GetEpisodeByID(episodeid)
	if err != nil {
		c.JSON(http.StatusNotFound, genericResult{
			Message: err.Error(),
			Result:  "failure",
		})
		return
	}
	resp := episodeToResponse(ep)
	c.JSON(200, resp)
}

// UpdateEpisode will update the POSTed episode's status
func (server *Server) UpdateEpisode(c *gin.Context) {
	var epUpdate episodeResponse
	if !c.Bind(&epUpdate) {
		c.JSON(http.StatusBadRequest, genericResult{
			Message: c.Errors.String(),
			Result:  "failure",
		})
		return
	}
	dbep, err := server.dbHandle.GetEpisodeByID(epUpdate.ID)
	if err != nil {
		genError(c, http.StatusBadRequest, fmt.Sprintf("Couldn't find Episode %d", epUpdate.ID))
		return
	}

	stat, err := types.EpisodeStatusFromString(epUpdate.Status)
	if err != nil {
		genError(c, http.StatusBadRequest, fmt.Sprintf("Invalid Status %s", epUpdate.Status))
		return
	}

	dbep.Status = stat
	server.dbHandle.SaveEpisode(dbep)

	c.JSON(200, episodeToResponse(dbep))
}

type episodeSearchReq struct {
	ShowName     string `form:"show_name" binding:"required"`
	EpisodeNum   string `form:"episode_number" binding:"required"`
	SeasonNumber string `form:"season_number" binding:"required"`
}

type episodeSearchResp struct {
	SourceType  string `json:"source_type"`
	Age         string `json:"age"`
	Title       string `json:"title"`
	IndexerName string `json:"indexer_name"`
	Size        int64  `json:"size"`
	Peers       string `json:"peers"`
	Quality     string `json:"quality"`
}

// EpisodeSearch searches configured Providers for episode files.
func (server *Server) EpisodeSearch(c *gin.Context) {
	h := server.dbHandle
	episodeid, err := strconv.ParseInt(c.Params.ByName("episodeid"), 10, 64)

	if err != nil {
		genError(c, http.StatusNotFound, fmt.Sprintf("Invalid episodeid: %v", c.Params.ByName("episodeid")))
		return
	}

	ep, err := h.GetEpisodeByID(episodeid)
	if err != nil {
		genError(c, http.StatusNotFound, err.Error())
		return
	}

	res := server.Providers.Search(ep.Show.Name, ep.Season, ep.Episode)
	//res, err := server.Providers["nzbsOrg"].TvSearch(ep.Show.Name, ep.Season, ep.Episode)
	if len(res) == 0 {
		genError(c, http.StatusNotFound, fmt.Sprintf("No results found for show: %s", ep.Show.Name))
		return
	}
	c.JSON(200, res)
}

type downloadReq struct {
	Provider string `form:"provider" binding:"required"`
	URL      string `form:"url" binding:"required"`
}

// DownloadEpisode takes a request to download a episode from a provider
func (server *Server) DownloadEpisode(c *gin.Context) {
	episodeid, err := strconv.ParseInt(c.Params.ByName("episodeid"), 10, 64)

	var reqJSON downloadReq

	if !c.Bind(&reqJSON) {
		genError(c, http.StatusBadRequest, c.Errors.String())
		return
	}

	if err != nil {
		genError(c, http.StatusNotFound, fmt.Sprintf("Invalid episodeid: %v", c.Params.ByName("episodeid")))
		return
	}

	ep, err := server.dbHandle.GetEpisodeByID(episodeid)
	if err != nil {
		genError(c, http.StatusNotFound, err.Error())
		return
	}
	ep.Status = types.SNATCHED

	prov, ok := server.Providers[reqJSON.Provider]
	if !ok {
		genError(c, http.StatusBadRequest, fmt.Sprintf("Unknown provider: %s", reqJSON.Provider))
		return
	}

	filename, filebytes, err := prov.GetURL(reqJSON.URL)

	if err != nil {
		genError(c, http.StatusInternalServerError, fmt.Sprintf("Error getting file: %s", err.Error()))
		return
	}

	dltype := prov.Type()
	destdir := ""
	switch dltype {
	case providers.NZB:
		destdir = server.config.Storage.NZBBlackhole
	case providers.TORRENT:
		destdir = server.config.Storage.TorrentBlackhole
	case providers.UNKNOWN:
		genError(c, http.StatusInternalServerError, fmt.Sprintf("Unknown provider type: %d", prov.Type()))
		return
	}

	dstfile, err := server.Broker.SaveToFile(destdir, filename, filebytes)
	if err != nil {
		genError(c, http.StatusInternalServerError, fmt.Sprintf("Error saving file: %s", err))
		return
	}
	server.dbHandle.SaveEpisode(ep)
	c.JSON(200, fmt.Sprintf("Downloaded %s from %s to %s", reqJSON.URL, reqJSON.Provider, dstfile))
}

type searchShowRequest struct {
	IndexerName string `form:"indexer_name" binding:"required"`
	SearchTerm  string `form:"name" binding:"required"`
}

// ShowSearch searches for the search term on the given indexer
func (server *Server) ShowSearch(c *gin.Context) {
	var reqJSON searchShowRequest

	if !c.Bind(&reqJSON) {
		genError(c, http.StatusBadRequest, c.Errors.String())
		return
	}

	if _, ok := server.indexers[reqJSON.IndexerName]; !ok {
		genError(c, http.StatusBadRequest, fmt.Sprintf("Unknown indexer: '%s'", reqJSON.IndexerName))
		return
	}

	series, err := server.indexers[reqJSON.IndexerName].Search(reqJSON.SearchTerm)
	if err != nil {
		genError(c, http.StatusInternalServerError, fmt.Sprintf("Error searching for show on %s: %s", reqJSON.IndexerName, err))
		return
	}
	resp := make([]jsonShow, len(series))
	for i, s := range series {
		resp[i] = server.showToResponse(&s)
	}
	c.JSON(http.StatusOK, resp)
}

type addShowRequest struct {
	IndexerName   string `json:"indexer_name" form:"indexer_name" binding:"required"`
	IndexerID     string `json:"indexerid" form:"indexerid" binding:"required"`
	QualityGroup  string `json:"quality_group"`
	EpisodeStatus string `json:"episode_status"`
	Location      string `json:"location"`
	Anime         bool   `json:"is_anime"`
	AirByDate     bool   `json:"is_air_by_date"`
}

// AddShow adds the current show to the database.
func (server *Server) AddShow(c *gin.Context) {
	h := server.dbHandle

	var reqJSON addShowRequest

	if !c.Bind(&reqJSON) {
		genError(c, http.StatusBadRequest, c.Errors.String())
		return
	}
	epStatus := server.config.MediaDefaults.EpisodeStatus
	var err error
	if reqJSON.EpisodeStatus != "" {
		epStatus, err = types.EpisodeStatusFromString(reqJSON.EpisodeStatus)
		if err != nil {
			genError(c, http.StatusBadRequest, fmt.Sprintf("Unknown EpisodeStatus string: %s", c.Errors.String()))
			return
		}
	}
	showQuality := h.GetQualityGroupFromStringOrDefault(reqJSON.QualityGroup)

	if _, ok := server.indexers[reqJSON.IndexerName]; !ok {
		genError(c, http.StatusBadRequest, fmt.Sprintf("Unknown indexer: '%s'", reqJSON.IndexerName))
		return
	}
	indexerID, err := strconv.ParseInt(reqJSON.IndexerID, 10, 64)
	if err != nil {
		c.JSON(500, fmt.Sprintf("Bad indexerid: %s", err.Error()))
		return
	}
	glog.Infof("Got id to add: %s", indexerID)
	// TODO: lame
	dbshow, err := server.indexers[reqJSON.IndexerName].GetShow(strconv.FormatInt(indexerID, 10))
	if err != nil {
		c.JSON(500, genericResult{
			Message: err.Error(),
			Result:  "failure",
		})
		return
	}
	dbshow.QualityGroup = *showQuality
	dbshow.Anime = reqJSON.Anime
	dbshow.AirByDate = reqJSON.AirByDate
	for i := range dbshow.Episodes {
		dbshow.Episodes[i].Status = epStatus
		dbshow.Episodes[i].Quality = quality.UNKNOWN
	}
	if dbshow.Location == "" {
		dbshow.Location = showToLocation(server.Broker.RootDirs[0], dbshow.Name)
		glog.Infof("Location not set on show %s: defaulting to %s", dbshow.Name, dbshow.Location)
	}

	err = h.AddShow(dbshow)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}

	_, err = server.createShowDirectory(dbshow)
	if err != nil {
		c.JSON(500, fmt.Sprintf("Error creating show directory: %s", err.Error()))
		return
	}

	c.JSON(200, server.showToResponse(dbshow))
}

// http://wiki.sabnzbd.org/user-scripts
type postprocessReq struct {
	Path       string `json:"path" form:"path" binding:"required"`
	SourceName string `json:"source_name" form:"source_name"`
}

func writeAndFlush(c *gin.Context, str string, args ...interface{}) {
	fmt.Fprintf(c.Writer, str, args)
	c.Writer.Flush()
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

	goodresults := []naming.ParseResult{}

	np := naming.NewNameParser("", naming.StandardRegexes)
	for _, file := range mediaFiles {
		nameres := np.Parse(file)
		if nameres.SeriesName == "" {
			writeAndFlush(c, "Couldn't parse series name from %s: skipping\n", file)
			continue
		}
		if len(nameres.AbsoluteEpisodeNumbers) == 0 && len(nameres.EpisodeNumbers) == 0 {
			writeAndFlush(c, "Could parse episode numbers from '%s'\n", reqJSON.Path)
			continue
		}
		goodresults = append(goodresults, nameres)
	}

	if len(goodresults) == 0 {
		writeAndFlush(c, "Couldn't find any media files in '%s'\n", reqJSON.Path)
		return
	}

	for _, res := range goodresults {
		cleanedName := naming.CleanSeriesName(res.SeriesName)
		dbshow, err := server.dbHandle.GetShowByName(cleanedName)
		if err != nil {
			writeAndFlush(c, "Couldn't find show with name: '%s'. Skipping.\n", cleanedName)
			continue
		}

		epnum := res.EpisodeNumbers[0]
		dbep, err := server.dbHandle.GetEpisodeByShowSeasonAndNumber(
			dbshow.ID, res.SeasonNumber, epnum)

		if err != nil {
			writeAndFlush(c, "Could find an season/episode for %v, %v\n", res.SeasonNumber, epnum)
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
			writeAndFlush(c, "File already exists at '%s'\n", expandedLoc)
			continue
		}

		err = server.Broker.MoveFile(reqJSON.Path, expandedLoc)
		if err != nil {
			writeAndFlush(c, "Error moving file to location: %s\n", err)
			continue
		}

		dbep.Location = expandedLoc
		dbep.Status = types.DOWNLOADED
		err = server.dbHandle.SaveEpisode(dbep)
		if err != nil {
			writeAndFlush(c, "Error saving new episode location: %s\n", err)
			continue
		}
	}
	c.String(200, "Added episodes.")

}

func showToLocation(path, name string) string {
	name = strings.Trim(name, " ")
	name = strings.Trim(name, ".")

	// Replace certain joining characters with a dash
	seps := regexp.MustCompile(`[\\/\*]`)
	name = seps.ReplaceAllString(name, "-")
	seps = regexp.MustCompile(`[:"<>|?]`)
	name = seps.ReplaceAllString(name, "")

	// Remove all other unrecognised characters - NB we do allow any printable characters
	legal := regexp.MustCompile(`[^[:alnum:]-.() ]`)
	name = legal.ReplaceAllString(name, "_")

	// Remove any double dashes caused by existing - in name
	name = strings.Replace(name, "--", "-", -1)

	newpath := filepath.Join(path, name)
	return newpath
}

func (server *Server) createShowDirectory(dbshow *db.Show) (string, error) {
	if dbshow.Location == "" {
		return "", errors.New("Show location not set")
	}

	createdDir, err := server.Broker.CreateDir(dbshow.Location)
	if err != nil {
		return "", fmt.Errorf("Error creating show directory: %s", err.Error())
	}
	return createdDir, nil
}

// DBHandler makes a database connection available to other handlers
func DBHandler(dbh *db.Handle) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("dbh", dbh)
		c.Next()
	}
}

// Logger provides a Logging middleware using glog
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		// before request
		c.Next()

		// after request
		end := time.Now()
		latency := end.Sub(t)

		glog.Infof("[GIN] |%3d| %12v | %s |%-7s %s\n%s",
			c.Writer.Status(),
			latency,
			c.ClientIP(),
			c.Request.Method,
			c.Request.URL.RequestURI(),
			c.Errors.String(),
		)
	}
}

// CORSMiddleware adds the right headers to make external API requrests happy
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
		}
	}
}

/*
*
* API:
*
* Base url: /api/APIKEY/
*
* Show:
*
* GET shows/ - all shows
* GET shows/:show_id - one show
* PUT shows/ - update show
* DELETE shows/:show_id - delete one show
*
* GET shows/:show_id/episodes/ - all episodes for show
* GET shows/:show_id/episodes/:episode_id - one episode
*
* TODO: settings, indexers, providers
 */

// StartServing does just what it says.
func (server *Server) StartServing() {
	glog.Fatal(
		http.ListenAndServe(
			server.config.WebServer.ListenAddress, server.Handler,
		),
	)
}

// Server contains all the information for the tv2go web server
type Server struct {
	Handler   http.Handler
	config    *config.Config
	Broker    *storage.Broker
	Providers providers.ProviderRegistry
	indexers  indexers.IndexerRegistry
	dbHandle  *db.Handle
}

func configGinEngine(s *Server) {
	r := gin.New()
	r.Use(Logger())

	r.Static("/a", "./webapp")

	api := r.Group("/api/:apistring")
	{
		api.Use(CORSMiddleware())
		api.OPTIONS("/*cors", func(c *gin.Context) {})
		api.GET("shows", s.Shows)
		api.GET("shows/:showid", s.Show)
		api.PUT("shows/:showid", s.UpdateShow)
		api.GET("shows/:showid/update", s.ShowUpdateFromIndexer)
		api.GET("shows/:showid/rescan", s.ShowUpdateFromDisk)
		api.POST("shows", s.AddShow)

		api.GET("shows/:showid/episodes", s.ShowEpisodes)
		api.GET("shows/:showid/episodes/:episodeid", s.Episode)
		api.GET("shows/:showid/episodes/:episodeid/search", s.EpisodeSearch)
		api.POST("shows/:showid/episodes/:episodeid/download", s.DownloadEpisode)
		api.PUT("shows/:showid/episodes", s.UpdateEpisode)

		api.GET("indexers/search", s.ShowSearch)
		api.GET("indexers", s.IndexerList)
		api.GET("statuses", s.StatusList)
		api.GET("quality_groups", s.QualityGroupList)

		api.POST("postprocess", s.Postprocess)
	}

	r.GET("/statusz", s.Statusz)

	s.Handler = r
}

//QualityGroupList serves a list of all known QualityGroups
func (server *Server) QualityGroupList(c *gin.Context) {
	qualityGroups, err := server.dbHandle.GetQualityGroups()
	if err != nil {
		genError(c, http.StatusInternalServerError, fmt.Sprintf("Error getting QualityGroups: %s", err))
		return
	}
	c.JSON(200, qualityGroups)
}

//StatusList returns all of the known Episode Statuses
func (server *Server) StatusList(c *gin.Context) {
	typeStrs := make([]string, len(types.EpisodeDefaults))
	for i, v := range types.EpisodeDefaults {
		typeStrs[i] = v.String()
	}
	c.JSON(200, typeStrs)
}

// IndexerList serves a list of all the known indexers
func (server *Server) IndexerList(c *gin.Context) {
	res := make([]string, len(server.indexers))
	i := 0
	for k := range server.indexers {
		res[i] = k
		i++
	}
	c.JSON(200, res)
}

// Statusz serves internal server information in JSON format
func (server *Server) Statusz(c *gin.Context) {
	marsh, err := json.MarshalIndent(server.config, "", "  ")
	if err != nil {
		genError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.Writer.Write(marsh)
}

// SetIndexers sets the tvdb index the server should use
func SetIndexers(idxs indexers.IndexerRegistry) func(*Server) {
	return func(s *Server) {
		s.indexers = idxs
	}
}

// NewServer creates a new server
func NewServer(cfg *config.Config, dbh *db.Handle, broker *storage.Broker, provReg providers.ProviderRegistry, options ...func(*Server)) *Server {
	t := &Server{
		dbHandle:  dbh,
		config:    cfg,
		Broker:    broker,
		Providers: provReg,
	}
	gin.SetMode("debug")
	configGinEngine(t)
	for _, option := range options {
		option(t)
	}
	return t
}
