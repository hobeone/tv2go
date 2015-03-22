package db

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestNameException(t *testing.T) {
	//flag.Set("logtostderr", "true")
	d := setupTest(t)
	se := &NameException{
		IndexerID: 123,
		Indexer:   "tvdb",
		Source:    "tvdb source one",
		Name:      "Testing 1 2 3",
	}
	err := d.db.Save(se).Error

	Expect(err).ToNot(HaveOccurred())
}
