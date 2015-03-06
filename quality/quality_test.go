package quality

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestQualityFromName(t *testing.T) {
	RegisterTestingT(t)
	q := QualityFromName("testing.dvdrip.avi", false)
	Expect(q).To(Equal(UNKNOWN))
	Expect(QualityFromName("testing.1080p.mkv", true)).To(Equal(FULLHDTV))
	Expect(QualityFromName("testing.1080p BluRay.mkv", true)).To(Equal(FULLHDBLURAY))
	Expect(QualityFromName("Archer (2009) - S04E12 - Sea Tunt (1) - HD TV.mkv", false)).To(Equal(UNKNOWN))
	Expect(QualityFromName("[HorribleSubs] Yowamushi Pedal - Grande Road - 20 [720p].mkv.torrent", true)).To(Equal(HDTV))
}

func TestQualityGroups(t *testing.T) {
	RegisterTestingT(t)

	qg := HDALL
	Expect(qg.AcceptableQuality(HDTV)).To(BeTrue())
	Expect(qg.AcceptableQuality(SDTV)).To(BeFalse())
}
