package nameexception

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/hobeone/tv2go/db"
)

type XEMProvider struct {
	name    string
	url     string
	indexer string
	client  *http.Client
	dbh     *db.Handle
}

func (msp *XEMProvider) Name() string {
	return msp.name
}

func (msp *XEMProvider) URL() string {
	return msp.url
}

func NewXem(dbh *db.Handle, indexer string) *XEMProvider {
	return &XEMProvider{
		name:    fmt.Sprintf("xem_%s_exceptions", indexer),
		url:     fmt.Sprintf("http://thexem.de/map/allNames?origin=%s&seasonNumbers=1", indexer),
		indexer: indexer,
		dbh:     dbh,
		client:  &http.Client{},
	}
}

func (msp *XEMProvider) GetExceptions() ([]byte, error) {
	resp, err := msp.client.Get(msp.URL())
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

type XEMResponse struct {
	Data    map[string][]map[string]int64
	Message string `json:"message"`
	Result  string `json:"result"`
}

func (msp *XEMProvider) ProcessExceptions(input []byte) error {
	exceptions := []*db.XEMException{}
	d := &XEMResponse{}
	err := json.Unmarshal(input, d)
	if err != nil {
		return err
	}
	for idxid, i := range d.Data {
		indexerid, _ := strconv.ParseInt(idxid, 10, 64)
		for _, n := range i {
			for k, v := range n {
				exceptions = append(exceptions, &db.XEMException{
					Indexer:   msp.indexer,
					IndexerID: indexerid,
					Name:      k,
					Season:    v,
				})
			}
		}
	}
	err = msp.dbh.SaveXEMExceptions(msp.indexer, exceptions)
	if err != nil {
		return err
	}
	return nil
}
