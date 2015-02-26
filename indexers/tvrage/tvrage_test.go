package tvrage

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/golang/glog"
	. "github.com/onsi/gomega"
)

func TestGet(t *testing.T) {

	RegisterTestingT(t)
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
		} else if strings.Contains(r.RequestURI, "showinfo.php") {
			content, err := ioutil.ReadFile("testdata/buffy_showinfo.xml")
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
	rage := NewTVRageIndexer(SetClient(httpClient))

	shows, err := rage.Search("buffy")
	if err != nil {
		t.Fatalf("Error getting shows: %s", err)
	}
	Expect(len(shows)).To(Equal(3))

	showid := strconv.FormatInt(shows[0].ID, 10)
	showinfo, err := rage.GetShow(showid)
	if err != nil {
		t.Fatalf("Error getting show: %s", err)
	}

	Expect(showinfo.Name).To(Equal("Buffy the Vampire Slayer"))
}
