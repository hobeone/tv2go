package tvdb

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/hobeone/tv2go/db"
	tvd "github.com/hobeone/tvdb"
)

// Trying out funcitonal api config as described here:
// http://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis

// TvdbIndexer implements the Indexer interface
type TvdbIndexer struct {
	tvdbClient *tvd.Client
}

// NewTvdbIndexer returns a new configured indexer
func NewTvdbIndexer(apiKey string, options ...func(*TvdbIndexer)) *TvdbIndexer {
	t := &TvdbIndexer{
		tvdbClient: tvd.NewClient(apiKey),
	}
	for _, option := range options {
		option(t)
	}
	return t
}

func (t *TvdbIndexer) setClient(c *http.Client) error {
	t.tvdbClient.HTTPClient = c
	return nil
}

// SetClient set's the httpclient the Indexer will use.
//
// Example:
//  NewTvdbIndexer(apikey, SetClient(httpclient))
func SetClient(c *http.Client) func(*TvdbIndexer) {
	return func(t *TvdbIndexer) {
		t.setClient(c)
	}
}

// Name returns the string name of this indexer
func (t *TvdbIndexer) Name() string {
	return "tvdb"
}

func filterEpisodes(eps []tvd.Episode) []tvd.Episode {
	dbeps := []tvd.Episode{}
	for _, ep := range eps {
		if ep.EpisodeNumber != 0 {
			dbeps = append(dbeps, ep)
		} else {
			glog.Error("Filtering episode that doesn't have an episode number")
		}
	}
	return dbeps

}
func tvdbEpsToEpisodes(eps []tvd.Episode) []db.Episode {
	f := filterEpisodes(eps)
	dbeps := make([]db.Episode, len(f))
	for i, ep := range f {
		dbeps[i] = tvdbToEpisode(&ep)
	}
	return dbeps
}

// GetShow gets TVDB information for the given ID.
func (t *TvdbIndexer) GetShow(tvdbidstr string) (*db.Show, error) {
	tvdbid, _ := strconv.ParseInt(tvdbidstr, 10, 64)
	glog.Infof("Getting showid %d from tvdbid", tvdbid)
	series, eps, err := t.tvdbClient.SeriesAllByID(int(tvdbid), "en")
	if err != nil {
		return nil, err
	}
	dbshow := t.tvdbToShow(series)
	dbshow.Episodes = tvdbEpsToEpisodes(eps)
	return dbshow, nil
}

// Search searches TVDB for all shows matching the given string.
func (t *TvdbIndexer) Search(term string) ([]db.Show, error) {
	res, err := t.tvdbClient.SearchSeries(term, "en")
	if err != nil {
		return nil, err
	}
	dbshows := t.tvdbSeriesSummaryToShow(res)
	return dbshows, nil
}

func (t *TvdbIndexer) tvdbSeriesSummaryToShow(sums []tvd.SeriesSummary) []db.Show {
	dbshows := make([]db.Show, len(sums))
	for i, sum := range sums {
		dbshows[i] = db.Show{
			Name:        sum.Name,
			Indexer:     t.Name(),
			IndexerID:   int64(sum.ID),
			Language:    sum.Language,
			Description: sum.Overview,
			ImdbID:      sum.IMDBID,
			Network:     sum.Network,
		}
	}

	return dbshows
}

// tvdbToShow converts the struct returned by Tvdb and creates a new db.Show struct.
func (t *TvdbIndexer) tvdbToShow(ts *tvd.Series) *db.Show {
	s := &db.Show{}
	t.updateDbShowFromSeries(s, ts)
	return s
}

func (t *TvdbIndexer) updateDbShowFromSeries(dbshow *db.Show, ts *tvd.Series) {
	dbshow.Name = ts.Name
	dbshow.Genre = strings.Join(ts.Genre, "|")
	dbshow.Status = ts.Status
	dbshow.StartYear = ts.FirstAired.Year()
	dbshow.Indexer = t.Name()
	dbshow.IndexerID = int64(ts.ID)
	dbshow.Network = ts.Network
	dbshow.Language = ts.Language
	dbshow.Airs = ts.AirsTime
	dbshow.ImdbID = ts.IMDBID
	dbshow.LastIndexerUpdate = time.Now()
	if ts.Runtime.Valid {
		dbshow.Runtime = int64(ts.Runtime.Value)
	}
}

// TVDBToEpisode converts a TVDB episode record to a tv2go database episode
func tvdbToEpisode(episode *tvd.Episode) db.Episode {
	dbep := db.Episode{}
	updateDbEpisodeFromTvdb(&dbep, episode)
	return dbep
}

func updateDbEpisodeFromTvdb(dbep *db.Episode, tvep *tvd.Episode) {
	dbep.Name = tvep.EpisodeName
	dbep.AirDate = tvep.FirstAired.UTC()
	dbep.Description = tvep.Overview
	dbep.Season = int64(tvep.SeasonNumber)
	dbep.Episode = int64(tvep.EpisodeNumber)
	if tvep.AbsoluteNumber.Valid {
		dbep.AbsoluteNumber = int64(tvep.AbsoluteNumber.Value)
	}
}

// UpdateShow updates the give Database show from TVDB
func (t *TvdbIndexer) UpdateShow(dbshow *db.Show, episodes []db.Episode) error {
	ts, eps, err := t.tvdbClient.SeriesAllByID(int(dbshow.IndexerID), "en")
	if err != nil {
		return err
	}
	t.updateDbShowFromSeries(dbshow, ts)

	for _, episode := range filterEpisodes(eps) {
		glog.Infof("Updating S:%d, E:%d for '%s (tvdb id: %d)'", episode.SeasonNumber, episode.EpisodeNumber, dbshow.Name, dbshow.IndexerID)
		epToUpdate := db.Episode{}
		for _, dbep := range episodes {
			if dbep.Season == int64(episode.SeasonNumber) && dbep.Episode == int64(episode.EpisodeNumber) {
				glog.Infof("Found existing episode for S:%d, E:%d for '%s (tvdb id: %d)'",
					episode.SeasonNumber, episode.EpisodeNumber, dbshow.Name, dbshow.IndexerID)
				epToUpdate = dbep
				break
			}
		}
		if epToUpdate.ID == 0 {
			glog.Infof("tvdb: found new episode for show %s, S%d E%d", dbshow.Name, episode.SeasonNumber, episode.EpisodeNumber)
		}
		updateDbEpisodeFromTvdb(&epToUpdate, &episode)

		if epToUpdate.ID == 0 {
			dbshow.Episodes = append(dbshow.Episodes, epToUpdate)
		}
	}

	return nil
}

// TESTING FUNCTIONS

// NewTestTvdbIndexer returns a new configured indexer
func NewTestTvdbIndexer(options ...func(*TvdbIndexer)) (*TvdbIndexer, *httptest.Server) {
	r := gin.Default()
	r.GET("/api//series/71256/all/en.xml",
		newFileHandler("../indexers/tvdb/testdata/daily_show_all.xml").ServeXMLFile)
	r.GET("/api//series/78874/all/en.xml",
		newFileHandler("../indexers/tvdb/testdata/firefly_all.xml").ServeXMLFile)

	testTvdbServer := httptest.NewServer(r)

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(testTvdbServer.URL)
		},
	}

	testTvdbClient := &http.Client{Transport: transport}

	testTvdb := NewTvdbIndexer("", SetClient(testTvdbClient))

	return testTvdb, testTvdbServer
}

type fileHandler struct {
	io.ReadCloser
}

func newFileHandler(filename string) *fileHandler {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	return &fileHandler{
		ReadCloser: f,
	}
}
func (h *fileHandler) ServeXMLFile(c *gin.Context) {
	c.Set("Content-Type", "text/xml; charset=utf-8")
	io.Copy(c.Writer, h)
}
