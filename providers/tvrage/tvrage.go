package tvrage

import (
	"github.com/davecgh/go-spew/spew"
	tvr "github.com/drbig/tvrage"
)

func SearchShowByName(name string) ([]tvr.Show, error) {
	show, err := tvr.Search(name)
	if err != nil {
		return show, err
	}
	spew.Dump(show)

	return show, nil
}

func GetShowInfo(showid int) (tvr.Episodes, error) {
	return tvr.EpisodeList(showid)
}
