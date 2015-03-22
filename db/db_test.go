package db

import (
	"testing"

	. "github.com/onsi/gomega"
)

func setupTest(t *testing.T) *Handle {
	d := NewMemoryDBHandle(false, true)
	LoadFixtures(t, d)
	RegisterTestingT(t)
	return d
}
