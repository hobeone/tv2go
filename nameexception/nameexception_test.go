package nameexception

import (
	"flag"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/golang/glog"
	"github.com/hobeone/tv2go/db"
	"github.com/hobeone/tv2go/test_helpers"
)

func setupTest(t *testing.T) *db.Handle {
	flag.Set("logtostderr", "true")

	dbh := db.NewMemoryDBHandle(true, true)

	return dbh
}
func OverrideAfter(fw *Provider) {
	fw.after = func(d time.Duration) <-chan time.Time {
		glog.Infof("Zero delay after call for testing.")
		return time.After(time.Duration(0))
	}
}
func TestPoll(t *testing.T) {
	d := setupTest(t)
	content, err := ioutil.ReadFile("testdata/tvdb_exceptions.txt")
	if err != nil {
		glog.Fatalf("Error reading test feed: %s", err.Error())
	}

	srv, client := test_helpers.ServeFile(200, string(content), "text/plain; charset=utf-8")
	defer srv.Close()

	n := NewProvider(
		"tvdb",
		"tvdb",
		"http://test/url",
		time.Hour*24,
		d,
	)
	OverrideAfter(n)
	n.Client = client
	c := make(chan int)
	w := func() { time.Sleep(1); c <- 1 }
	go w()
	n.Poll(c)
}

func TestPollOnce(t *testing.T) {
	d := setupTest(t)
	content, err := ioutil.ReadFile("testdata/tvdb_exceptions.txt")
	if err != nil {
		glog.Fatalf("Error reading test feed: %s", err.Error())
	}

	srv, client := test_helpers.ServeFile(200, string(content), "text/plain; charset=utf-8")
	defer srv.Close()

	n := NewProvider(
		"tvdb source one",
		"tvdb",
		"http://test/url",
		time.Hour*24,
		d,
	)
	n.Client = client
	res, err := n.PollOnce()
	if err != nil {
		t.Fatalf("Error polling feed: %s", err)
	}
	for _, e := range res {
		if e.Name == "" {
			t.Fatal("exception had an empty name")
		}
		if strings.Contains(e.Name, `\`) {
			t.Fatalf("Name %s had unexpected backslash", e.Name)
		}
	}
	if len(res) != 15 {
		t.Fatalf("Expected 15 items in list, got %d", len(res))
	}

	err = n.DBH.SaveNameExceptions(n.Name, res)
}
