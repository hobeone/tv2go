package providers

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/onsi/gomega"
)

func TestNyaa(t *testing.T) {
	RegisterTestingT(t)
	//flag.Set("logtostderr", "true")
	body, err := ioutil.ReadFile("testdata/nyaa_yowamushi.rss")
	if err != nil {
		t.Fatalf("Error reading test file %s", err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(200)
		w.Write(body)
	}))
	defer server.Close()

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	httpClient := &http.Client{Transport: transport}

	n := NewNyaaTorrents()
	n.Client = httpClient
	res, err := n.TvSearch("Yowamushi Pedal", 1, 1)
	Expect(err).ToNot(HaveOccurred())
	Expect(len(res)).To(Equal(100))
}
