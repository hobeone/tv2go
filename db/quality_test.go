package db

import (
	"testing"

	"github.com/hobeone/tv2go/quality"
	. "github.com/onsi/gomega"
)

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
