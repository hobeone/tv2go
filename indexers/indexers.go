package indexers

import "github.com/hobeone/tv2go/db"

// Indexer defines the interface for Indexing clients
type Indexer interface {
	Search(string) ([]db.Show, error)
	GetShow(string) (*db.Show, error)
	UpdateShow(*db.Show) error
	Name() string
}

// IndexerRegistry provides a convenient way of keeping a list of all known
// indexers.
type IndexerRegistry map[string]Indexer
