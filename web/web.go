package web

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/hobeone/tvrage"
)

func Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"data":    gin.H{"pid": 100},
		"message": "Pong",
		"result":  "success",
	})
}

func Shows(c *gin.Context) {
	c.JSON(200, SHOWSRESP)
}

func Show(c *gin.Context) {
	c.JSON(200, SHOWRESP)
}

func ShowSeasons(c *gin.Context) {
	// &indexerid=152831&season=6
	c.JSON(200, SHOWSEASONSRESP)
}

func History(c *gin.Context) {}

func Logs(c *gin.Context)   {}
func Future(c *gin.Context) {}

// cmd=show.addnew&tvdbid=73871
func AddShow(c *gin.Context) {
	tvdbid := c.Params.ByName("tvdbid")
	tvrageid := c.Params.ByName("tvrageid")

	if tvdbid != "" {
		return
	} else if tvrageid != "" {
		rageid, err := strconv.ParseInt(tvrageid, 10, 64)
		if err != nil {
			c.JSON(500, "Bad tvrageid")
			return
		}
		tvrage.Get(rageid)
	}
}

func createServer() *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	r.GET("/api/:apistring/*cmdname", func(c *gin.Context) {
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
		glog.Error(cmdName)
		c.String(500, "Unknown command: '%v'")
	})

	return r
}

func StartServer() {
	r := createServer()
	glog.Fatal(r.Run(":9000"))
}
