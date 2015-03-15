package web

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	. "github.com/onsi/gomega"

	"github.com/gin-gonic/gin"
	"github.com/hobeone/tv2go/config"
	"github.com/hobeone/tv2go/db"
	"github.com/hobeone/tv2go/indexers"
	"github.com/hobeone/tv2go/indexers/tvdb"
	"github.com/hobeone/tv2go/providers"
	"github.com/hobeone/tv2go/storage"
)

var basedir, _ = filepath.Abs("")

func setupTest(t *testing.T) (*db.Handle, *Server) {
	//flag.Set("logtostderr", "true")
	gin.SetMode("test")

	dbh := db.NewMemoryDBHandle(false, true)

	broker, err := storage.NewBroker("testdata")
	if err != nil {
		t.Fatalf("Error creating storage broker: %s", err.Error())
	}
	providers := providers.ProviderRegistry{}

	s := NewServer(config.NewTestConfig(), dbh, broker, providers)
	return dbh, s
}

var GoldenShowResponse = fmt.Sprintf(`{
	"id": 1,
	"air_by_date": false,
	"airs": "",
	"cache": {
		"Banner": 0,
		"Poster": 0
	},
	"anime": false,
	"indexerid": 1,
	"language": "",
	"network": "",
	"paused": false,
	"quality_group": "HDALL",
	"name": "show1",
	"sports": false,
	"status": "",
	"subtitles": false,
	"tvdbid": 1,
	"tvrage_id": 0,
	"tvrage_name": "",
	"location": "%s/testdata/show1"
}`, basedir)

func TestShow(t *testing.T) {
	dbh, eng := setupTest(t)
	db.LoadFixtures(t, dbh)
	RegisterTestingT(t)

	// Invalid
	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/1/shows/NOTFOUND", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.Handler.ServeHTTP(response, req)
	if response.Code != 500 {
		t.Fatalf("Expected 500 response code, got %d", response.Code)
	}

	// Unknown
	response = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/api/1/shows/0", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.Handler.ServeHTTP(response, req)
	if response.Code != 404 {
		t.Fatalf("Expected 404 response code, got %d", response.Code)
	}

	// Success
	response = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/api/1/shows/1", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.Handler.ServeHTTP(response, req)
	if response.Code != 200 {
		t.Fatalf("Expected 200 response code, got %d", response.Code)
	}
	Expect(response.Body.String()).Should(MatchJSON(GoldenShowResponse))
}

var ShowsGoldenResp = fmt.Sprintf(`[
{
	"id": 1,
	"air_by_date": false,
	"airs": "",
	"cache": {
		"Banner": 0,
		"Poster": 0
	},
	"anime": false,
	"indexerid": 1,
	"language": "",
	"network": "",
	"paused": false,
	"quality_group": "HDALL",
	"name": "show1",
	"sports": false,
	"status": "",
	"subtitles": false,
	"tvdbid": 1,
	"tvrage_id": 0,
	"tvrage_name": "",
	"location": "%s/testdata/show1"
},
{
	"id": 2,
	"air_by_date": false,
	"airs": "",
	"cache": {
		"Banner": 0,
		"Poster": 0
	},
	"anime": false,
	"indexerid": 2,
	"language": "",
	"network": "",
	"paused": false,
	"quality_group": "",
	"name": "show2",
	"sports": false,
	"status": "",
	"subtitles": false,
	"tvdbid": 2,
	"tvrage_id": 0,
	"tvrage_name": "",
	"location": "%s/testdata/show2"
}
]`, basedir, basedir)

func TestShows(t *testing.T) {
	dbh, eng := setupTest(t)
	db.LoadFixtures(t, dbh)
	RegisterTestingT(t)

	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/1/shows", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.Handler.ServeHTTP(response, req)
	if response.Code != 200 {
		t.Fatalf("Expected 200 response code, got %d", response.Code)
	}
	Expect(response.Body.String()).Should(MatchJSON(ShowsGoldenResp))
}

const EpisodeGolden = `{
	"id": 1,
	"showid": 1,
	"name": "show1episode1",
	"season": 1,
	"episode": 1,
	"airdate": "2006-01-01",
	"description": "",
	"file_size": 0,
	"file_size_human": "",
	"location": "",
	"quality": "Unknown",
	"release_name": "",
	"status": "WANTED"
}`

func TestEpisode(t *testing.T) {
	dbh, eng := setupTest(t)
	db.LoadFixtures(t, dbh)
	RegisterTestingT(t)

	//Success
	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/1/shows/1/episodes/1", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.Handler.ServeHTTP(response, req)
	if response.Code != 200 {
		t.Fatalf("Expected 200 response code, got %d", response.Code)
	}
	Expect(response.Body.String()).Should(MatchJSON(EpisodeGolden))

	//invalid - missing show id
	response = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/api/1/shows/episodes/1", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.Handler.ServeHTTP(response, req)
	if response.Code != 405 {
		t.Fatalf("Expected 405 response code, got %d", response.Code)
	}
	Expect(response.Body.String()).Should(Equal("Method Not Allowed\n"))

	//Valid - showid is ignored
	response = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/api/1/shows/XX/episodes/1", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.Handler.ServeHTTP(response, req)
	if response.Code != 200 {
		t.Fatalf("Expected 200 response code, got %d", response.Code)
	}
	Expect(response.Body.String()).Should(MatchJSON(EpisodeGolden))

	//nonexisting episode
	response = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/api/1/shows/1/episodes/100000", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.Handler.ServeHTTP(response, req)
	if response.Code != 404 {
		t.Fatalf("Expected 404 response code, got %d", response.Code)
	}
	Expect(response.Body.String()).Should(MatchJSON(`{
		"message": "record not found",
		"result": "failure"
	}`))
}

func TestAddShow(t *testing.T) {
	dbh, eng := setupTest(t)
	db.LoadFixtures(t, dbh)
	RegisterTestingT(t)
	tvdbIndexer, server := tvdb.NewTestTvdbIndexer()
	eng.indexers = indexers.IndexerRegistry{
		"tvdb": tvdbIndexer,
	}
	defer server.Close()

	//Success
	response := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/api/1/shows", strings.NewReader(`{"indexer_name":"tvdb","indexerid":"78874"}\n`))
	req.Header.Add("content-type", "application/json;charset=UTF-8")
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.Handler.ServeHTTP(response, req)
	if response.Code != 200 {
		spew.Dump(response.Body)
		t.Fatalf("Expected 200 response code, got %d", response.Code)
	}
}

func TestShowUpdateFromDisk(t *testing.T) {
	dbh, eng := setupTest(t)
	db.LoadFixtures(t, dbh)
	RegisterTestingT(t)

	//Success
	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/1/shows/1/rescan", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.Handler.ServeHTTP(response, req)
	if response.Code != 200 {
		t.Fatalf("Expected 200 response code, got %d", response.Code)
	}
	Expect(response.Body.String()).Should(MatchJSON(GoldenShowResponse))
}

func TestNameCleaner(t *testing.T) {
	RegisterTestingT(t)

	dir := "/"

	teststr := showToLocation(dir, "a/b/c")
	Expect(teststr).To(Equal("/a-b-c"))

	teststr = showToLocation(dir, "abc")
	Expect(teststr).To(Equal("/abc"))

	teststr = showToLocation(dir, "a\"c")
	Expect(teststr).To(Equal("/ac"))

	teststr = showToLocation(dir, ".a.b..")
	Expect(teststr).To(Equal("/a.b"))

	teststr = showToLocation(dir, ".a.b (YEAR)")
	Expect(teststr).To(Equal("/a.b (YEAR)"))

	teststr = showToLocation(dir, "Marvel's Agents of S.H.I.E.L.D")
	Expect(teststr).To(Equal("/Marvel's Agents of S.H.I.E.L.D"))
}

func TestQualityGroupList(t *testing.T) {
	dbh, eng := setupTest(t)
	db.LoadFixtures(t, dbh)
	RegisterTestingT(t)

	//Success
	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/1/quality_groups", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.Handler.ServeHTTP(response, req)
	if response.Code != 200 {
		t.Fatalf("Expected 200 response code, got %d", response.Code)
	}
}
