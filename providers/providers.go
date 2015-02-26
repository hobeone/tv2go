package providers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/hobeone/rss2go/feed"
)

type SearchResult struct {
	Show    string
	Episode string
	Url     string
	Query   string
}

type ProviderRegistry map[string]Provider

type ProviderResult struct {
	Type        string `json:"type"`
	Age         int64  `json:"age"` //hours
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	Quality     string `json:"quality"`
	IndexerName string `json:"indexer"`
	URL         string `json:"url"`
}

type Provider interface {
	TvSearch(string, int64, int64) ([]ProviderResult, error)
}

type NzbsOrg struct {
	URL    string
	APIKEY string
	Client *http.Client
	// TODO: figure out how to set up an interface for this:
	//Logger glog.Logger
}

func NewNzbsOrg(key string, options ...func(*NzbsOrg)) *NzbsOrg {
	n := &NzbsOrg{
		APIKEY: key,
		URL:    "https://nzbs.org/api",
		Client: &http.Client{},
	}
	for _, option := range options {
		option(n)
	}
	return n
}

func SetClient(c *http.Client) func(*NzbsOrg) {
	return func(n *NzbsOrg) {
		n.Client = c
	}
}

// TvSearch searches for a given tv show with optional episode and season
// constraints.
//
// API: t=tvsearch&q=beverly%20hillbillies&season=3&ep=1
//  ?t=tvsearch&rid=5615&cat=5030,5070. Include &extended=1 to return extended information in the search results.
func (n *NzbsOrg) TvSearch(showName string, season, ep int64) ([]ProviderResult, error) {
	u := url.Values{}
	u.Add("apikey", n.APIKEY)
	u.Add("t", "tvsearch")
	u.Add("q", showName)
	u.Add("season", strconv.FormatInt(season, 10))
	u.Add("ep", strconv.FormatInt(ep, 10))
	urlStr := u.Encode()
	spew.Dump(urlStr)

	queryUrl, _ := url.Parse(n.URL)
	queryUrl.RawQuery = urlStr
	spew.Dump(queryUrl)
	resp, err := n.Client.Get(queryUrl.String())

	if err != nil {
		return nil, fmt.Errorf("Error getting url '%s': %s\n", urlStr, err)
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response: %s", err)
	}

	_, stories, err := feed.ParseFeed(queryUrl.String(), respBody)
	if err != nil {
		return nil, fmt.Errorf("Error parsing feed: %s", err)
	}

	parsedRes := make([]ProviderResult, len(stories))
	for i, story := range stories {
		parsedRes[i] = ProviderResult{
			Type:        "NZB",
			Age:         int64(time.Since(story.Published).Hours()),
			Name:        story.Title,
			IndexerName: "nzbsOrg",
			URL:         story.Link,
		}
	}
	return parsedRes, nil
}
