package providers

import (
	"net/http"
	"testing"
)

func TestNzbsOrg(t *testing.T) {
	n := NewNzbsOrg("2bced8cb8f532520cbc3a367fd3962f1", SetClient(&http.Client{}))
	n.TvSearch("Archer (2009)", "5", "1")
}
