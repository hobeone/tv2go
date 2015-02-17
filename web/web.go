package web

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/hobeone/tv2go/config"
	"github.com/hobeone/tv2go/db"
	"github.com/hobeone/tv2go/indexers/tvdb"
)

type genericResult struct {
	Message string `json:"message"`
	Result  string `json:"result"`
}

func genError(c *gin.Context, status int, msg string) {
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
	Cache         jsonShowCache `json:"cache"`
	Anime         bool          `json:"anime"`
	IndexerID     int64         `json:"indexerid"`
	Language      string        `json:"language"`
	Network       string        `json:"network"`
	NextEpAirdate string        `json:"next_ep_airdate"`
	Paused        bool          `json:"paused"`
	Quality       string        `json:"quality"`
	Name          string        `json:"name"`
	Sports        bool          `json:"sports"`
	Status        string        `json:"status"`
	Subtitles     bool          `json:"subtitles"`
	TVDBID        int64         `json:"tvdbid"`
	TVRageID      int64         `json:"tvrage_id"`
	TVRageName    string        `json:"tvrage_name"`
	SeasonList    []int64       `json:"season_list"`
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
		jsonshows[i] = jsonShow{
			ID:        s.ID,
			AirByDate: s.AirByDate,
			//Cache
			Anime:     s.Anime,
			IndexerID: s.IndexerID,
			Language:  s.Language,
			Network:   s.Network,
			//NextEpAirdate: s.NextEpAirdate(),
			Paused:    s.Paused,
			Quality:   strconv.FormatInt(s.Quality, 10),
			Name:      s.Name,
			Sports:    s.Sports,
			Status:    s.Status,
			Subtitles: s.Subtitles,
			TVDBID:    s.IndexerID,
			//TVdbid, rageid + name
		}
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
	s, err := h.GetShowById(tvdbid)
	if err != nil {
		genError(c, http.StatusNotFound, "Show not found")
		return
	}
	response := jsonShow{
		ID:        s.ID,
		AirByDate: s.AirByDate,
		//Cache
		Anime:     s.Anime,
		IndexerID: s.IndexerID,
		Language:  s.Language,
		Network:   s.Network,
		//NextEpAirdate: s.NextEpAirdate(),
		Paused:     s.Paused,
		Quality:    strconv.FormatInt(s.Quality, 10),
		Name:       s.Name,
		Sports:     s.Sports,
		Status:     s.Status,
		Subtitles:  s.Subtitles,
		TVDBID:     s.IndexerID,
		SeasonList: h.ShowSeasons(s),
		//TVdbid, rageid + name
	}

	c.JSON(http.StatusOK, response)
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
	show, err := h.GetShowById(showid)
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

func episodesToResponse(eps []db.Episode) []episodeResponse {
	resp := make([]episodeResponse, len(eps))
	for i, ep := range eps {
		resp[i] = episodeResponse{
			ID:          ep.ID,
			ShowID:      ep.ShowId,
			AirDate:     ep.AirDateString(),
			Description: ep.Description,
			FileSize:    ep.FileSize,
			Location:    ep.Location,
			Name:        ep.Name,
			Quality:     ep.Quality,
			ReleaseName: ep.ReleaseName,
			Status:      ep.Status,
			Season:      ep.Season,
			Episode:     ep.Episode,
		}
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
	resp := episodeResponse{
		ID:          ep.ID,
		ShowID:      ep.ShowId,
		AirDate:     ep.AirDateString(),
		Description: ep.Description,
		FileSize:    ep.FileSize,
		Location:    ep.Location,
		Name:        ep.Name,
		Quality:     ep.Quality,
		ReleaseName: ep.ReleaseName,
		Status:      ep.Status,
	}
	c.JSON(200, resp)
}

// UpdateEpisode will update the POSTed episode
func (server *Server) UpdateEpisode(c *gin.Context) {
	var epUpdate episodeResponse
	if !c.Bind(&epUpdate) {
		c.JSON(http.StatusBadRequest, genericResult{
			Message: c.Errors.String(),
			Result:  "failure",
		})
		return
	}
	episode := epUpdate
	c.JSON(200, episode)
}

type searchShowRequest struct {
	IndexerName string `form:"indexer_name" binding:"required"`
	SearchTerm  string `form:"name" binding:"required"`
}

// ShowSearch searches for the search term on the given indexer
func (server *Server) ShowSearch(c *gin.Context) {
	var reqJSON searchShowRequest

	if !c.Bind(&reqJSON) {
		c.JSON(http.StatusBadRequest, genericResult{
			Message: c.Errors.String(),
			Result:  "failure",
		})
		return
	}
	series, err := server.tvdbIndexer.Search(reqJSON.SearchTerm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, genericResult{
			Message: err.Error(),
			Result:  "failure",
		})
		return
	}
	c.JSON(http.StatusOK, series)
}

type addShowRequest struct {
	IndexerName string `json:"indexer_name" form:"indexer_name" binding:"required"`
	IndexerID   string `json:"indexer_id" form:"indexer_id" binding:"required"`
}

// AddShow adds the current show to the database.
func (server *Server) AddShow(c *gin.Context) {
	h := server.dbHandle

	var reqJSON addShowRequest

	if !c.Bind(&reqJSON) {
		c.JSON(http.StatusBadRequest, genericResult{
			Message: c.Errors.String(),
			Result:  "failure",
		})
		return
	}

	// Assume TVDB only for now
	// TODO:
	// indexer.GetIndexerFromString(reqJSON.IndexerName)

	indexerID, err := strconv.ParseInt(reqJSON.IndexerID, 10, 64)
	if err != nil {
		c.JSON(500, fmt.Sprintf("Bad indexerid: %s", err.Error()))
		return
	}
	glog.Infof("Got id to add: %s", indexerID)
	dbshow, err := server.tvdbIndexer.GetShow(indexerID)
	if err != nil {
		c.JSON(500, genericResult{
			Message: err.Error(),
			Result:  "failure",
		})
		return
	}
	err = h.AddShow(dbshow)
	if err != nil {
		c.JSON(500, err.Error())
	}
	response := jsonShow{
		ID:        dbshow.ID,
		AirByDate: dbshow.AirByDate,
		//Cache
		Anime:     dbshow.Anime,
		IndexerID: dbshow.IndexerID,
		Language:  dbshow.Language,
		Network:   dbshow.Network,
		//NextEpAirdate: dbshow.NextEpAirdate(),
		Paused:     dbshow.Paused,
		Quality:    strconv.FormatInt(dbshow.Quality, 10),
		Name:       dbshow.Name,
		Sports:     dbshow.Sports,
		Status:     dbshow.Status,
		Subtitles:  dbshow.Subtitles,
		TVDBID:     dbshow.IndexerID,
		SeasonList: h.ShowSeasons(dbshow),
		//TVdbid, rageid + name
	}

	c.JSON(200, response)
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

// StartServer does just what it says.
func StartServer(cfg *config.Config, dbh *db.Handle) {
	s := NewServer(cfg, dbh)
	glog.Fatal(http.ListenAndServe(cfg.WebServer.ListenAddress, s.Handler))
}

// Server contains all the information for the tv2go web server
type Server struct {
	Handler     http.Handler
	config      *config.Config
	tvdbIndexer *tvdb.TvdbIndexer
	dbHandle    *db.Handle
}

func configGinEngine(s *Server) {
	r := gin.New()
	r.Use(Logger())
	r.Use(CORSMiddleware())

	api := r.Group("/api/:apistring")
	{
		api.OPTIONS("/*cors", func(c *gin.Context) {})
		api.GET("shows", s.Shows)
		api.GET("shows/:showid", s.Show)
		api.POST("shows", s.AddShow)

		api.GET("shows/:showid/episodes", s.ShowEpisodes)
		api.GET("shows/:showid/episodes/:episodeid", s.Episode)
		api.PUT("shows/:showid/episodes", s.UpdateEpisode)

		api.GET("indexers/search", s.ShowSearch)
	}

	s.Handler = r
}

// SetTvdbIndexer sets the tvdb index the server should use
func SetTvdbIndexer(t *tvdb.TvdbIndexer) func(*Server) {
	return func(s *Server) {
		s.tvdbIndexer = t
	}
}

// NewServer creates a new server
func NewServer(cfg *config.Config, dbh *db.Handle, options ...func(*Server)) *Server {
	t := &Server{
		dbHandle: dbh,
		config:   cfg,
	}
	configGinEngine(t)
	for _, option := range options {
		option(t)
	}
	return t
}
