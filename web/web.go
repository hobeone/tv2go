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
	"github.com/hobeone/tv2go/indexers/tvrage"
)

type JSONShowCache struct {
	Banner int
	Poster int
}
type JSONShow struct {
	ID            int64         `json:"id"`
	AirByDate     int           `json:"air_by_date"`
	Cache         JSONShowCache `json:"cache"`
	Anime         int           `json:"anime"`
	IndexerID     int64         `json:"indexerid"`
	Language      string        `json:"language"`
	Network       string        `json:"network"`
	NextEpAirdate string        `json:"next_ep_airdate"`
	Paused        int           `json:"paused"`
	Quality       string        `json:"quality"`
	Name          string        `json:"name"`
	Sports        int           `json:"sports"`
	Status        string        `json:"status"`
	Subtitles     int           `json:"subtitles"`
	TVDBID        int64         `json:"tvdbid"`
	TVRageID      int64         `json:"tvrage_id"`
	TVRageName    string        `json:"tvrage_name"`
	SeasonList    []int64       `json:"season_list"`
}

func Btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

func Shows(c *gin.Context) {
	h := c.MustGet("dbh").(*db.Handle)
	shows, err := h.GetAllShows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, "")
	}
	jsonshows := make([]JSONShow, len(shows))
	for i, s := range shows {
		jsonshows[i] = JSONShow{
			ID:        s.ID,
			AirByDate: Btoi(s.AirByDate),
			//Cache
			Anime:     Btoi(s.Anime),
			IndexerID: s.IndexerID,
			Language:  s.Language,
			Network:   s.Network,
			//NextEpAirdate: s.NextEpAirdate(),
			Paused:    Btoi(s.Paused),
			Quality:   strconv.FormatInt(s.Quality, 10),
			Name:      s.Name,
			Sports:    Btoi(s.Sports),
			Status:    s.Status,
			Subtitles: Btoi(s.Subtitles),
			TVDBID:    s.IndexerID,
			//TVdbid, rageid + name
		}
	}
	c.JSON(200, jsonshows)
}

type GenericResult struct {
	Message string `json:"message"`
	Result  string `json:"result"`
}

func genError(c *gin.Context, status int, msg string) {
	c.JSON(status, GenericResult{
		Message: msg,
		Result:  "failure",
	})
}

func Show(c *gin.Context) {
	h := c.MustGet("dbh").(*db.Handle)
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
	response := JSONShow{
		ID:        s.ID,
		AirByDate: Btoi(s.AirByDate),
		//Cache
		Anime:     Btoi(s.Anime),
		IndexerID: s.IndexerID,
		Language:  s.Language,
		Network:   s.Network,
		//NextEpAirdate: s.NextEpAirdate(),
		Paused:     Btoi(s.Paused),
		Quality:    strconv.FormatInt(s.Quality, 10),
		Name:       s.Name,
		Sports:     Btoi(s.Sports),
		Status:     s.Status,
		Subtitles:  Btoi(s.Subtitles),
		TVDBID:     s.IndexerID,
		SeasonList: h.ShowSeasons(s),
		//TVdbid, rageid + name
	}

	c.JSON(http.StatusOK, response)
}

func ShowEpisodes(c *gin.Context) {
	h := c.MustGet("dbh").(*db.Handle)
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

func episodesToResponse(eps []db.Episode) []EpisodeResponse {
	resp := make([]EpisodeResponse, len(eps))
	for i, ep := range eps {
		resp[i] = EpisodeResponse{
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

type EpisodeResponse struct {
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

func Episode(c *gin.Context) {
	h := c.MustGet("dbh").(*db.Handle)
	episodeid, err := strconv.ParseInt(c.Params.ByName("episodeid"), 10, 64)

	if err != nil {
		c.JSON(http.StatusNotFound, GenericResult{
			Message: fmt.Sprintf("Invalid episodeid: %v", c.Params.ByName("episodeid")),
			Result:  "failure",
		})
		return
	}

	ep, err := h.GetEpisodeByID(episodeid)
	if err != nil {
		c.JSON(http.StatusNotFound, GenericResult{
			Message: err.Error(),
			Result:  "failure",
		})
		return
	}
	resp := EpisodeResponse{
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
func UpdateEpisode(c *gin.Context) {
	var epUpdate EpisodeResponse
	if !c.Bind(&epUpdate) {
		c.JSON(http.StatusBadRequest, GenericResult{
			Message: c.Errors.String(),
			Result:  "failure",
		})
		return
	}
	episode := epUpdate
	c.JSON(200, episode)
}

// cmd=show.addnew&tvdbid=73871
func AddShow(c *gin.Context) {
	tvdbid := c.Request.Form.Get("tvdbid")
	tvrageid := c.Request.Form.Get("tvrageid")

	h := c.MustGet("dbh").(*db.Handle)
	glog.Info("Adding show with args: %s", c.Request.URL)
	if tvdbid != "" {
		glog.Infof("Got tvdbid to add: %s", tvdbid)
		tvdbid, err := strconv.ParseInt(tvdbid, 10, 64)
		if err != nil {
			c.JSON(500, "Bad tvrageid")
			return
		}
		s, eps, err := tvdb.GetShowById(tvdbid)
		if err != nil {
			c.JSON(500, GenericResult{
				Message: err.Error(),
				Result:  "failure",
			})
			return
		}
		dbshow := tvdb.TVDBToShow(s)
		for _, ep := range eps {
			dbshow.Episodes = append(dbshow.Episodes, tvdb.ConvertTvdbEpisodeToDbEpisode(ep))
		}
		err = h.AddShow(&dbshow)
		if err != nil {
			c.JSON(500, err.Error())
		}
		c.JSON(200, gin.H{
			"data": gin.H{
				"name": dbshow.Name,
			},
			"message": fmt.Sprintf("%s has been queued to be added", dbshow.Name),
			"result":  "success",
		})
	}
	if tvrageid != "" {
		glog.Infof("Got tvrage to add: %s", tvdbid)
		rageid, err := strconv.ParseInt(tvrageid, 10, 64)
		if err != nil {
			c.JSON(500, "Bad tvrageid")
			return
		}
		showinfo, err := tvrage.GetShowInfo(rageid)
		if err != nil {
			c.JSON(500, "Couldn't get info from tvrage")
			return
		}
		rageshow := tvrage.TVRageToShow(showinfo)
		h := c.MustGet("dbh").(*db.Handle)
		err = h.AddShow(&rageshow)
		if err != nil {
			c.JSON(500, err.Error())
		}
		c.JSON(200, gin.H{
			"data": gin.H{
				"name": rageshow.Name,
			},
			"message": fmt.Sprintf("%s has been queued to be added", rageshow.Name),
			"result":  "success",
		})
	}
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

func createServer(dbh *db.Handle) *gin.Engine {
	r := gin.New()
	r.Use(Logger())
	r.Use(CORSMiddleware())

	r.Use(DBHandler(dbh))

	api := r.Group("/api/:apistring")
	{
		api.OPTIONS("/*cors", func(c *gin.Context) {})
		api.GET("shows", Shows)
		api.GET("shows/:showid", Show)
		api.GET("shows/:showid/episodes", ShowEpisodes)
		api.GET("shows/:showid/episodes/:episodeid", Episode)

		api.PUT("shows/:showid/episodes", UpdateEpisode)
	}

	return r
}

func StartServer(cfg *config.Config, dbh *db.Handle) {
	r := createServer(dbh)
	glog.Fatal(http.ListenAndServe(cfg.WebServer.ListenAddress, r))
}
