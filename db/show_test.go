package db

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestShowValidations(t *testing.T) {
	d := setupTest(t)

	// Empty name
	dbshow := Show{
		IndexerID: 1,
		Location:  "testlocation",
	}

	err := d.SaveShow(&dbshow)
	Expect(err).To(MatchError("Show Name can not be empty"))

	// Empty indexerid
	dbshow = Show{
		Name:     "testshow",
		Location: "testlocation",
	}

	err = d.SaveShow(&dbshow)
	Expect(err).To(MatchError("IndexerID can not be unset"))

}

func TestShowDuplicateName(t *testing.T) {
	d := setupTest(t)
	dbshow, err := d.GetShowByID(1)
	if err != nil {
		t.Fatalf("Error getting show 1")
	}
	newshow := *dbshow
	newshow.ID = 0
	err = d.SaveShow(&newshow)
	Expect(err).To(MatchError("UNIQUE constraint failed: show.name"))
}

func TestAddShow(t *testing.T) {
	d := setupTest(t)
	dbshow := Show{
		Name:      "testshow1",
		Location:  "testlocation",
		IndexerID: 1,
		Indexer:   "tvdb",
	}
	err := d.AddShow(&dbshow)
	Expect(err).ToNot(HaveOccurred())
}
func TestGetShowByName(t *testing.T) {
	d := setupTest(t)
	dbshow, err := d.GetShowByName("show1")
	Expect(err).ToNot(HaveOccurred())
	Expect(dbshow.Name).To(Equal("show1"))
	Expect(dbshow.QualityGroup.Name).To(Equal("HDALL"))

	dbshow, err = d.GetShowByName("")
	Expect(err).To(MatchError("record not found"))

	dbshow, err = d.GetShowByName("SHOW1")
	Expect(err).To(MatchError("record not found"))

	dbshow, err = d.GetShowByNameIgnoreCase("SHOW1")
	Expect(err).ToNot(HaveOccurred())
	Expect(dbshow.Name).To(Equal("show1"))

}

func TestNextAirdateForShow(t *testing.T) {
	d := setupTest(t)
	dbshow, err := d.GetShowByName("show1")
	Expect(err).ToNot(HaveOccurred())

	eps, err := d.GetShowEpisodes(dbshow)
	Expect(err).ToNot(HaveOccurred())

	futureDate := time.Now().UTC().Add(time.Second * 60)
	eps[0].AirDate = futureDate
	err = d.SaveEpisode(&eps[0])
	Expect(err).ToNot(HaveOccurred())

	showtime := d.NextAirdateForShow(dbshow)
	Expect(showtime).To(Equal(&futureDate))
}
