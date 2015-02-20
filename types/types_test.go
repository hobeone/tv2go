package types

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestQualityEnum(t *testing.T) {
	RegisterTestingT(t)
	d := Quality(2)
	Expect(d).To(Equal(NONE))

	q, err := QualityFromString("SDTV")
	Expect(err).ToNot(HaveOccurred(), "Error getting quality: %s", err)
	Expect(q).To(Equal(SDTV))
}

func TestStatusEnum(t *testing.T) {
	RegisterTestingT(t)
	d := EpisodeStatus(1)
	Expect(d).To(Equal(UNAIRED))
}
