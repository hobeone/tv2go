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
	ExceptionProviders map[string]nameexception.Provider
	Storage            *storage.Broker
	shutdownChan       chan (int)
}

//NewDaemon creates a new Daemon using the given config.
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
		"nzbsOrg":      providers.NewNzbsOrg(nzborgKey),
		"nyaaTorrents": providers.NewNyaaTorrents(),
	}

	broker, err := storage.NewBroker(cfg.Storage.Directories...)
	if err != nil {
		panic(fmt.Sprintf("Error creating storage broker: %s", err))
	}
	d.Storage = broker

	d.ExceptionProviders = map[string]nameexception.Provider{
		"tvdb":        nameexception.NewMidgetSpyTvdb(d.DBH),
		"thexem_tvdb": nameexception.NewXEM(d.DBH, "tvdb"),
		"thexem_rage": nameexception.NewXEM(d.DBH, "rage"),
	}

	return d
}

// Run will start the daemon which will run forever.
func (d *Daemon) Run() {
	exitChan := make(chan int)
	for _, ep := range d.ExceptionProviders {
		poller := nameexception.NewProviderPoller(ep, time.Hour*24, d.DBH)
		go poller.Poll(exitChan)
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

// PollProviders sets up goroutines that poll the configured providers on a set
// interval.  It then listens for new results and sends them for processing.
func (d *Daemon) PollProviders() {
	respChan := make(chan (providers.ProviderResult))
	for _, p := range d.Providers {
		go providers.NewProviderPoller(p, time.Minute*15, d.DBH, respChan).Poll()
	}
	for {
		select {
		case resp := <-respChan:
			err := d.ProcessProviderResult(resp)
			if err != nil {
				glog.Error(err.Error())
			}
		}
	}
}

// ProcessProviderResult takes a ProviderResult parses the name and sees if
// there it matches a show we are interested in.  If so it will try to download
// the url for that episode and send it to the right handler for that file
// type.
func (d *Daemon) ProcessProviderResult(r providers.ProviderResult) error {
	var np *naming.NameParser
	if r.Anime {
		np = naming.NewNameParser(naming.AnimeRegex)
	} else {
		np = naming.NewNameParser(naming.StandardRegexes)
	}
	pr := np.Parse(r.Name)

	//TODO: make this work with more kinds of episodes:
	if len(pr.EpisodeNumbers) == 0 && len(pr.AbsoluteEpisodeNumbers) == 0 {
		return fmt.Errorf("Provider result %s had no episodes, skipping", pr.OriginalName)
	}

	dbshow, season, err := d.DBH.GetShowByAllNames(pr.SeriesName)

	if err != nil {
		return fmt.Errorf("Couldn't match '%s' to any known show name: %s", pr.SeriesName, err)
	}

	if season > -1 {
		pr.SeasonNumber = season
	}

	ep, err := d.DBH.GetEpisodeByShowSeasonAndNumber(dbshow.ID, pr.SeasonNumber, pr.FirstEpisode())
	if err != nil {
		return fmt.Errorf("Can't find episode in DB Show: %s S%dE%d: %s", dbshow.Name, pr.SeasonNumber, pr.FirstEpisode(), err)
	}

	glog.Infof("Found matching episode in db: %s S%dE%d: %s", dbshow.Name, pr.SeasonNumber, pr.FirstEpisode(), ep.Name)
	// Don't like this, super fragile
	p, ok := d.Providers[r.ProviderName]
	if !ok {
		return fmt.Errorf("This daemon doesn't know about provider %s, skipping", r.ProviderName)
	}

	if ep.Status == types.WANTED {
		destPath := ""
		switch p.Type() {
		case providers.NZB:
			destPath = d.Config.Storage.NZBBlackhole
		}
		filename, filecont, err := p.GetURL(r.URL)
		if err != nil {
			return fmt.Errorf("Couldn't download %s: %s", r.URL, err)
		}
		fname, err := d.Storage.SaveToFile(destPath, filename, filecont)
		if err != nil {
			return fmt.Errorf("Error saving file to %s: %s", fname, err)
		}
		ep.Status = types.SNATCHED
		err = d.DBH.SaveEpisode(ep)
		if err != nil {
			glog.Errorf("Error saving episode %s: %s", ep.Name, err)
		}
	}
	return nil
}
