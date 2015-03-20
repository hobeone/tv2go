package nameexception

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/hobeone/tv2go/db"
)

// MidgetSpyProvider provides the functionality to parse and update scene
// exceptions from Midget Spy's files.
type MidgetSpyProvider struct {
	name    string
	url     string
	indexer string
	client  *http.Client
	dbh     *db.Handle
}

// Name implements part of the Provider interface.
func (msp *MidgetSpyProvider) Name() string {
	return msp.name
}

// URL implements part of the Provider interface.
func (msp *MidgetSpyProvider) URL() string {
	return msp.url
}

// NewMidgetSpyTvdb returns a new MidgetSpyProvider configured to look at the
// TVDB scene exceptions.
func NewMidgetSpyTvdb(dbh *db.Handle) *MidgetSpyProvider {
	return &MidgetSpyProvider{
		name:    "midgetspy_tvdb",
		indexer: "tvdb",
		url:     "https://midgetspy.github.io/sb_tvdb_scene_exceptions/exceptions.txt",
		dbh:     dbh,
		client:  &http.Client{},
	}
}

// GetExceptions gets the exception url
func (msp *MidgetSpyProvider) GetExceptions() ([]byte, error) {
	resp, err := msp.client.Get(msp.URL())
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

//ProcessExceptions takes the returned data from the midgetspy exceptions and
//adds them to the DB.
//
//Format:
//110381: 'Archer',
//80552: 'Kitchen Nightmares (US)', 'Kitchen Nightmares US',
func (msp *MidgetSpyProvider) ProcessExceptions(input []byte) error {
	scanner := bufio.NewScanner(bytes.NewReader(input))
	excepts := []*db.NameException{}
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
				Indexer:   msp.indexer,
				IndexerID: indexerid,
				Name:      name,
				Source:    msp.Name(),
			}
			excepts = append(excepts, se)
		}
	}
	if err := scanner.Err(); err != nil {
		glog.Errorf("Error reading %s: %s", msp.URL(), err)
		return err
	}

	glog.Infof("Got %d results from provider %s", len(excepts), msp.Name())
	err := msp.dbh.SaveNameExceptions(msp.Name(), excepts)
	if err != nil {
		return err
	}
	return nil
}
