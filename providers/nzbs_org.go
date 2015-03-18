package providers

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/hobeone/tv2go/rss"
	"golang.org/x/net/html/charset"
)

// NzbsOrg provides a client for the nzbs.org provider
type NzbsOrg struct {
	URL    string
	APIKEY string
	NZBProvider
}

// NewNzbsOrg creates a new nzbs.org client
func NewNzbsOrg(key string, options ...func(*NzbsOrg)) *NzbsOrg {
	n := &NzbsOrg{
		APIKEY: key,
		URL:    "https://nzbs.org/api",
		NZBProvider: NZBProvider{
			BaseProvider: NewBaseProvider("nzbsOrg"),
		},
	}
	for _, option := range options {
		option(n)
	}
	return n
}

// SetClient is used in the NewNzbsOrg constructor to set the http.Client to
// use for all http/s calls.
func SetClient(c *http.Client) func(*NzbsOrg) {
	return func(n *NzbsOrg) {
		n.Client = c
	}
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

func (n *NzbsOrg) GetNewItems() ([]ProviderResult, error) {
	u := url.Values{}
	u.Add("apikey", n.APIKEY)
	u.Add("t", "tvsearch")
	u.Add("cat", "5030,5040,5060,5070")
	u.Add("attrs", "rageid,tvdbid,season,episode")
	urlStr := u.Encode()

	queryURL, _ := url.Parse(n.URL)
	queryURL.RawQuery = urlStr
	glog.Infof("nzbs.org: Getting new items with %s", queryURL.String())
	resp, err := n.Client.Get(queryURL.String())

	if err != nil {
		return nil, fmt.Errorf("Error getting url '%s': %s\n", queryURL, err)
	}

	defer resp.Body.Close()

	r := rss.Rss{}
	defer resp.Body.Close()
	d := xml.NewDecoder(resp.Body)
	d.Strict = false
	d.CharsetReader = charset.NewReaderByName
	d.DefaultSpace = "DefaultSpace"
	d.Entity = xml.HTMLEntity

	err = d.Decode(&r)
	if err != nil {
		glog.Errorf("Error decoding nzbs.org response: %s", err)
		return nil, err
	}

	results := make([]ProviderResult, len(r.Items))
	for i, story := range r.Items {
		parsedTime := &time.Time{}
		pt, err := time.Parse(time.RFC1123Z, story.PubDate)
		parsedTime = &pt
		if err != nil {
			glog.Warningf("Couldn't parse time '%s': %s", story.PubDate, err.Error())
			parsedTime = nil
		}
		size := parseIntOrZero(story.Enclosure.Length)
		var rageID int64
		var tvdbID int64
		for _, attr := range story.NewzNabAttr {
			switch attr.Name {
			case "rageid":
				rageID = parseIntOrZero(attr.Value)
			case "tvdbid":
				tvdbID = parseIntOrZero(attr.Value)
			}
		}

		results[i] = ProviderResult{
			Type:         n.Type().String(),
			Age:          parsedTime,
			Name:         story.Title,
			ProviderName: n.Name(),
			URL:          story.Link,
			Size:         size,
			TVRageID:     rageID,
			TVDBID:       tvdbID,
		}
	}
	return results, nil
}

func parseIntOrZero(str string) (i int64) {
	if str == "" {
		return
	}
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		glog.Warningf("Couldn't parse to int: '%s': %s", err.Error())
		i = 0
	}
	return
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
	glog.Infof("Searching nzbs.org with %s", queryURL.String())
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
			Type:         "NZB",
			Age:          parsedTime,
			Name:         story.Title,
			ProviderName: n.Name(),
			URL:          story.Link,
			Size:         size,
		}
	}
	return results, nil
}
