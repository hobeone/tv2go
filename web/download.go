package web

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hobeone/tv2go/providers"
	"github.com/hobeone/tv2go/types"
)

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
