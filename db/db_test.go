package db

import (
	"testing"
	"time"

	"github.com/hobeone/tv2go/types"
	. "github.com/onsi/gomega"
)

func setupTest(t *testing.T) *Handle {
	d := NewMemoryDBHandle(false, true)
	LoadFixtures(t, d)
	RegisterTestingT(t)
	return d
}

func TestGetEpisodeById(t *testing.T) {
	d := setupTest(t)
	ep, err := d.GetEpisodeByID(1)
	Expect(err).ToNot(HaveOccurred())
	Expect(ep.ID).To(BeNumerically("==", 1))
}

func TestShowValidations(t *testing.T) {
	d := setupTest(t)

	// Empty name
	dbshow := Show{
		IndexerID: 1,
		Location:  "testlocation",
	}

	err := d.SaveShow(&dbshow)
	Expect(err).To(MatchError("Name can not be empty"))

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

func TestEpisodeValidations(t *testing.T) {
	d := setupTest(t)

	dbep, err := d.GetEpisodeByID(1)
	if err != nil {
		t.Fatalf("Error getting episode 1")
	}

	err = dbep.BeforeSave()
	Expect(err).ToNot(HaveOccurred())

	dbep.Name = ""

	err = d.SaveEpisode(dbep)
	Expect(err).To(MatchError("Name can not be empty"))

	dbep, _ = d.GetEpisodeByID(1)
	dbep.Episode = 0
	err = d.SaveEpisode(dbep)
	Expect(err).To(MatchError("Episode must be set"))

	dbep, _ = d.GetEpisodeByID(1)
	dbep.Status = types.UNKNOWN
	err = d.SaveEpisode(dbep)
	Expect(err).To(MatchError("Status must be set"))

	dbep, _ = d.GetEpisodeByID(1)
	dbep.Quality = 0
	err = d.SaveEpisode(dbep)
	Expect(err).To(MatchError("Quality must be set"))
}

func TestAddShow(t *testing.T) {
	d := setupTest(t)
	dbshow := Show{
		Name:      "testshow1",
		Location:  "testlocation",
		IndexerID: 1,
	}
	err := d.AddShow(&dbshow)
	Expect(err).ToNot(HaveOccurred())
}

func TestGetShowByName(t *testing.T) {
	d := setupTest(t)
	dbshow, err := d.GetShowByName("show1")
	Expect(err).ToNot(HaveOccurred())
	Expect(dbshow.Name).To(Equal("show1"))

	dbshow, err = d.GetShowByName("")
	Expect(err).To(MatchError("record not found"))
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

	showtime, err := d.NextAirdateForShow(dbshow)
	Expect(err).ToNot(HaveOccurred())
	Expect(showtime).To(Equal(futureDate))
}
