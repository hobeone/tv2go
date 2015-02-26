package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/hobeone/tv2go/config"
	"github.com/hobeone/tv2go/db"
	"github.com/hobeone/tv2go/indexers"
	"github.com/hobeone/tv2go/indexers/tvdb"
	"github.com/hobeone/tv2go/providers"
	"github.com/hobeone/tv2go/storage"
	"github.com/hobeone/tv2go/web"
)

// ShowUpdater is ment to be run as a background goroutine for that looks for
// Show which need updates from their indexer.
func ShowUpdater(h *db.Handle) {
	oldage := time.Duration(86400) * time.Second
	for {
		shows, err := h.GetAllShows()
		if err != nil {
			glog.Errorf("Error getting shows: %s", err)
			break
		}
		glog.Infof("Got %d shows, checking if they need updates", len(shows))
		for _, s := range shows {
			if time.Now().Sub(s.LastIndexerUpdate) > oldage {
				glog.Infof("%s hasn't been updated in more than %v", s.Name, oldage)
				dbeps, err := h.GetShowEpisodes(&s)
				if err != nil {
					glog.Errorf("Error getting show episodes from db: %s", err)
					continue
				}
				err = tvdb.NewTvdbIndexer("").UpdateShow(&s)
				if err != nil {
					glog.Errorf("Error updating show: %s", err.Error())
					continue
				}
				glog.Infof("Saving %d episodes", len(dbeps))
				h.SaveShow(&s)
			}
		}
		glog.Info("Updated shows, sleeping.")
		time.Sleep(time.Duration(5) * time.Second)
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
	//go ShowUpdater(d.DBH)

	idxReg := indexers.IndexerRegistry{
		"tvdb": tvdb.NewTvdbIndexer("90D7DF3AE9E4841E"),
	}

	//Ghetto until real provider setup done
	nzborgKey := ""
	for _, p := range cfg.Providers {
		if p.Name == "nzbsOrg" {
			nzborgKey = p.API
		}
	}
	if nzborgKey == "" {
		panic("No Nzbs.org API key set in config")
	}
	provReg := providers.ProviderRegistry{
		// get key from cfg
		"nzbsOrg": providers.NewNzbsOrg(nzborgKey),
	}

	broker, err := storage.NewBroker(cfg.Storage.Directories...)
	if err != nil {
		panic(fmt.Sprintf("Error creating storage broker: %s", err))
	}

	webserver := web.NewServer(cfg, d.DBH, broker, provReg, web.SetIndexers(idxReg))

	webserver.StartServing()
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
