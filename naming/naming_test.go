package naming

import (
	"testing"

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

	np := NewNameParser("foo")
	names := []string{
		"TV/Archer (2009)/Season 04/Archer (2009) - S04E12 - Sea Tunt (1) - HD TV.mkv",
		"TV/Archer (2009)/Season 04/Archer (2009) - S04E13 - Sea Tunt (2) - HD TV.mkv",
		"TV/Archer (2009)/Season 05/Archer (2009) - S05E05 - Archer Vice Southbound and Down.mkv",
		"TV/Archer (2009)/Season 05/Archer (2009) - S05E04 - Archer Vice House Call.mkv",
	}
	for _, f := range names {
		r := np.Parse(f)
		Expect(r.SeriesName).To(Equal("Archer (2009)"))
	}
}

func TestRegex(t *testing.T) {
	tests := []string{
		"Show.Name.S01E02.Source.Quality.Etc-Group",
		"Show Name - S01E02 - My Ep Name",
		"Show.Name.S01.E03.My.Ep.Name",
		"Show.Name.S01E02E03.Source.Quality.Etc-Group",
		"Show Name - S01E02-03 - My Ep Name",
		"Show.Name.S01.E02.E03",
	}

	for _, str := range tests {
		_, matched := regexNamedMatch(NameRegexes[0], str)
		Expect(matched).Should(BeTrue())
	}
}
