package providers

import (
	"encoding/xml"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/golang/glog"
	"github.com/hobeone/tv2go/rss"
)

type NyaaTorrents struct {
	URL string
	TorrentProvider
}

// NewNzbsOrg creates a new nzbs.org client
func NewNyaaTorrents(options ...func(*NyaaTorrents)) *NyaaTorrents {
	n := &NyaaTorrents{
		URL: "http://www.nyaa.se/",
		TorrentProvider: TorrentProvider{
			BaseProvider: NewBaseProvider("nyaaTorrents"),
		},
	}
	for _, option := range options {
		option(n)
	}
	return n
}

var descRegex = regexp.MustCompile(`^(?i)(?P<seeders>\d+)\s+seeder\(s\), (?P<leechers>\d+)\s+leecher\(s\),\s+(?P<downloads>\d+)\s+download\(s\) - (?P<size>[^-]+) -`)

func getMetadata(desc string) map[string]string {
	match := descRegex.FindStringSubmatch(desc)
	result := make(map[string]string)
	if match == nil {
		return result
	}
	for i, name := range descRegex.SubexpNames() {
		result[name] = match[i]
	}
	return result
}

func (n *NyaaTorrents) TvSearch(showName string, season, ep int64) ([]ProviderResult, error) {
	u := url.Values{}
	u.Add("page", "rss")
	u.Add("term", showName)
	u.Add("sort", "2")    // descending by seeders
	u.Add("cats", "1_37") // eng translated anime
	urlStr := u.Encode()

	queryURL, _ := url.Parse(n.URL)
	queryURL.RawQuery = urlStr

	glog.Infof("Searching NyaaTorrents with %s", queryURL.String())
	resp, err := n.Client.Get(queryURL.String())
	if err != nil {
		glog.Errorf("Error searching nyaaTorrents: %s", err)
		return nil, err
	}

	r := rss.Rss{}
	defer resp.Body.Close()
	d := xml.NewDecoder(resp.Body)
	d.Strict = false
	//d.CharsetReader = charset.NewReader
	d.DefaultSpace = "DefaultSpace"
	d.Entity = xml.HTMLEntity

	err = d.Decode(&r)
	if err != nil {
		glog.Errorf("Error decoding nyaaTorrents response: %s", err)
		return nil, err
	}

	glog.Infof("Got %d items from NyaaTorrents", len(r.Items))
	results := make([]ProviderResult, len(r.Items))
	for i, item := range r.Items {
		parsedTime := &time.Time{}
		pt, err := time.Parse(time.RFC1123Z, item.PubDate)
		parsedTime = &pt
		if err != nil {
			glog.Warningf("Couldn't parse time '%s': %s", item.PubDate, err.Error())
			parsedTime = nil
		}

		meta := getMetadata(item.Description)
		bytes, _ := humanize.ParseBytes(meta["size"])

		seeders, _ := strconv.ParseInt(meta["seeders"], 10, 64)

		results[i] = ProviderResult{
			Name:         item.Title,
			URL:          item.Link,
			Size:         int64(bytes),
			Seeders:      seeders,
			Age:          parsedTime,
			Type:         n.Type().String(),
			ProviderName: n.Name(),
		}
	}

	return results, nil
}

// GetNewItems scans the RSS feed for all new items.
func (n *NyaaTorrents) GetNewItems() ([]ProviderResult, error) {
	return []ProviderResult{}, nil
}
