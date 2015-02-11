package web

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/golang/glog"
	"github.com/hobeone/tv2go/config"
	"github.com/hobeone/tv2go/db"
	"github.com/hobeone/tv2go/indexers/tvdb"
	"github.com/hobeone/tv2go/indexers/tvrage"
)

func Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"data":    gin.H{"pid": os.Getpid()},
		"message": "Pong",
		"result":  "success",
	})
}

type JSONShowCache struct {
	Banner int
	Poster int
}
type JSONShow struct {
	AirByDate     int           `json:"air_by_date"`
	Cache         JSONShowCache `json:"cache"`
	Anime         int           `json:"anime"`
	IndexerID     int64         `json:"indexerid"`
	Language      string        `json:"language"`
	Network       string        `json:"network"`
	NextEpAirdate string        `json:"next_ep_airdate"`
	Paused        int           `json:"paused"`
	Quality       string        `json:"quality"`
	ShowName      string        `json:"show_name"`
	Sports        int           `json:"sports"`
	Status        string        `json:"status"`
	Subtitles     int           `json:"subtitles"`
	TVDBID        int64         `json:"tvdbid"`
	TVRageID      int64         `json:"tvrage_id"`
	TVRageName    string        `json:"tvrage_name"`
	SeasonList    []int64       `json:"season_list"`
}

//JSONShow example.
type JSONShowMap map[string]JSONShow

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
	jsonshows := make(JSONShowMap, len(shows))
	for _, s := range shows {
		jsonshows[s.Name] = JSONShow{
			AirByDate: Btoi(s.AirByDate),
			//Cache
			Anime:     Btoi(s.Anime),
			IndexerID: s.IndexerID,
			Language:  s.Language,
			Network:   s.Network,
			//NextEpAirdate: s.NextEpAirdate(),
			Paused:    Btoi(s.Paused),
			Quality:   strconv.FormatInt(s.Quality, 10),
			ShowName:  s.Name,
			Sports:    Btoi(s.Sports),
			Status:    s.Status,
			Subtitles: Btoi(s.Subtitles),
			TVDBID:    s.IndexerID,
			//TVdbid, rageid + name
		}
	}
	c.JSON(200, gin.H{
		"data": jsonshows,
	})
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
	tvdbidstr := c.Request.Form.Get("tvdbid")
	tvdbid, err := strconv.ParseInt(tvdbidstr, 10, 64)
	if err != nil {
		genError(c, http.StatusInternalServerError, "invalid show id")
		return
	}
	s, err := h.GetShowById(tvdbid)
	if err != nil {
		genError(c, http.StatusNotFound, "Show not found")
		return
	}
	seasonList := map[int64]bool{}
	for _, ep := range s.Episodes {
		seasonList[ep.Season] = true
	}
	keys := make([]int64, len(seasonList))

	i := 0
	for k := range seasonList {
		keys[i] = k
		i++
	}
	response := gin.H{
		"data": JSONShow{
			AirByDate: Btoi(s.AirByDate),
			//Cache
			Anime:     Btoi(s.Anime),
			IndexerID: s.IndexerID,
			Language:  s.Language,
			Network:   s.Network,
			//NextEpAirdate: s.NextEpAirdate(),
			Paused:     Btoi(s.Paused),
			Quality:    strconv.FormatInt(s.Quality, 10),
			ShowName:   s.Name,
			Sports:     Btoi(s.Sports),
			Status:     s.Status,
			Subtitles:  Btoi(s.Subtitles),
			TVDBID:     s.IndexerID,
			SeasonList: h.ShowSeasons(s),
			//TVdbid, rageid + name
		},
		"message": "",
		"result":  "success",
	}

	c.JSON(http.StatusOK, response)
}

type ShowSeasonResponse struct {
	AirDate string `json:"airdate"`
	Name    string `json:"name"`
	Quality string `json:"quality"`
	Status  string `json:"status"`
}

type ShowSeasonsForm struct {
	TVDBID int64 `form:"tvdbid" binding:"required"`
	Season int64 `form:"season"`
}

// ShowSeasons returns detailed episode information for the given show and season
func ShowSeasons(c *gin.Context) {
	h := c.MustGet("dbh").(*db.Handle)
	var formVals ShowSeasonsForm
	if !c.BindWith(&formVals, binding.Form) {
		return
	}

	eps, err := h.GetShowSeason(formVals.TVDBID, formVals.Season)
	if err != nil {
		c.JSON(http.StatusNotFound, GenericResult{
			Message: err.Error(),
			Result:  "failure",
		})
		return
	}
	result := make(map[string]ShowSeasonResponse, len(eps))
	for _, ep := range eps {
		result[strconv.FormatInt(ep.Episode, 10)] = ShowSeasonResponse{
			AirDate: ep.AirDateString(),
			Name:    ep.Name,
			Quality: ep.Quality,
			Status:  ep.Status,
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"data":    result,
		"message": "",
		"result":  "success",
	})
}

type EpisodeResponse struct {
	AirDate       string `json:"airdate"`
	Description   string `json:"description"`
	FileSize      int64  `json:"file_size"`
	FileSizeHuman string `json:"file_size_human"`
	Location      string `json:"location"`
	Name          string `json:"name"`
	Quality       string `json:"quality"`
	ReleaseName   string `json:"release_name"`
	Status        string `json:"status"`
}

// &tvdbid=101501&season=4&episode=8&full_path=1
type EpisodeRequestForm struct {
	TVDBID   int64 `form:"tvdbid" binding:"required"`
	Season   int64 `form:"season"`
	Episode  int64 `form:"episode"`
	FullPath int   `form:"full_path"`
}

func Episode(c *gin.Context) {
	h := c.MustGet("dbh").(*db.Handle)
	var formVals EpisodeRequestForm
	if !c.BindWith(&formVals, binding.Form) {
		c.JSON(http.StatusBadRequest, GenericResult{
			Message: "Bad Request",
			Result:  "failure",
		})
		return
	}

	ep, err := h.GetShowEpisodeBySeasonAndNumber(
		formVals.TVDBID, formVals.Season, formVals.Episode,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, GenericResult{
			Message: err.Error(),
			Result:  "failure",
		})
		return
	}
	resp := EpisodeResponse{
		AirDate:     ep.AirDateString(),
		Description: ep.Description,
		FileSize:    ep.FileSize,
		Location:    ep.Location,
		Name:        ep.Name,
		Quality:     ep.Quality,
		ReleaseName: ep.ReleaseName,
		Status:      ep.Status,
	}
	c.JSON(200, gin.H{
		"data":    resp,
		"message": "",
		"result":  "success",
	})
}

func History(c *gin.Context) {}

func Logs(c *gin.Context)   {}
func Future(c *gin.Context) {}

// cmd=show.addnew&tvdbid=73871
func AddShow(c *gin.Context) {
	tvdbid := c.Request.Form.Get("tvdbid")
	tvrageid := c.Request.Form.Get("tvrageid")

	glog.Info("Adding show with args: %s", c.Request.URL)
	if tvdbid != "" {
		glog.Infof("Got tvdbid to add: %s", tvdbid)
		tvdbid, err := strconv.ParseInt(tvdbid, 10, 64)
		if err != nil {
			c.JSON(500, "Bad tvrageid")
			return
		}
		s, err := tvdb.GetShowById(tvdbid)
		dbshow := tvdb.TVDBToShow(s)
		h := c.MustGet("dbh").(*db.Handle)
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
func createServer(dbh *db.Handle) *gin.Engine {
	r := gin.New()
	r.Use(Logger())

	r.Use(DBHandler(dbh))
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	r.GET("/api/:apistring/", func(c *gin.Context) {
		apistring := c.Params.ByName("apistring")
		glog.Info("Got API KEY:", apistring)
		glog.Info("Got URL: ", c.Request.URL)

		err := c.Request.ParseForm()
		if err != nil {
			glog.Error(err.Error())
			c.String(500, err.Error())
			return
		}

		cmdName := c.Request.Form.Get("cmd")

		switch cmdName {
		case "sb.ping":
			Ping(c)
			return
		case "shows":
			Shows(c)
			return
		case "show":
			Show(c)
			return
		case "show.getbanner":
			c.String(200, "TODO")
			return

		case "show.seasons":
			ShowSeasons(c)
			return

		case "show.addnew":
			AddShow(c)
			return

		case "episode":
			Episode(c)
			return

		case "future":
			Future(c)
			return

		case "history":
			History(c)
			return

		case "logs":
			Logs(c)
			return

		case "":
			c.String(200, "No command given")
			return
		}
		c.String(500, fmt.Sprintf("Unknown command: '%v'", cmdName))
	})

	return r
}

func StartServer(cfg *config.Config, dbh *db.Handle) {
	r := createServer(dbh)
	glog.Fatal(http.ListenAndServe(cfg.WebServer.ListenAddress, r))
}
