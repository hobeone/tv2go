package indexers

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/hobeone/tv2go/db"
)

type Indexer interface {
	Search(string) ([]db.Show, error)
	GetShow(string) (*db.Show, error)
	UpdateShow(*db.Show) error
}

type IndexerRegistry map[string]Indexer

type TestIndexer struct{}

func (t *TestIndexer) UpdateShow(d *db.Show) error {
	spew.Dump(t)
	return nil
}
