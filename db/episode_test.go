package db

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestGetEpisodeById(t *testing.T) {
	d := setupTest(t)
	ep, err := d.GetEpisodeByID(1)
	Expect(err).ToNot(HaveOccurred())
	Expect(ep.ID).To(BeNumerically("==", 1))
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
}
