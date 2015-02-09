package tvrage

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	tvr "github.com/drbig/tvrage"

	"github.com/golang/glog"
)

func TestGet(t *testing.T) {

	// Test server that always responds with 200 code, and specific payload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.RequestURI)
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.RequestURI, "search.php") {
			content, err := ioutil.ReadFile("testdata/buffy_search.xml")
			if err != nil {
				glog.Fatalf("Error reading test feed: %s", err.Error())
			}
			w.Write(content)
		} else if strings.Contains(r.RequestURI, "episode_list.php") {
			content, err := ioutil.ReadFile("testdata/buffy_episode_list.xml")
			if err != nil {
				glog.Fatalf("Error reading test feed: %s", err.Error())
			}
			w.Write(content)

		}
	}))
	defer server.Close()

	// Make a transport that reroutes all traffic to the example server
	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	// Make a http.Client with the transport
	httpClient := &http.Client{Transport: transport}
	tvr.Client = httpClient

	shows, err := SearchShowByName("buffy")
	if err != nil {
		t.Fatalf("Error getting shows: %s", err)
	}
	showinfo, err := GetShowInfo(shows[0].ID)
	if err != nil {
		t.Fatalf("Error getting show: %s", err)
	}
	spew.Dump(err)
	spew.Dump(showinfo)
}
