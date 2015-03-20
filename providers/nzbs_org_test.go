package providers

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

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

/*
func TestPolling(t *testing.T) {

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

	RegisterTestingT(t)
	n := NewNzbsOrg("API_KEY")
	n.URL = server.URL

	retchan := make(chan (ProviderResult))
	poller := NewProviderPoller(n, time.Minute*15, time.Time{}, retchan)
	reader := func() {
		for {
			spew.Dump(<-retchan)
		}
	}
	go reader()
	go poller.Poll()
}
*/
