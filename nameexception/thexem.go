package nameexception

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/hobeone/tv2go/db"
)

//XEMProvider represents the map of XEM information for a given indexer.
type XEMProvider struct {
	name    string
	url     string
	indexer string
	client  *http.Client
	dbh     *db.Handle
}

//Name returns the name of the XEMProvider
func (msp *XEMProvider) Name() string {
	return msp.name
}

//URL returns the URL for the XEM information for this indexer
func (msp *XEMProvider) URL() string {
	return msp.url
}

// NewXEM returns a new XEMProvider with the given information set.
func NewXEM(dbh *db.Handle, indexer string) *XEMProvider {
	return &XEMProvider{
		name:    fmt.Sprintf("xem_%s_exceptions", indexer),
		url:     fmt.Sprintf("http://thexem.de/map/allNames?origin=%s&seasonNumbers=1", indexer),
		indexer: indexer,
		dbh:     dbh,
		client:  &http.Client{},
	}
}

// GetExceptions gets the url for this provider and returns the content as a
// byte slice.
func (msp *XEMProvider) GetExceptions() ([]byte, error) {
	resp, err := msp.client.Get(msp.URL())
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

type xemResponse struct {
	Data    map[string][]map[string]int64
	Message string `json:"message"`
	Result  string `json:"result"`
}

// ProcessExceptions takes the XEM response parses out the exception map and
// saves them to the database.
func (msp *XEMProvider) ProcessExceptions(input []byte) error {
	exceptions := []*db.NameException{}
	d := &xemResponse{}
	err := json.Unmarshal(input, d)
	if err != nil {
		return err
	}
	for idxid, i := range d.Data {
		indexerid, _ := strconv.ParseInt(idxid, 10, 64)
		for _, n := range i {
			for k, v := range n {
				exceptions = append(exceptions, &db.NameException{
					Source:    msp.Name(),
					Indexer:   msp.indexer,
					IndexerID: indexerid,
					Name:      k,
					Season:    v,
				})
			}
		}
	}
	err = msp.dbh.SaveNameExceptions(msp.Name(), exceptions)
	if err != nil {
		return err
	}
	return nil
}
