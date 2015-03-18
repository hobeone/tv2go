package web

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hobeone/tv2go/db"
	"github.com/hobeone/tv2go/types"
)

func episodeToResponse(ep *db.Episode) episodeResponse {
	return episodeResponse{
		ID:              ep.ID,
		ShowID:          ep.ShowId,
		AirDate:         ep.AirDateString(),
		Description:     ep.Description,
		FileSize:        ep.FileSize,
		Location:        ep.Location,
		Name:            ep.Name,
		Quality:         ep.Quality.String(),
		ReleaseName:     ep.ReleaseName,
		Status:          ep.Status.String(),
		Season:          ep.Season,
		Episode:         ep.Episode,
		AbsoluteEpisode: ep.AbsoluteNumber,
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
	ID              int64  `json:"id" form:"id" binding:"required"`
	ShowID          int64  `json:"showid" form:"showid" binding:"required"`
	Name            string `json:"name" form:"name" binding:"required"`
	Season          int64  `json:"season" form:"season"`
	Episode         int64  `json:"episode" form:"episode"`
	AbsoluteEpisode int64  `json:"absolute_episode" form:"absolute_episode"`
	AirDate         string `json:"airdate" form:"airdate"`
	Description     string `json:"description" form:"description"`
	FileSize        int64  `json:"file_size" form:"file_size"`
	FileSizeHuman   string `json:"file_size_human" form:"file_size_human"`
	Location        string `json:"location" form:"location"`
	Quality         string `json:"quality" form:"quality"`
	ReleaseName     string `json:"release_name" form:"release_name"`
	Status          string `json:"status" form:"status"`
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
