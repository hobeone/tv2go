package web

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/gin-gonic/gin"
	"github.com/hobeone/tv2go/db"
)

func setupTest(t *testing.T) (*db.Handle, *gin.Engine) {
	flag.Set("logtostderr", "false")
	gin.SetMode("test")

	dbh := db.NewMemoryDBHandle(false, true)
	e := createServer(dbh)
	return dbh, e
}

const GoldenShowResponse = `{
	"id":1,
		"air_by_date":0,
		"cache":{
			"Banner":0,
			"Poster":0
		},
		"anime":0,
		"indexerid":1,
		"language":"",
		"network":"",
		"next_ep_airdate":"",
		"paused":0,
		"quality":"0",
		"name":"show1",
		"sports":0,
		"status":"",
		"subtitles":0,
		"tvdbid":1,
		"tvrage_id":0,
		"tvrage_name":"",
		"season_list":[1]
}`

func TestShow(t *testing.T) {
	dbh, eng := setupTest(t)
	db.LoadFixtures(t, dbh)
	RegisterTestingT(t)

	// Invalid
	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/1/shows/NOTFOUND", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.ServeHTTP(response, req)
	if response.Code != 500 {
		t.Fatalf("Expected 500 response code, got %d", response.Code)
	}

	// Unknown
	response = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/api/1/shows/0", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.ServeHTTP(response, req)
	if response.Code != 404 {
		t.Fatalf("Expected 404 response code, got %d", response.Code)
	}

	// Success
	response = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/api/1/shows/1", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.ServeHTTP(response, req)
	if response.Code != 200 {
		t.Fatalf("Expected 200 response code, got %d", response.Code)
	}
	Expect(response.Body.String()).Should(MatchJSON(GoldenShowResponse))
}

const ShowsGoldenResp = `[
		{
			"id": 1,
			"air_by_date":0,
			"cache":{
				"Banner":0,
				"Poster":0
			},
			"anime":0,
			"indexerid":1,
			"language":"",
			"network":"",
			"next_ep_airdate":"",
			"paused":0,
			"quality":"0",
			"name":"show1",
			"sports":0,
			"status":"",
			"subtitles":0,
			"tvdbid":1,
			"tvrage_id":0,
			"tvrage_name":"",
			"season_list":null
		},
		{
			"id": 2,
			"air_by_date":0,
			"cache":{
				"Banner":0,
				"Poster":0
			},
			"anime":0,
			"indexerid":2,
			"language":"",
			"network":"",
			"next_ep_airdate":"",
			"paused":0,
			"quality":"0",
			"name":"show2",
			"sports":0,
			"status":"",
			"subtitles":0,
			"tvdbid":2,
			"tvrage_id":0,
			"tvrage_name":"",
			"season_list":null
		}
	]`

func TestShows(t *testing.T) {
	dbh, eng := setupTest(t)
	db.LoadFixtures(t, dbh)
	RegisterTestingT(t)

	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/1/shows", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.ServeHTTP(response, req)
	if response.Code != 200 {
		t.Fatalf("Expected 200 response code, got %d", response.Code)
	}
	Expect(response.Body.String()).Should(MatchJSON(ShowsGoldenResp))
}

const EpisodeGolden = `{
	"id": 1,
	"showid": 1,
	"name": "show1episode1",
	"season": 0,
	"episode": 0,
	"airdate": "2006-01-01",
	"description": "",
	"file_size": 0,
	"file_size_human": "",
	"location": "",
	"quality": "",
	"release_name": "",
	"status": ""
}`

func TestEpisode(t *testing.T) {
	dbh, eng := setupTest(t)
	db.LoadFixtures(t, dbh)
	RegisterTestingT(t)

	//Success
	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/1/shows/1/episodes/1", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.ServeHTTP(response, req)
	if response.Code != 200 {
		t.Fatalf("Expected 200 response code, got %d", response.Code)
	}
	Expect(response.Body.String()).Should(MatchJSON(EpisodeGolden))

	//invalid - missing show id
	response = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/api/1/shows/episodes/1", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.ServeHTTP(response, req)
	if response.Code != 405 {
		t.Fatalf("Expected 405 response code, got %d", response.Code)
	}
	Expect(response.Body.String()).Should(Equal("Method Not Allowed\n"))

	//Valid - showid is ignored
	response = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/api/1/shows/XX/episodes/1", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.ServeHTTP(response, req)
	if response.Code != 200 {
		t.Fatalf("Expected 200 response code, got %d", response.Code)
	}
	Expect(response.Body.String()).Should(MatchJSON(EpisodeGolden))

	//nonexisting episode
	response = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/api/1/shows/1/episodes/100000", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.ServeHTTP(response, req)
	if response.Code != 404 {
		t.Fatalf("Expected 404 response code, got %d", response.Code)
	}
	Expect(response.Body.String()).Should(MatchJSON(`{
		"message": "Record Not Found",
		"result": "failure"
	}`))
}
