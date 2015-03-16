package daemon

import (
	"fmt"
	"time"

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
	"github.com/mgutz/logxi/v1"
)

// Daemon contains everything needed to run a Tv2Go daemon.
type Daemon struct {
	Config             *config.Config
	DBH                *db.Handle
	Indexers           indexers.IndexerRegistry
	Providers          providers.ProviderRegistry
	ExceptionProviders map[string]*nameexception.Provider
	Storage            *storage.Broker
	shutdownChan       chan (int)
}

func NewDaemon(cfg *config.Config) *Daemon {
	var dbh *db.Handle
	if cfg.DB.Type == "memory" {
		dbh = db.NewMemoryDBHandle(cfg.DB.Verbose, cfg.DB.UpdateDb)
	} else {
		dbh = db.NewDBHandle(cfg.DB.Path, cfg.DB.Verbose, cfg.DB.UpdateDb)
	}
	d := &Daemon{
		Config: cfg,
		DBH:    dbh,
	}

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
	d.Providers = providers.ProviderRegistry{
		"nzbsOrg": providers.NewNzbsOrg(nzborgKey),
		//		"nyaaTorrents": providers.NewNyaaTorrents(),
	}

	broker, err := storage.NewBroker(cfg.Storage.Directories...)
	if err != nil {
		panic(fmt.Sprintf("Error creating storage broker: %s", err))
	}
	d.Storage = broker

	d.ExceptionProviders = map[string]*nameexception.Provider{
		"tvdb": nameexception.NewProvider(
			"tvdb",
			"tvdb",
			"https://midgetspy.github.io/sb_tvdb_scene_exceptions/exceptions.txt",
			time.Hour*24,
			d.DBH,
		),
	}

	return d
}

func (d *Daemon) Run() {
	exitChan := make(chan int)
	for _, ep := range d.ExceptionProviders {
		go ep.Poll(exitChan)
	}

	go d.ShowUpdater()
	go d.ProviderPoller()
	webserver := web.NewServer(d.Config, d.DBH, d.Storage, d.Providers, web.SetIndexers(d.Indexers))

	webserver.StartServing()
}

// ShowUpdater is ment to be run as a background goroutine for that looks for
// Show which need updates from their indexer.
func (d *Daemon) ShowUpdater() {
	oldage := time.Duration(86400) * time.Second
	for {
		shows, err := d.DBH.GetAllShows()
		if err != nil {
			log.Error("Error getting shows.", "error", err)
			break
		}
		log.Debug("Got shows from db", "shows", len(shows))
		for _, s := range shows {
			if time.Now().Sub(s.LastIndexerUpdate) > oldage {
				log.Info("%s hasn't been updated in more than %v", s.Name, oldage)
				dbeps, err := d.DBH.GetShowEpisodes(&s)
				if err != nil {
					log.Error("Error getting show episodes from db", "err", err)
					continue
				}
				if _, ok := d.Indexers[s.Indexer]; !ok {
					log.Error("Unknown indexer for show", "indexer", s.Indexer, "show", s.Name)
					continue
				}
				err = d.Indexers[s.Indexer].UpdateShow(&s)
				if err != nil {
					log.Error("Error updating show", "show", s.Name, "err", err.Error())
					continue
				}
				log.Info("Saving %d episodes", len(dbeps))
				d.DBH.SaveShow(&s)
			}
		}
		toSleep := time.Duration(900) * time.Second
		log.Info("Updated shows, sleeping.", "time", toSleep.String())
		time.Sleep(toSleep)
	}
}

func (d *Daemon) ProviderPoller() {
	interval := time.Second * 900
	for {
		np := naming.NewNameParser("", naming.StandardRegexes)
		//Hack to stop hitting providers to much
		//TODO: make this per provider etc
		lastPoll := d.DBH.GetLastPollTime("providers")
		if time.Since(lastPoll) < interval {
			toSleep := interval - time.Since(lastPoll)
			log.Info("Povider: sleeping until next update", "interval", interval.String(), "sleeptime", toSleep.String())
			time.Sleep(toSleep)
		} else {
			log.Info("lastpoll was less than the min interval", "lastpoll", lastPoll.String())
		}
		for name, p := range d.Providers {
			log.Info("Getting new items from provider", "provider", name)
			res, err := p.GetNewItems()
			if err != nil {
				log.Error("Error getting new items from provider", "provider", name, "err", err)
				continue
			}
			for _, r := range res {
				pr := np.Parse(r.Name)
				dbshow, err := d.DBH.GetShowByName(pr.SeriesName)
				if err != nil {
					log.Warn("Couldn't find show in database, skipping.", "show", pr.SeriesName)
					continue
				}
				ep, err := d.DBH.GetEpisodeByShowSeasonAndNumber(dbshow.ID, pr.SeasonNumber, pr.EpisodeNumbers[0])
				if err != nil {
					log.Error("Can't find episode in DB", "season", pr.SeasonNumber, "episode", pr.EpisodeNumbers[0], "show", dbshow.Name)
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
						log.Error("Couldn't download episode", "url", r.URL, "err", err)
						continue
					}
					fname, err := d.Storage.SaveToFile(destPath, filename, filecont)
					if err != nil {
						log.Error("Error saving file", "path", fname, "err", err)
					}
					ep.Status = types.SNATCHED
					err = d.DBH.SaveEpisode(ep)
					if err != nil {
						log.Error("Error saving episode", "name", ep.Name, "err", err)
					}
				}
			}
		}
		d.DBH.SetLastPollTime("providers")
	}
}
