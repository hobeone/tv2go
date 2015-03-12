package providers

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestNzbsOrg(t *testing.T) {
	//n := NewNzbsOrg("API_KEY", SetClient(&http.Client{}))
	//n.TvSearch("Archer (2009)", 5, 1)
	//
	content, err := ioutil.ReadFile("testdata/nzbs_org_archer.json")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}

	str := tvSearchResponse{}
	err = json.Unmarshal(content, &str)
	if err != nil {
		t.Fatalf("Error unmarshaling file: %s", err)
	}

	for _, item := range str.Channel.Items {
		fmt.Println(item.Title)
		fmt.Println(item.Link)
		fmt.Println(item.PubDate)
		pt, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			t.Fatalf("Error unmarshaling file: %s", err)
		}
		fmt.Println(pt)

		fmt.Println(item.Category)
		fmt.Println(item.Enclosure.Attributes.Length)
	}
}

func TestNzbsOrgGetURL(t *testing.T) {
	//flag.Set("logtostderr", "true")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Header().Set("Content-Disposition", "attachment; filename=test.nzb")
		w.WriteHeader(200)
		fmt.Fprintln(w, "testing123")
	}))
	defer server.Close()

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	httpClient := &http.Client{Transport: transport}

	n := NewNzbsOrg("API_KEY", SetClient(httpClient))
	fname, content, err := n.GetURL("http://testing/download/nzb")
	if err != nil {
		t.Fatalf("Error downloading url: %s", err)
	}
	if fname != "test.nzb" {
		t.Fatalf("Expected filename to = 'test.nzb', got '%s'", fname)
	}
	if strings.TrimSpace(string(content)) != "testing123" {
		t.Fatalf("Expected content to be 'testing123' got '%s'", string(content))
	}
}

func TestNzbsOrgGetNewItems(t *testing.T) {
	RegisterTestingT(t)
	body, err := ioutil.ReadFile("testdata/nzbs_org_feed.rss")
	if err != nil {
		t.Fatalf("Error reading test file %s", err)
	}
	flag.Set("logtostderr", "true")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(200)
		w.Write(body)
	}))
	defer server.Close()

	n := NewNzbsOrg("API_KEY")
	n.URL = server.URL

	res, err := n.GetNewItems()
	if err != nil {
		t.Fatalf("Error getting new items: %s", err)
	}
	Expect(res).To(HaveLen(100))
}
