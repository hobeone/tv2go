package types

import "fmt"

type EpisodeStatus int

// String() function will return the english name
// that we want out constant Day be recognized as
func (status EpisodeStatus) String() string {
	return statuses[status]
}

func EpisodeStatusFromString(s string) (EpisodeStatus, error) {
	for i, es := range statuses {
		if es == s {
			return EpisodeStatus(i), nil
		}
	}
	return UNKNOWN, fmt.Errorf("Unknown Episode Status: %s", s)
}

var statuses = [...]string{
	"UNKNOWN",
	"UNAIRED",
	"SNATCHED",
	"WANTED",
	"DOWNLOADED",
	"SKIPPED",
	"ARCHIVED",
	"IGNORED",
	"SNATCHED_PROPER",
	"SUBTITLED",
	"FAILED",
	"SNATCHED_BEST",
}

// Episode Status Enum
const (
	UNKNOWN         EpisodeStatus = 0 + iota // should never happen
	UNAIRED                                  // episodes that haven't aired yet
	SNATCHED                                 // qualified with quality
	WANTED                                   // episodes we don't have but want to get
	DOWNLOADED                               // qualified with quality
	SKIPPED                                  // episodes we don't want
	ARCHIVED                                 // episodes that you don't have locally (counts toward download completion stats)
	IGNORED                                  // episodes that you don't want included in your download stats
	SNATCHED_PROPER                          // qualified with quality
	SUBTITLED                                // qualified with quality
	FAILED                                   //episode downloaded or snatched we don't want
	SNATCHED_BEST                            // episode redownloaded using best quality
)

var EpisodeDefaults = []EpisodeStatus{
	WANTED,
	SKIPPED,
	IGNORED,
}
