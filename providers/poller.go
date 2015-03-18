package providers

import (
	"time"

	"github.com/golang/glog"
	"github.com/hobeone/tv2go/db"
)

type ProviderPoller struct {
	Interval     time.Duration
	ResponseChan chan (ProviderResult)
	Provider     Provider
	LastPoll     time.Time
	DBH          *db.Handle
	after        func(time.Duration) <-chan time.Time // Allow for mocking out in test.
}

func NewProviderPoller(p Provider, interval time.Duration, dbh *db.Handle, respChan chan (ProviderResult)) *ProviderPoller {
	return &ProviderPoller{
		Interval:     interval,
		ResponseChan: respChan,
		Provider:     p,
		DBH:          dbh,
		after:        time.After,
	}
}

// Poll is designed to be run in a goroutine and polls the Provider for new
// items returning responses on p.ResponseChannel()
func (p *ProviderPoller) Poll() {
	p.LastPoll = p.DBH.GetLastPollTime(p.Provider.Name())

	timeSinceLastPoll := time.Since(p.LastPoll)
	toSleep := time.Duration(0)
	if timeSinceLastPoll < p.Interval {
		toSleep = p.Interval - timeSinceLastPoll
		glog.Infof("%s last poll (%s) was to soon, sleeping %s", p.Provider.Name(), p.LastPoll, toSleep.String())
	}
	for {
		select {
		case <-time.After(toSleep):
			glog.Infof("%s poller waking up", p.Provider.Name())
			toSleep = p.Interval
			resp, err := p.Provider.GetNewItems()
			if err != nil {
				glog.Errorf("error polling provider %s: %s", p.Provider.Name(), err)
				break
			}
			err = p.DBH.SetLastPollTime(p.Provider.Name())
			if err != nil {
				glog.Errorf("error saving last poll time to db: %s", err)
			}

			glog.Infof("Got %d results from provider %s", len(resp), p.Provider.Name())
			for _, r := range resp {
				p.ResponseChan <- r
			}
			glog.Infof("%s done processing, sleeping %s", p.Provider.Name(), toSleep.String())
		}
	}
}
