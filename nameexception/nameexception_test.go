package nameexception

import (
	"testing"
	"time"

	"github.com/golang/glog"
	"github.com/hobeone/tv2go/db"
)

func setupTest(t *testing.T) *db.Handle {
	dbh := db.NewMemoryDBHandle(true, true)

	return dbh
}

func OverrideAfter(fw *ProviderPoller) {
	fw.after = func(d time.Duration) <-chan time.Time {
		glog.Infof("Zero delay after call for testing.")
		return time.After(time.Duration(0))
	}
}
