package main

import (
	"flag"
	"time"

	"github.com/golang/glog"
	"github.com/hobeone/tv2go/config"
	"github.com/hobeone/tv2go/db"
	"github.com/hobeone/tv2go/indexers/tvdb"
	"github.com/hobeone/tv2go/web"
)

// ShowUpdater is ment to be run as a background goroutine for that looks for
// Show which need updates from their indexer.
func ShowUpdater(dbh *db.Handle) {
	h := db.NewDBHandle("test.db", true, true)
	oldage := time.Duration(86400) * time.Second
	for {
		shows, _ := h.GetAllShows()
		glog.Infof("Got %d shows, checking if they need updates", len(shows))
		for _, s := range shows {
			if time.Now().Sub(s.LastIndexerUpdate) > oldage {
				glog.Infof("%s hasn't been updated in more than %v", s.Name, oldage)
				dbeps, err := h.GetShowEpisodes(&s)
				if err != nil {
					glog.Errorf("Error getting show episodes from db: %s", err)
					continue
				}
				dbshow, dbeps, err := tvdb.UpdateDBShow(s, dbeps)
				if err != nil {
					glog.Errorf("Error updating show: %s", err.Error())
					continue
				}
				glog.Infof("Saving %d episodes", len(dbeps))
				h.DB().Save(&dbshow)
			}
		}
		glog.Info("Updated shows, sleeping.")
		time.Sleep(time.Duration(60) * time.Second)
	}
}

// Daemon contains everything needed to run a Tv2Go daemon.
type Daemon struct {
	Config *config.Config
	DBH    *db.Handle
}

// NewDaemon creates a new Daemon with the given config.
func NewDaemon(cfg *config.Config) *Daemon {
	var dbh *db.Handle
	if cfg.DB.Type == "memory" {
		dbh = db.NewMemoryDBHandle(cfg.DB.Verbose, cfg.DB.UpdateDb)
	} else {
		dbh = db.NewDBHandle(cfg.DB.Path, cfg.DB.Verbose, cfg.DB.UpdateDb)
	}
	return &Daemon{
		Config: cfg,
		DBH:    dbh,
	}
}

func runDaemon(cfg *config.Config) {
	d := NewDaemon(cfg)
	go ShowUpdater(d.DBH)

	web.StartServer(cfg, d.DBH)
}

const defaultConfig = "~/.config/tv2go/config.json"

func loadConfig(configFile string) *config.Config {
	if len(configFile) == 0 {
		glog.Infof("No --config_file given.  Using default: %s\n",
			defaultConfig)
		configFile = defaultConfig
	}

	glog.Infof("Got config file: %s\n", configFile)
	config := config.NewConfig()
	err := config.ReadConfig(configFile)
	if err != nil {
		glog.Fatal(err)
	}
	return config
}

func main() {
	defer glog.Flush()
	flag.Set("logtostderr", "true")

	cfgfile := flag.String("config_file", defaultConfig, "Config file to use")

	flag.Parse()

	cfg := loadConfig(*cfgfile)

	runDaemon(cfg)
}
