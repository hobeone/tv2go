package nameexception

import (
	"bufio"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/hobeone/tv2go/db"
)

//Provider represents a Name Exception provider.  These provide alternate names
//for a given indexer id.  These are used when matching show names to a
//canonical name.
type Provider struct {
	Name         string
	URL          string
	Client       *http.Client
	Indexer      string
	PollInterval time.Duration
	DBH          *db.Handle
	after        func(d time.Duration) <-chan time.Time // Allow for mocking out in test.
}

// NewProvider returns a new Provider with the given arguments set.
func NewProvider(name string, indexer string, url string, interval time.Duration, d *db.Handle) *Provider {
	return &Provider{
		Name:         name,
		URL:          url,
		Client:       &http.Client{},
		Indexer:      indexer,
		PollInterval: interval,
		DBH:          d,
		after:        time.After,
	}
}

func (p *Provider) afterWithJitter(d time.Duration) <-chan time.Time {
	s := d + time.Duration(rand.Int63n(60))*time.Second
	glog.Infof("Waiting %v until next poll of %s", s, p.URL)
	return p.after(s)
}

// Poll is designed to be run in a goroutine and wraps the polling and sleeping
// behavior for each exception list.
func (p *Provider) Poll(exitChan chan int) {
	lastrun := p.DBH.GetNameExceptionHistory(p.Name)
	toSleep := time.Since(lastrun)
	if toSleep > p.PollInterval {
		toSleep = time.Duration(0)
	}

	for {
		select {
		case <-exitChan:
			glog.Infof("%s name exception provider got exit signal.", p.Name)
			p.DBH.SetNameExceptionHistory(p.Name)
			return
		case <-p.afterWithJitter(toSleep):
			toSleep = p.PollInterval
			nameExceptions, err := p.PollOnce()
			if err != nil {
				break
			}
			err = p.DBH.SaveNameExceptions(p.Name, nameExceptions)
		}
	}
}

// PollOnce will retrieve the Exception list url, parse the contents and return
// db.NameExceptions ready to be saved to the DB.
func (p *Provider) PollOnce() ([]*db.NameException, error) {
	resp, err := p.Client.Get(p.URL)
	excepts := []*db.NameException{}
	if err != nil {
		return excepts, fmt.Errorf("Error getting url %s: %s", p.URL, err)
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			glog.Infof("Unknown line format for: '%s'", line)
			continue
		}
		indexerid, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil {
			glog.Errorf("Couldn't parse indexerid %s: %s", parts[0], err)
			continue
		}
		names := strings.Split(parts[1], ",")
		for _, name := range names {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			name = strings.Trim(name, "'")
			name = strings.Replace(name, `\`, "", -1)
			se := &db.NameException{
				Indexer:   p.Indexer,
				IndexerID: indexerid,
				Name:      name,
				Source:    p.Name,
			}
			excepts = append(excepts, se)
		}
	}
	if err := scanner.Err(); err != nil {
		glog.Errorf("Error reading %s: %s", p.URL, err)
		return excepts, err
	}

	return excepts, nil
}
