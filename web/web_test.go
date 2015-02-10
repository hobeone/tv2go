package web

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
	. "github.com/onsi/gomega"

	"github.com/gin-gonic/gin"
	"github.com/hobeone/tv2go/db"
)

func setupTest(t *testing.T) (*db.Handle, *gin.Engine) {
	dbh := db.NewMemoryDBHandle(true, true)
	e := createServer(dbh)
	return dbh, e
}

func TestPing(t *testing.T) {
	dbh, eng := setupTest(t)
	db.LoadFixtures(t, dbh)
	RegisterTestingT(t)

	pingGoldenResponse := fmt.Sprintf(`{
	"data":{
		"pid":%d
	},
	"message":"Pong",
	"result":"success"
}`, os.Getpid())

	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/1/?cmd=sb.ping", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.ServeHTTP(response, req)
	if response.Code != 200 {
		t.Fatalf("Expected 200 response code, got %d", response.Code)
	}

	Expect(response.Body.String()).Should(MatchJSON(pingGoldenResponse))
}

const GoldenShowResponse = `{
	"data":{
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
		"show_name":"show1",
		"sports":0,
		"status":"",
		"subtitles":0,
		"tvdbid":1,
		"tvrage_id":0,
		"tvrage_name":"",
		"season_list":[1]
	},
	"message":"",
	"result":"success"
}`

func TestShow(t *testing.T) {
	dbh, eng := setupTest(t)
	db.LoadFixtures(t, dbh)
	RegisterTestingT(t)

	// Invalid
	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/1/?cmd=show&tvdbid=NOTFOUND", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.ServeHTTP(response, req)
	if response.Code != 500 {
		t.Fatalf("Expected 500 response code, got %d", response.Code)
	}

	// Unknown
	response = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/api/1/?cmd=show&tvdbid=0", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.ServeHTTP(response, req)
	if response.Code != 404 {
		t.Fatalf("Expected 404 response code, got %d", response.Code)
	}

	// Success
	response = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/api/1/?cmd=show&tvdbid=1", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.ServeHTTP(response, req)
	if response.Code != 200 {
		t.Fatalf("Expected 200 response code, got %d", response.Code)
	}
	Expect(response.Body.String()).Should(MatchJSON(GoldenShowResponse))
}

const ShowsGoldenResp = `{
	"data":{
		"show1":{
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
			"show_name":"show1",
			"sports":0,
			"status":"",
			"subtitles":0,
			"tvdbid":1,
			"tvrage_id":0,
			"tvrage_name":"",
			"season_list":null
		},
		"show2":{
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
			"show_name":"show2",
			"sports":0,
			"status":"",
			"subtitles":0,
			"tvdbid":2,
			"tvrage_id":0,
			"tvrage_name":"",
			"season_list":null
		}
	}
}`

func TestShows(t *testing.T) {
	dbh, eng := setupTest(t)
	db.LoadFixtures(t, dbh)
	RegisterTestingT(t)

	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/1/?cmd=shows", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.ServeHTTP(response, req)
	if response.Code != 200 {
		t.Fatalf("Expected 200 response code, got %d", response.Code)
	}
	Expect(response.Body.String()).Should(MatchJSON(ShowsGoldenResp))
}

const ShowSeasonsGolden = `{
	"data":{
		"1":{
			"airdate":"2006-01-01T00:00:00Z",
			"name":"show1episode1",
			"quality":"",
			"status":""
		},
		"2":{
			"airdate":"",
			"name":"show1episode2",
			"quality":"",
			"status":""
		}
	}
}`

func TestShowSeasons(t *testing.T) {
	dbh, eng := setupTest(t)
	db.LoadFixtures(t, dbh)
	RegisterTestingT(t)

	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/1/?cmd=show.seasons&indexerid=1&season=1", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.ServeHTTP(response, req)
	if response.Code != 200 {
		t.Fatalf("Expected 200 response code, got %d", response.Code)
	}
	spew.Dump(response.Body)
	Expect(response.Body.String()).Should(MatchJSON(ShowSeasonsGolden))
}

const EpisodeGolden = `{
	"data":{
		"airdate":"2006-01-10",
		"description":"",
		"file_size":0,
		"file_size_human":"",
		"location":"",
		"name":"show1episode1",
		"quality":"",
		"release_name":"",
		"status":""
	},
	"message":"",
	"result":"success"
}`

func TestEpisode(t *testing.T) {
	dbh, eng := setupTest(t)
	db.LoadFixtures(t, dbh)
	RegisterTestingT(t)

	//Success
	response := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/1/?cmd=episode&tvdbid=1&season=1&episode=1&full_path=1", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.ServeHTTP(response, req)
	if response.Code != 200 {
		t.Fatalf("Expected 200 response code, got %d", response.Code)
	}
	Expect(response.Body.String()).Should(MatchJSON(EpisodeGolden))

	//Invalid - missing tvdbid
	response = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/api/1/?cmd=episode&season=1&episode=1&full_path=1", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.ServeHTTP(response, req)
	if response.Code != 400 {
		t.Fatalf("Expected 400 response code, got %d", response.Code)
	}
	Expect(response.Body.String()).Should(MatchJSON(`{
		"message": "Bad Request",
		"result": "failure"
	}`))

	//Invalid - non integer tvdbid
	response = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/api/1/?cmd=episode&tvdbid=a&season=1&episode=1&full_path=1", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.ServeHTTP(response, req)
	if response.Code != 400 {
		t.Fatalf("Expected 400 response code, got %d", response.Code)
	}
	Expect(response.Body.String()).Should(MatchJSON(`{
		"message": "Bad Request",
		"result": "failure"
	}`))

	//nonexisting episode
	response = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/api/1/?cmd=episode&tvdbid=a&season=1&episode=100", nil)
	Expect(err).ToNot(HaveOccurred(), "Error creating request: %s", err)

	eng.ServeHTTP(response, req)
	spew.Dump(response.Body)
	if response.Code != 400 {
		t.Fatalf("Expected 400 response code, got %d", response.Code)
	}
	Expect(response.Body.String()).Should(MatchJSON(`{
		"message": "Bad Request",
		"result": "failure"
	}`))

}
