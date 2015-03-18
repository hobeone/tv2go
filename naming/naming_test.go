package naming

import (
	"strconv"
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

func TestNameParse(t *testing.T) {
	RegisterTestingT(t)

	np := NewNameParser("foo", StandardRegexes)
	names := [][]string{
		[]string{"Worlds.Toughest.Jobs.S01E04.Cattle.Ranching.HDTV.x264-C4TV", "Worlds.Toughest.Jobs", "1", "4"},
		[]string{"The.Flash.2014.S01E15.HDTV.x264-LOL", "The.Flash.2014", "1", "15"},
	}
	for i, f := range names {
		r := np.ParseFile(f[0])
		Expect(r.SeriesName).To(Equal(names[i][1]))
		Expect(strconv.FormatInt(r.SeasonNumber, 10)).To(Equal(names[i][2]))
		Expect(strconv.FormatInt(r.EpisodeNumbers[0], 10)).To(Equal(names[i][3]))
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
		r := np.ParseFile(f)
		Expect(r.SeriesName).To(Equal("Archer (2009)"))
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

func TestFullSanitizeName(t *testing.T) {
	RegisterTestingT(t)
	Expect(FullSanitizeSceneName(`Marvel's.Agents.of.S.H.I.E.L.D.`)).To(Equal("marvels agents of s h i e l d"))
	Expect(FullSanitizeSceneName(`Adventure.Time`)).To(Equal("adventure time"))
}
