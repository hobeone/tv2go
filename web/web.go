package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/hobeone/tv2go/config"
	"github.com/hobeone/tv2go/db"
	"github.com/hobeone/tv2go/indexers"
	"github.com/hobeone/tv2go/naming"
	"github.com/hobeone/tv2go/providers"
	"github.com/hobeone/tv2go/storage"
	"github.com/hobeone/tv2go/types"
)

type genericResult struct {
	Message string `json:"message"`
	Result  string `json:"result"`
}

func genError(c *gin.Context, status int, msg string) {
	logger.Info("Error serving request", "url", c.Request.URL.String(), "err", msg)
	c.JSON(status, genericResult{
		Message: msg,
		Result:  "failure",
	})
}

func showToLocation(path, name string) string {
	newpath := filepath.Join(path, naming.CleanSeriesName(name))
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

// StartServing does just what it says.
func (server *Server) StartServing() {
	logger.Fatal("Error starting server", "err",
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
	gin.SetMode(gin.TestMode)
	configGinEngine(t)
	for _, option := range options {
		option(t)
	}
	return t
}
