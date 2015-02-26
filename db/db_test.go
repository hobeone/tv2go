package db

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestGetEpisodeById(t *testing.T) {
	d := NewMemoryDBHandle(true, true)

	LoadFixtures(t, d)

	spew.Dump(d.GetEpisodeByID(1))
}
