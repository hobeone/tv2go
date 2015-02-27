package providers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
)

type ProviderRegistry map[string]Provider

type ProviderResult struct {
	Type        string     `json:"type"`
	Age         *time.Time `json:"age,omitempty"` //hours
	Name        string     `json:"name"`
	Size        int64      `json:"size"`
	Quality     string     `json:"quality"`
	IndexerName string     `json:"indexer"`
	URL         string     `json:"url"`
}

// Provider defines the interface a tv2go provider must implement
type Provider interface {
	TvSearch(string, int64, int64) ([]ProviderResult, error)
	//need better name
	//Get file contents, leave it to something else to save it to disk
	GetURL(URL string) (string, []byte, error)
	// Return what kind of providers this is for: NZB/Torrent
	Type() ProviderType
}

// ProviderType is for the constants below
type ProviderType int

// Different kinds of providers
const (
	NZB ProviderType = 0 + iota
	TORRENT
)

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

func (n *NzbsOrg) Type() ProviderType {
	return NZB
}

func (n *NzbsOrg) GetURL(u string) (string, []byte, error) {
	glog.Infof("Getting URL '%s'", u)
	resp, err := n.Client.Get(u)
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
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return filename, b, err
	}
	return filename, b, nil
}

type tvSearchResponse struct {
	Channel struct {
		Items []struct {
			Category  string `json:"category"`
			Link      string `json:"link"`
			PubDate   string `json:"pubDate"`
			Title     string `json:"title"`
			Enclosure struct {
				Attributes struct {
					Length string `json:"length"`
					URL    string `json:"url"`
				} `json:"@attributes"`
			} `json:"enclosure"`
		} `json:"item"`
	} `json:"channel"`
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
	u.Add("o", "json")
	urlStr := u.Encode()

	queryURL, _ := url.Parse(n.URL)
	queryURL.RawQuery = urlStr
	resp, err := n.Client.Get(queryURL.String())

	if err != nil {
		return nil, fmt.Errorf("Error getting url '%s': %s\n", queryURL, err)
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response: %s", err)
	}
	parsedResponse := tvSearchResponse{}
	err = json.Unmarshal(respBody, &parsedResponse)
	if err != nil {
		glog.Infof("Couldn't parse '%s': %s", respBody, err.Error())
		return nil, fmt.Errorf("Error parsing feed: %s", err)
	}

	results := make([]ProviderResult, len(parsedResponse.Channel.Items))
	for i, story := range parsedResponse.Channel.Items {
		parsedTime := &time.Time{}
		pt, err := time.Parse(time.RFC1123Z, story.PubDate)
		parsedTime = &pt
		if err != nil {
			glog.Warningf("Couldn't parse time '%s': %s", story.PubDate, err.Error())
			parsedTime = nil
		}
		size, err := strconv.ParseInt(story.Enclosure.Attributes.Length, 10, 64)
		if err != nil {
			glog.Warningf("Couldn't parse size to int: '%s': %s", err.Error())
			size = 0
		}

		results[i] = ProviderResult{
			Type:        "NZB",
			Age:         parsedTime,
			Name:        story.Title,
			IndexerName: "nzbsOrg",
			URL:         story.Link,
			Size:        size,
		}
	}
	return results, nil
}
