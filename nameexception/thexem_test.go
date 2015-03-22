package nameexception

import (
	"io/ioutil"
	"testing"

	"github.com/golang/glog"
	"github.com/hobeone/tv2go/test_helpers"
)

func TestXEM(t *testing.T) {
	d := setupTest(t)
	content, err := ioutil.ReadFile("testdata/xem_tvdb_seasons.json")
	if err != nil {
		glog.Fatalf("Error reading test feed: %s", err.Error())
	}

	srv, client := test_helpers.ServeFile(200, string(content), "application/json")
	defer srv.Close()

	xem := NewXEM(d, "tvdb")
	xem.client = client
	xem.url = "http://test"

	res, err := xem.GetExceptions()
	if err != nil {
		t.Fatalf("Error getting expected test exceptions: %s", err)
	}
	err = xem.ProcessExceptions(res)
	if err != nil {
		t.Fatalf("Error processing xem exceptions: %s", err)
	}
}

func TestXEMRage(t *testing.T) {
	d := setupTest(t)
	content, err := ioutil.ReadFile("testdata/xem_rage_seasons.json")
	if err != nil {
		glog.Fatalf("Error reading test feed: %s", err.Error())
	}

	srv, client := test_helpers.ServeFile(200, string(content), "application/json")
	defer srv.Close()

	xem := NewXEM(d, "rage")
	xem.client = client
	xem.url = "http://test"

	res, err := xem.GetExceptions()
	if err != nil {
		t.Fatalf("Error getting expected test exceptions: %s", err)
	}
	err = xem.ProcessExceptions(res)
	if err != nil {
		t.Fatalf("Error processing xem exceptions: %s", err)
	}
}
