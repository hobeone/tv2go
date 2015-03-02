package providers

import (
	"fmt"
	"time"
)

// ProviderRegistry provides an easy way to map providers to string names
type ProviderRegistry map[string]Provider

// ProviderResult describes the information that Providers will return from searches
type ProviderResult struct {
	Type         string     `json:"type"`
	Age          *time.Time `json:"age,omitempty"` //hours
	Name         string     `json:"name"`
	Size         int64      `json:"size"`
	Quality      string     `json:"quality"`
	ProviderName string     `json:"indexer"`
	URL          string     `json:"url"`
}

// Provider defines the interface a tv2go provider must implement
type Provider interface {
	TvSearch(string, int64, int64) ([]ProviderResult, error)
	//need better name
	//Get file contents, leave it to something else to save it to disk
	GetURL(URL string) (string, []byte, error)
	// Return what kind of providers this is for: NZB/Torrent
	Type() ProviderType
}

// ProviderType is for the constants below
type ProviderType int

// String() function will return the english name
// that we want out constant Day be recognized as
func (t ProviderType) String() string {
	return types[t]
}

//ProviderTypeFromString converts a string name to a ProviderType
func ProviderTypeFromString(s string) (ProviderType, error) {
	for i, pt := range types {
		if pt == s {
			return ProviderType(i), nil
		}
	}
	return UNKNOWN, fmt.Errorf("Unknown Provider Type: %s", s)
}

// Different kinds of providers
const (
	NZB ProviderType = 0 + iota
	TORRENT
	UNKNOWN
)

var types = [...]string{
	"NZB",
	"TORRENT",
	"UNKNOWN",
}
