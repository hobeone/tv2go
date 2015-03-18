package providers

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/golang/glog"
)

// ProviderRegistry provides an easy way to map providers to string names
type ProviderRegistry map[string]Provider

func (pr ProviderRegistry) Search(showname string, season, epnum int64) []ProviderResult {
	res := []ProviderResult{}
	for _, provider := range pr {
		resultset, err := provider.TvSearch(showname, season, epnum)
		if err == nil {
			res = append(res, resultset...)
		}
	}
	return res
}

// ProviderResult describes the information that Providers will return from searches
type ProviderResult struct {
	Type         string     `json:"type"`
	Age          *time.Time `json:"age,omitempty"` //hours
	Name         string     `json:"name"`
	Size         int64      `json:"size"`
	Quality      string     `json:"quality"`
	ProviderName string     `json:"indexer"`
	URL          string     `json:"url"`
	Seeders      int64      `json:"seeders"`
	TVRageID     int64      `json:"tvrage_id"`
	TVDBID       int64      `json:"tvdb_id"`
	Season       string     `json:"season"`
	Episode      string     `json:"episode"`
}

// Provider defines the interface a tv2go provider must implement
type Provider interface {
	Name() string

	TvSearch(string, int64, int64) ([]ProviderResult, error)
	//need better name
	//Get file contents, leave it to something else to save it to disk
	GetURL(URL string) (string, []byte, error)
	// Return what kind of providers this is for: NZB/Torrent
	Type() ProviderType

	// Get new items on the provider.  Will usually mean hitting a rss feed or
	// something.
	GetNewItems() ([]ProviderResult, error)
}

// BaseProvider is the struct used for shared functionality of all providers.
type BaseProvider struct {
	ProviderName string
	Client       *http.Client
	PollInterval time.Duration
	after        func(time.Duration) <-chan (time.Time)
}

func NewBaseProvider(name string) *BaseProvider {
	return &BaseProvider{
		ProviderName: name,
		Client:       &http.Client{},
		after:        time.After,
		PollInterval: time.Minute * 15, // reasonable default
	}
}

func (b *BaseProvider) Name() string {
	return b.ProviderName
}

//GetURL is designed to be used to download a file from a URL
func (b *BaseProvider) GetURL(u string) (string, []byte, error) {
	glog.Infof("Getting URL %s", u)
	resp, err := b.Client.Get(u)
	if err != nil {
		return "", nil, err
	}

	filename := ""
	contHeader := resp.Header.Get("Content-Disposition")
	res := strings.Split(contHeader, "; ")
	for _, res := range res {
		if strings.HasPrefix(res, "filename=") {
			filename = strings.Split(res, "=")[1]
		}
	}

	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return filename, content, err
	}
	return filename, content, nil
}

func (p *BaseProvider) AfterWithJitter(d time.Duration) <-chan time.Time {
	s := d + time.Duration(rand.Int63n(60))*time.Second
	fmt.Printf("%v\n", s)
	return p.after(d)
}

// TorrentProvider is the base type for Torrent based providers
type TorrentProvider struct {
	*BaseProvider
}

func (t *TorrentProvider) Type() ProviderType {
	return TORRENT
}

type NZBProvider struct {
	*BaseProvider
}

func (t *NZBProvider) Type() ProviderType {
	return NZB
}

// ProviderType is for the constants below
type ProviderType int

// String() function will return the english name
// that we want out constant Day be recognized as
func (t ProviderType) String() string {
	return types[t]
}

//ProviderTypeFromString converts a string name to a ProviderType
func ProviderTypeFromString(s string) (ProviderType, error) {
	for i, pt := range types {
		if pt == s {
			return ProviderType(i), nil
		}
	}
	return UNKNOWN, fmt.Errorf("Unknown Provider Type: %s", s)
}

// Different kinds of providers
const (
	NZB ProviderType = 0 + iota
	TORRENT
	UNKNOWN
)

var types = [...]string{
	"NZB",
	"TORRENT",
	"UNKNOWN",
}
