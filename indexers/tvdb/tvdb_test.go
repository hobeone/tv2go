package tvdb

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	. "github.com/onsi/gomega"
)

func testTools(code int, body string) (*httptest.Server, *http.Client) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprintln(w, body)
	}))

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	httpClient := &http.Client{Transport: transport}

	return server, httpClient
}

func TestGetShowById(t *testing.T) {
	RegisterTestingT(t)
	content, err := ioutil.ReadFile("testdata/daily_show_all.xml")
	if err != nil {
		glog.Fatalf("Error reading test feed: %s", err.Error())
	}

	httpserver, httpclient := testTools(200, string(content))
	defer httpserver.Close()

	client := NewTvdbIndexer("90D7DF3AE9E4841E", SetClient(httpclient))
	show, eps, err := client.GetShowByID(100)
	Expect(err).ToNot(HaveOccurred(), "Error getting show: %s", err)
	spew.Dump(show)
	Expect(len(eps)).Should(Equal(2669), "Eps is too long")
}
