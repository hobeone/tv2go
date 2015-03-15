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
	"github.com/hobeone/tv2go/indexers/tvrage"
	"github.com/hobeone/tv2go/nameexception"
	"github.com/hobeone/tv2go/naming"
	"github.com/hobeone/tv2go/providers"
	"github.com/hobeone/tv2go/storage"
	"github.com/hobeone/tv2go/types"
	"github.com/hobeone/tv2go/web"
)

// ShowUpdater is ment to be run as a background goroutine for that looks for
// Show which need updates from their indexer.
func (d *Daemon) ShowUpdater() {
	oldage := time.Duration(86400) * time.Second
	for {
		shows, err := d.DBH.GetAllShows()
		if err != nil {
			glog.Errorf("Error getting shows: %s", err)
			break
		}
		glog.Infof("Got %d shows, checking if they need updates", len(shows))
		for _, s := range shows {
			if time.Now().Sub(s.LastIndexerUpdate) > oldage {
				glog.Infof("%s hasn't been updated in more than %v", s.Name, oldage)
				dbeps, err := d.DBH.GetShowEpisodes(&s)
				if err != nil {
					glog.Errorf("Error getting show episodes from db: %s", err)
					continue
				}
				if _, ok := d.Indexers[s.Indexer]; !ok {
					glog.Errorf("Unknown indexer %s for show %s", s.Indexer, s.Name)
					continue
				}
				err = d.Indexers[s.Indexer].UpdateShow(&s)
				if err != nil {
					glog.Errorf("Error updating show: %s", err.Error())
					continue
				}
				glog.Infof("Saving %d episodes", len(dbeps))
				d.DBH.SaveShow(&s)
			}
		}
		glog.Info("Updated shows, sleeping.")
		time.Sleep(time.Duration(900) * time.Second)
	}
}

func (d *Daemon) ProviderPoller(providerReg *providers.ProviderRegistry, broker *storage.Broker) {
	for {
		np := naming.NewNameParser("", naming.StandardRegexes)
		for name, p := range *providerReg {
			glog.Infof("Getting new items from %s", name)
			res, err := p.GetNewItems()
			if err != nil {
				glog.Errorf("Error getting new items from %s: %s", name, err)
				continue
			}
			for _, r := range res {
				pr := np.Parse(r.Name)
				dbshow, err := d.DBH.GetShowByName(pr.SeriesName)
				if err != nil {
					glog.Warningf("Couldn't find show '%s' in database, skipping.", pr.SeriesName)
					continue
				}
				ep, err := d.DBH.GetEpisodeByShowSeasonAndNumber(dbshow.ID, pr.SeasonNumber, pr.EpisodeNumbers[0])
				if err != nil {
					glog.Errorf("Don't know about Season '%d', Episode '%d', for Show %s", pr.SeasonNumber, pr.EpisodeNumbers[0], dbshow.Name)
					continue
				}
				if ep.Status == types.WANTED {
					p.GetURL(r.URL)
					destPath := ""
					switch p.Type() {
					case providers.NZB:
						destPath = d.Config.Storage.NZBBlackhole
					}
					filename, filecont, err := p.GetURL(r.URL)
					if err != nil {
						glog.Errorf("Couldn't download episode %s: %s", r.URL, err)
						continue
					}
					fname, err := broker.SaveToFile(destPath, filename, filecont)
					if err != nil {
						glog.Errorf("Error saving to %s: %s", fname, err)
					}
					ep.Status = types.SNATCHED
					err = d.DBH.SaveEpisode(ep)
					if err != nil {
						glog.Errorf("Error saving episode %s: %s", ep.Name, err)
					}
				}
			}
		}
		glog.Info("Updated Providers, sleeping.")
		time.Sleep(time.Duration(900) * time.Second)
	}
}

// Daemon contains everything needed to run a Tv2Go daemon.
type Daemon struct {
	Config   *config.Config
	DBH      *db.Handle
	Indexers indexers.IndexerRegistry
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

	d.Indexers = indexers.IndexerRegistry{
		"tvdb":   tvdb.NewTvdbIndexer("90D7DF3AE9E4841E"),
		"tvrage": tvrage.NewTVRageIndexer(),
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
		"nzbsOrg": providers.NewNzbsOrg(nzborgKey),
		//		"nyaaTorrents": providers.NewNyaaTorrents(),
	}

	broker, err := storage.NewBroker(cfg.Storage.Directories...)
	if err != nil {
		panic(fmt.Sprintf("Error creating storage broker: %s", err))
	}

	exceptionProviders := map[string]*nameexception.Provider{
		"tvdb": nameexception.NewProvider(
			"tvdb",
			"tvdb",
			"https://midgetspy.github.io/sb_tvdb_scene_exceptions/exceptions.txt",
			time.Hour*24,
			d.DBH,
		),
	}
	exitChan := make(chan int)
	for _, ep := range exceptionProviders {
		go ep.Poll(exitChan)
	}

	go d.ShowUpdater()
	go d.ProviderPoller(&provReg, broker)
	webserver := web.NewServer(cfg, d.DBH, broker, provReg, web.SetIndexers(d.Indexers))

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
	flag.Set("alsologtostderr", "true")

	cfgfile := flag.String("config_file", defaultConfig, "Config file to use")

	flag.Parse()

	cfg := loadConfig(*cfgfile)

	runDaemon(cfg)
}
