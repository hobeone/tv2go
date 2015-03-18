package daemon

import (
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
	go d.PollProviders()
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
			glog.Errorf("Error getting shows: %s", err)
			break
		}
		glog.Infof("Got %d shows from db", len(shows))
		for _, s := range shows {
			if time.Now().Sub(s.LastIndexerUpdate) > oldage {
				glog.Infof("%s hasn't been updated in more than %v", s.Name, oldage)
				dbeps, err := d.DBH.GetShowEpisodes(&s)
				if err != nil {
					glog.Errorf("Error getting show episodes from db: %s", err)
					continue
				}
				if _, ok := d.Indexers[s.Indexer]; !ok {
					glog.Errorf("Unknown indexer '%s' for show %s", s.Indexer, s.Name)
					continue
				}
				episodes, err := d.DBH.GetShowEpisodes(&s)
				if err != nil {
					glog.Errorf("error getting episodes for show %s: %s", s.Name, err)
				}
				err = d.Indexers[s.Indexer].UpdateShow(&s, episodes)
				if err != nil {
					glog.Errorf("Error updating show %s: %s", s.Name, err.Error())
					continue
				}
				glog.Infof("Saving %d episodes for %s", len(dbeps), s.Name)
				err = d.DBH.SaveShow(&s)
				if err != nil {
					glog.Errorf("error saving show %s to db: %s", s.Name, err)
				}
			}
		}
		toSleep := time.Duration(900) * time.Second
		glog.Infof("Updated shows, sleeping %s", toSleep.String())
		time.Sleep(toSleep)
	}
}

func (d *Daemon) PollProviders() {
	respChan := make(chan (providers.ProviderResult))
	for _, p := range d.Providers {
		go providers.NewProviderPoller(p, time.Minute*15, d.DBH, respChan).Poll()
	}
	for {
		select {
		case resp := <-respChan:
			d.ProcessProviderResult(resp)
		}
	}
}

func (d *Daemon) matchShowName(name string) (*db.Show, error) {
	glog.Infof("Trying to match provider result %s", name)

	glog.Infof("Trying to find an exact match in the database for %s", name)
	dbshow, err := d.DBH.GetShowByName(name)
	if err == nil {
		glog.Infof("Matched name %s to show %s", name, dbshow.Name)
		return dbshow, nil
	}
	glog.Infof("Couldn't find show with name %s in database.", name)

	sceneName := naming.FullSanitizeSceneName(name)
	glog.Infof("Converting name '%s' to scene name '%s'", name, sceneName)

	dbshow, err = d.DBH.GetShowFromNameException(sceneName)
	if err == nil {
		glog.Infof("Matched provider result %s to show %s", sceneName, dbshow.Name)
		return dbshow, nil
	}
	glog.Infof("Couldn't find a match scene name %s", sceneName)

	return nil, fmt.Errorf("Couldn't find a match for show %s", name)
}

func (d *Daemon) ProcessProviderResult(r providers.ProviderResult) {
	np := naming.NewNameParser("", naming.StandardRegexes)
	pr := np.Parse(r.Name)

	//TODO: make this work with more kinds of episodes:
	if len(pr.EpisodeNumbers) == 0 {
		glog.Infof("Provider result %s had no episodes, skipping", pr.OriginalName)
		return
	}

	dbshow, err := d.matchShowName(pr.SeriesName)

	if err != nil {
		return
	}

	ep, err := d.DBH.GetEpisodeByShowSeasonAndNumber(dbshow.ID, pr.SeasonNumber, pr.EpisodeNumbers[0])
	if err != nil {
		glog.Errorf("Can't find episode in DB Show: %s S%dE%d", dbshow.Name, pr.SeasonNumber, pr.EpisodeNumbers[0])
		return
	}
	// Don't like this, super fragile
	p := d.Providers[r.ProviderName]
	if ep.Status == types.WANTED {
		destPath := ""
		switch p.Type() {
		case providers.NZB:
			destPath = d.Config.Storage.NZBBlackhole
		}
		filename, filecont, err := p.GetURL(r.URL)
		if err != nil {
			glog.Errorf("Couldn't download %s: %s", r.URL, err)
			return
		}
		fname, err := d.Storage.SaveToFile(destPath, filename, filecont)
		if err != nil {
			glog.Errorf("Error saving file to %s: %s", fname, err)
			return
		}
		ep.Status = types.SNATCHED
		err = d.DBH.SaveEpisode(ep)
		if err != nil {
			glog.Errorf("Error saving episode %s: %s", ep.Name, err)
		}
	}
	return
}
