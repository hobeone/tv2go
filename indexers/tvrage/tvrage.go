package tvrage

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/hobeone/tv2go/db"
	tvr "github.com/hobeone/tvrage"
)

// TVRageIndexer implements the Indexer interface using the tvdb client
type TVRageIndexer struct {
	httpClient *http.Client
}

// NewTVRageIndexer returns a new indexer
func NewTVRageIndexer(options ...func(*TVRageIndexer)) *TVRageIndexer {
	t := &TVRageIndexer{
		httpClient: &http.Client{},
	}
	for _, option := range options {
		option(t)
	}
	return t
}

// SetClient set's the httpclient the Indexer will use.
//
// Example:
//  NewTvdbIndexer(SetClient(httpclient))
func SetClient(c *http.Client) func(*TVRageIndexer) {
	return func(t *TVRageIndexer) {
		tvr.Client = c
		t.httpClient = c
	}
}

// Returns the string name of this indexer.
func (t *TVRageIndexer) Name() string {
	return "tvrage"
}

// Search returns matches from TVRage for the given name
func (t *TVRageIndexer) Search(name string) ([]db.Show, error) {
	shows, err := tvr.Search(name)
	if err != nil {
		return nil, err
	}

	dbshows := make([]db.Show, len(shows))
	for i, show := range shows {
		dbshows[i] = tvrageToShow(&show)
	}

	return dbshows, nil
}

// GetShow returns show information (show + episodes) for the given id.
func (t *TVRageIndexer) GetShow(showid string) (*db.Show, error) {
	rageid, err := strconv.ParseInt(showid, 10, 64)
	if err != nil {
		return nil, err
	}

	glog.Infof("Getting showid %d from TVRage.", rageid)
	show, err := tvr.Get(rageid)
	if err != nil {
		glog.Errorf("Error getting showid %d from TVRage: %s", rageid, err)
		return &db.Show{}, err
	}

	dbshow := tvrageToShow(show)

	eps, err := tvr.EpisodeList(int(rageid))
	if err != nil {
		glog.Errorf("Error getting episodes for showid '%d' from TVRage: %s", rageid, err.Error())
		return &dbshow, err
	}

	dbshow.Episodes = make([]db.Episode, len(eps))
	for i, ep := range eps {
		dbshow.Episodes[i] = tvrageToEp(&ep)
	}

	return &dbshow, nil
}

// UpdateShow updates the give Database show from TVRage
func (t *TVRageIndexer) UpdateShow(dbshow *db.Show) error {
	return nil
}

func tvrageToShow(ts *tvr.Show) db.Show {
	s := db.Show{
		Name:           ts.Name,
		Genre:          strings.Join(ts.Genres, "|"),
		Classification: ts.Classification,
		Status:         ts.Status,
		StartYear:      int(ts.Started),
		IndexerID:      ts.ID,
		Indexer:        "tvrage",
	}
	return s
}

func tvrageToEp(te *tvr.Episode) db.Episode {
	e := db.Episode{
		Name:           te.Title,
		AirDate:        te.AirDate.UTC(),
		Episode:        int64(te.Number),
		Season:         int64(te.Season),
		AbsoluteNumber: int64(te.Ordinal),
	}
	return e
}
