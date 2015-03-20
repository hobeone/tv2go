package nameexception

import (
	"time"

	"github.com/golang/glog"
	"github.com/hobeone/tv2go/db"
)

// ProviderPoller provides the generic functionality to poll a given Name
// Exception Provider every so often.
type ProviderPoller struct {
	Interval time.Duration
	Provider Provider
	LastPoll time.Time
	DBH      *db.Handle
	after    func(time.Duration) <-chan time.Time // Allow for mocking out in test.
}

//NewProviderPoller returns a new configured ProviderPoller
func NewProviderPoller(p Provider, interval time.Duration, dbh *db.Handle) *ProviderPoller {
	return &ProviderPoller{
		Interval: interval,
		Provider: p,
		DBH:      dbh,
		after:    time.After,
	}
}

// Poll is designed to be run in a goroutine.  It will get new information from
// the given provider every Interval and update the last polled time for that
// provider in the db.
func (p *ProviderPoller) Poll(exitChan chan int) {
	p.LastPoll = p.DBH.GetLastPollTime(p.Provider.Name())

	timeSinceLastPoll := time.Since(p.LastPoll)
	toSleep := time.Duration(0)
	if timeSinceLastPoll < p.Interval {
		toSleep = p.Interval - timeSinceLastPoll
		glog.Infof("%s last poll (%s) was to soon, sleeping %s", p.Provider.Name(), p.LastPoll, toSleep.String())
	}
	for {
		select {
		case <-exitChan:
			glog.Infof("%s name exception provider got exit signal.", p.Provider.Name())
			return

		case <-p.after(toSleep):
			glog.Infof("%s poller waking up", p.Provider.Name())
			toSleep = p.Interval
			b, err := p.Provider.GetExceptions()
			if err != nil {
				glog.Errorf("error polling provider %s: %s", p.Provider.Name(), err)
				break
			}
			glog.Infof("Got %d bytes from %s", len(b), p.Provider.URL())
			err = p.DBH.SetLastPollTime(p.Provider.Name())
			if err != nil {
				glog.Errorf("error saving last poll time to db: %s", err)
			}
			p.Provider.ProcessExceptions(b)

			glog.Infof("%s done processing, sleeping %s", p.Provider.Name(), toSleep.String())
		}
	}
}
