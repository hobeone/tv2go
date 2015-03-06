package types

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestStatusEnum(t *testing.T) {
	RegisterTestingT(t)
	d := EpisodeStatus(1)
	Expect(d).To(Equal(UNAIRED))
}
