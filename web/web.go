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

	"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/hobeone/tv2go/config"
	"github.com/hobeone/tv2go/db"
	"github.com/hobeone/tv2go/indexers"
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
	Location      string        `json:"location"`
}

func showToResponse(dbshow *db.Show) jsonShow {
	return jsonShow{
		ID:        dbshow.ID,
		AirByDate: dbshow.AirByDate,
		//Cache
		Anime:     dbshow.Anime,
		IndexerID: dbshow.IndexerID,
		Language:  dbshow.Language,
		Network:   dbshow.Network,
		//NextEpAirdate: dbshow.NextEpAirdate(),
		Paused:    dbshow.Paused,
		Quality:   dbshow.Quality.String(),
		Name:      dbshow.Name,
		Sports:    dbshow.Sports,
		Status:    dbshow.Status,
		Subtitles: dbshow.Subtitles,
		TVDBID:    dbshow.IndexerID,
		Location:  dbshow.Location,
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
		jsonshows[i] = showToResponse(&s)
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
	response := showToResponse(s)
	c.JSON(http.StatusOK, response)
}

// UpdateShow will update the POSTed show's information
func (server *Server) UpdateShow(c *gin.Context) {
	var showUpdate jsonShow
	if !c.Bind(&showUpdate) {
		genError(c, http.StatusBadRequest, c.Errors.String())
		return
	}
	dbshow, err := server.dbHandle.GetShowById(showUpdate.ID)
	if err != nil {
		genError(c, http.StatusBadRequest, fmt.Sprintf("Couldn't find Show %d: %s", showUpdate.ID, err.Error()))
		return
	}

	dbshow.Location = showUpdate.Location
	dbshow.Anime = showUpdate.Anime
	dbshow.Paused = showUpdate.Paused
	dbshow.AirByDate = showUpdate.AirByDate
	server.dbHandle.SaveShow(dbshow)

	c.JSON(200, showToResponse(dbshow))
}

// ShowUpdateFromDisk scans the show's location to find the episodes that exist.  It then tries to match these to the episodes in the database.
func (server *Server) ShowUpdateFromDisk(c *gin.Context) {
	id := c.Params.ByName("showid")
	showid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		genError(c, http.StatusInternalServerError, "invalid show id")
		return
	}

	dbshow, err := server.dbHandle.GetShowById(showid)
	if err != nil {
		genError(c, http.StatusNotFound, "Show not found")
		return
	}
	parseRes, err := storage.LoadEpisodesFromDisk(dbshow.Location)
	if err != nil {
		genError(c, http.StatusInternalServerError, fmt.Sprintf("Error loading information from disk: %s", err))
		return
	}

	dbeps := []*db.Episode{}
	for _, pr := range parseRes {
		spew.Dump(pr)
		if len(pr.EpisodeNumbers) == 0 {
			glog.Errorf("Didn't get episode number from parse result for %s", pr.OriginalName)
			continue
		}
		dbep, err := server.dbHandle.GetEpisodeByShowSeasonAndNumber(showid, pr.SeasonNumber, pr.EpisodeNumbers[0])
		if err != nil {
			glog.Errorf("Couldn't find episode by show, season, number: %d, %d, %d", showid, pr.SeasonNumber, pr.EpisodeNumbers[0])
			continue
		}
		dbep.Location = pr.OriginalName
		dbeps = append(dbeps, dbep)
	}
	server.dbHandle.SaveEpisodes(dbeps)
	c.JSON(200, showToResponse(dbshow))
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
	dbshow, err := h.GetShowById(tvdbid)
	if err != nil {
		genError(c, http.StatusNotFound, "Show not found")
		return
	}

	err = server.indexers["tvdb"].UpdateShow(dbshow)
	if err != nil {
		genError(c, http.StatusInternalServerError, fmt.Sprintf("Error updating show: %s", err.Error()))
		return
	}

	h.SaveShow(dbshow)

	err = createShowDirectory(dbshow)
	if err != nil {
		c.JSON(500, fmt.Sprintf("Error creating show directory: %s", err.Error()))
		return
	}

	c.JSON(200, showToResponse(dbshow))
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
	series, err := server.indexers["tvdb"].Search(reqJSON.SearchTerm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, genericResult{
			Message: err.Error(),
			Result:  "failure",
		})
		return
	}
	resp := make([]jsonShow, len(series))
	for i, s := range series {
		resp[i] = showToResponse(&s)
	}
	c.JSON(http.StatusOK, resp)
}

type addShowRequest struct {
	IndexerName   string `json:"indexer_name" form:"indexer_name" binding:"required"`
	IndexerID     string `json:"indexerid" form:"indexerid" binding:"required"`
	ShowQuality   string `json:"show_quality"`
	EpisodeStatus string `json:"episode_status"`
	Location      string `json:"location"`
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
	epStatus := server.config.MediaDefaults.EpisodeStatus
	var err error
	if reqJSON.EpisodeStatus != "" {
		epStatus, err = types.EpisodeStatusFromString(reqJSON.EpisodeStatus)
		if err != nil {
			c.JSON(http.StatusBadRequest, genericResult{
				Message: fmt.Sprintf("Unknown EpisodeStatus string: %s", c.Errors.String()),
				Result:  "failure",
			})
			return
		}
	}
	showQuality := server.config.MediaDefaults.ShowQuality
	if reqJSON.ShowQuality != "" {
		showQuality, err = types.QualityFromString(reqJSON.ShowQuality)
		if err != nil {
			c.JSON(http.StatusBadRequest, genericResult{
				Message: fmt.Sprintf("Unknown Quality string: %s", c.Errors.String()),
				Result:  "failure",
			})
			return
		}
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
	// TODO: lame
	dbshow, err := server.indexers["tvdb"].GetShow(strconv.FormatInt(indexerID, 10))
	if err != nil {
		c.JSON(500, genericResult{
			Message: err.Error(),
			Result:  "failure",
		})
		return
	}
	dbshow.Quality = showQuality
	for i := range dbshow.Episodes {
		dbshow.Episodes[i].Status = epStatus
		dbshow.Episodes[i].Quality = types.NONE
	}

	if dbshow.Location == "" {
		dbshow.Location = showToLocation(server.config.Storage.Directories[0], dbshow.Name)
	}

	err = h.AddShow(dbshow)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}

	err = createShowDirectory(dbshow)
	if err != nil {
		c.JSON(500, fmt.Sprintf("Error creating show directory: %s", err.Error()))
		return
	}

	c.JSON(200, showToResponse(dbshow))
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
	legal := regexp.MustCompile(`[^[:alnum:]-. ]`)
	name = legal.ReplaceAllString(name, "_")

	// Remove any double dashes caused by existing - in name
	name = strings.Replace(name, "--", "-", -1)

	newpath := filepath.Join(path, name)
	return newpath
}

func createShowDirectory(dbshow *db.Show) error {
	if dbshow.Location == "" {
		return errors.New("Show location not set")
	}

	err := os.MkdirAll(dbshow.Location, 0755)
	if err != nil {
		fmt.Errorf("Error creating show directory: %s", err.Error())
	}
	return nil
}

func rescanShowFromDisk(dbshow *db.Show) error {
	if dbshow.Location == "" {
		return errors.New("Show location not set")
	}

	_, err := os.Stat(dbshow.Location)
	if os.IsNotExist(err) {
		err = createShowDirectory(dbshow)
		if err != nil {
			return fmt.Errorf("Show directory didn't exist and got error trying to create it: %s", err.Error())
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error stating show directory: %s", err.Error())
	}

	walkfunc := func(path string, f os.FileInfo, err error) error {
		if err != nil {
			glog.Errorf("Got error when walking path %s: %s", path, err.Error())
		}
		if f.IsDir() {
			return nil
		}
		/*
		* Assumed File Structure:
		* ShowName/Season 01/ShowName - S01E01 - Episode Name.mkv
		*
		* OR
		*
		* ShowName/Episode 01-Name.mkv
		 */
		p, err := filepath.Rel(dbshow.Location, path)
		spew.Dump(p)
		spew.Dump(err)
		//parseFileNameToEpisode(path)
		return nil
	}

	filepath.Walk(dbshow.Location, walkfunc)

	return nil
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
	Handler  http.Handler
	config   *config.Config
	indexers indexers.IndexerRegistry
	dbHandle *db.Handle
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
		api.PUT("shows/:showid/episodes", s.UpdateEpisode)

		api.GET("indexers/search", s.ShowSearch)
	}

	r.GET("/statusz", s.Statusz)

	s.Handler = r
}

func (s *Server) Statusz(c *gin.Context) {
	marsh, err := json.MarshalIndent(s.config, "", "  ")
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
