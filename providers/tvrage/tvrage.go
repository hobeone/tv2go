package tvrage

import (
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.com/hobeone/tv2go/db"
	tvr "github.com/hobeone/tvrage"
)

func SearchShowByName(name string) ([]tvr.Show, error) {
	show, err := tvr.Search(name)
	if err != nil {
		return show, err
	}
	spew.Dump(show)

	return show, nil
}

func GetShowInfo(showid int64) (*tvr.Show, error) {
	glog.Infof("Getting showid %d from TVRage.", showid)
	show, err := tvr.Get(showid)
	if err != nil {
		glog.Errorf("Error getting showid %d from TVRage: %s", showid, err)
		return &tvr.Show{}, err
	}
	return show, nil
}

func TVRageToShow(ts *tvr.Show) db.Show {
	s := db.Show{
		Name:           ts.Name,
		Genre:          strings.Join(ts.Genres, "|"),
		Classification: ts.Classification,
		Status:         ts.Status,
		StartYear:      int(ts.Started),
		IndexerID:      ts.ID,
	}
	return s
}
