package nameexception

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/golang/glog"
	"github.com/hobeone/tv2go/test_helpers"
)

func TestMidgetSpy(t *testing.T) {
	d := setupTest(t)
	//flag.Set("logtostderr", "true")
	content, err := ioutil.ReadFile("testdata/tvdb_exceptions.txt")
	if err != nil {
		glog.Fatalf("Error reading test feed: %s", err.Error())
	}

	srv, client := test_helpers.ServeFile(200, string(content), "text/plain; charset=utf-8")
	defer srv.Close()

	msp := NewMidgetSpyTvdb(d)

	poller := NewProviderPoller(msp, time.Minute, d)

	OverrideAfter(poller)
	msp.client = client
	msp.url = "http://test"
	c := make(chan int)
	w := func() { time.Sleep(1); c <- 1 }
	go w()
	poller.Poll(c)
}
