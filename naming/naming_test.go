package naming

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	. "github.com/onsi/gomega"
)

func TestMediaFile(t *testing.T) {
	RegisterTestingT(t)

	tests := map[string]bool{
		"sAmPle123_test.mkv":     false,
		"._res_test.mkv":         false,
		"Show_S01E02_Extras.mkv": false,
		"Show_S01E02.txt":        false,
		"Show_S01E02.mkv":        true,
	}
	for str, testval := range tests {
		Expect(IsMediaFile(str)).To(Equal(testval), "Expected %s to return %v from IsMediaFile", str, testval)
	}
}

func TestNameParser(t *testing.T) {
	RegisterTestingT(t)

	np := NewNameParser("foo", StandardRegexes)
	names := []string{
		"TV/Archer (2009)/Season 04/Archer (2009) - S04E12 - Sea Tunt (1) - HD TV.mkv",
		"TV/Archer (2009)/Season 04/Archer (2009) - S04E13 - Sea Tunt (2) - HD TV.mkv",
		"TV/Archer (2009)/Season 05/Archer (2009) - S05E05 - Archer Vice Southbound and Down.mkv",
		"TV/Archer (2009)/Season 05/Archer (2009) - S05E04 - Archer Vice House Call.mkv",
	}
	for _, f := range names {
		r := np.Parse(f)
		Expect(r.SeriesName).To(Equal("Archer (2009)"))
		spew.Dump(r)
	}
}

func TestAnimeRegex(t *testing.T) {
	RegisterTestingT(t)
	for _, regex := range AnimeRegex {
		for _, ts := range regex.TestStrings {
			matches, _ := regexNamedMatch(&regex.Regex, ts.String)
			Expect(matches).To(Equal(ts.MatchGroups), "Didn't get expected match groups for '%s'", ts.String)
		}

	}

}

func TestRegex(t *testing.T) {
	RegisterTestingT(t)
	for _, regex := range StandardRegexes {
		for _, ts := range regex.TestStrings {
			matches, _ := regexNamedMatch(&regex.Regex, ts.String)
			Expect(matches).To(Equal(ts.MatchGroups), "Didn't get expected match groups for regex '%s' on '%s'", regex.Name, ts.String)
		}
	}
}
