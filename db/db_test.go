package db

import (
	"testing"
	"time"

	"github.com/hobeone/tv2go/quality"
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

func TestEpisodeValidations(t *testing.T) {
	d := setupTest(t)

	dbep, err := d.GetEpisodeByID(1)
	if err != nil {
		t.Fatalf("Error getting episode 1")
	}

	err = dbep.BeforeSave()
	Expect(err).ToNot(HaveOccurred())

	dbep, _ = d.GetEpisodeByID(1)
	dbep.Episode = 0
	err = d.SaveEpisode(dbep)
	Expect(err).To(MatchError("Episode must be set"))

	dbep, _ = d.GetEpisodeByID(1)
	dbep.Status = types.UNKNOWN
	err = d.SaveEpisode(dbep)
	Expect(err).To(MatchError("Status must be set"))

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
func TestQualityGroup(t *testing.T) {
	d := setupTest(t)
	//flag.Set("logtostderr", "true")
	qg := quality.QualityGroup{
		Name: "HDALL_Test",
		Qualities: []quality.Quality{
			quality.UNKNOWN,
			quality.FULLHDBLURAY,
			quality.Quality(-1), // Should be dropped on save
		},
	}

	err := d.db.Save(&qg).Error
	if err != nil {
		t.Fatalf("Couldn't save test QualityGroup: %s", err)
	}
	Expect(qg.ID).ToNot(BeZero())

	dbqg := &quality.QualityGroup{}
	err = d.db.Find(dbqg, qg.ID).Error
	if err != nil {
		t.Fatalf("Couldn't find test QualityGroup: %s", err)
	}

	Expect(dbqg.Includes(quality.FULLHDBLURAY)).To(BeTrue())
	Expect(dbqg.Includes(quality.SDTV)).To(BeFalse())
}

func TestGetQualityGroups(t *testing.T) {
	d := setupTest(t)
	//flag.Set("logtostderr", "true")

	qualityGroups, err := d.GetQualityGroups()
	Expect(err).ToNot(HaveOccurred())
	Expect(qualityGroups).To(HaveLen(5))
}

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
