package tvdb

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestMe(t *testing.T) {

	show, eps, err := GetShowById(110381)

	spew.Dump(err)

	spew.Dump(show)

	spew.Dump(eps)
}
