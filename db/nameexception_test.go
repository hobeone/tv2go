package db

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestNameExceptionValidation(t *testing.T) {
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

func TestGetShowFromNameException(t *testing.T) {
	d := setupTest(t)

	dbshow := &Show{
		Name:      "Log Horizon",
		Indexer:   "tvdb",
		IndexerID: 272128,
	}
	err := d.AddShow(dbshow)
	if err != nil {
		t.Fatalf("Error saving show: %s", err)
	}

	ne := []*NameException{
		&NameException{
			Source:    "xem",
			Indexer:   "tvdb",
			IndexerID: 272128,
			Name:      "Log Horizon 2nd Season",
			Season:    2,
		},
		&NameException{
			Source:    "xem",
			Indexer:   "rage",
			IndexerID: 38112,
			Name:      "Log Horizon 2nd Season",
			Season:    2,
		},
	}

	d.SaveNameExceptions("xem", ne)

	_, _, err = d.GetShowFromNameException("test")
	Expect(err).To(HaveOccurred())

	show, season, err := d.GetShowFromNameException("Log Horizon 2nd Season")
	Expect(season).To(Equal(int64(2)))
	Expect(show.Name).To(Equal("Log Horizon"))
	Expect(show.Indexer).To(Equal("tvdb"))
}
