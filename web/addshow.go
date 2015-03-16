package web

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hobeone/tv2go/quality"
	"github.com/hobeone/tv2go/types"
)

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
	logger.Info("Got id to add", "id", indexerID)
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
		logger.Info("Location not set on show, using default", "show", dbshow.Name, "default", dbshow.Location)
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
