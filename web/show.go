package web

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hobeone/tv2go/db"
	"github.com/hobeone/tv2go/storage"
	"github.com/hobeone/tv2go/types"
)

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
			logger.Error("Didn't get episode number from parse result", "name", pr.OriginalName)
			continue
		}
		dbep, err := server.dbHandle.GetEpisodeByShowSeasonAndNumber(showid, pr.SeasonNumber, pr.EpisodeNumbers[0])
		if err != nil {
			logger.Error("Couldn't find episode by show, season, number", "id", showid, "season", pr.SeasonNumber, "episode", pr.EpisodeNumbers[0])
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
