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
	RegisterTestingT(t)
	tests := map[string][]string{
		"standard_repeat": []string{
			"Show.Name.S01E02.S01E03.Source.Quality.Etc-Group",
			"Show Name - S01E02 - S01E03 - S01E04 - Ep Name",
		},
		"fov_repeat": []string{
			"Show.Name.1x02.1x03.Source.Quality.Etc-Group",
			"Show Name - 1x02 - 1x03 - 1x04 - Ep Name",
		},
		"standard": []string{
			"Show.Name.S01E02.Source.Quality.Etc-Group",
			"Show Name - S01E02 - My Ep Name",
			"Show.Name.S01.E03.My.Ep.Name",
			"Show.Name.S01E02E03.Source.Quality.Etc-Group",
			"Show Name - S01E02-03 - My Ep Name",
			"Show.Name.S01.E02.E03",
		},
		"fov": []string{
			"Show_Name.1x02.Source_Quality_Etc-Group",
			"Show Name - 1x02 - My Ep Name",
			"Show_Name.1x02x03x04.Source_Quality_Etc-Group",
			"Show Name - 1x02-03-04 - My Ep Name",
		},
		"scene_date_format": []string{
			"Show.Name.2010.11.23.Source.Quality.Etc-Group",
			"Show Name - 2010-11-23 - Ep Name",
		},
		/*
			"scene_sports_format": []string{
				"Show.Name.100.Event.2010.11.23.Source.Quality.Etc-Group",
				"Show.Name.2010.11.23.Source.Quality.Etc-Group",
				"Show Name - 2010-11-23 - Ep Name",
			},
		*/
		"stupid": []string{
			"tpz-abc102",
		},
		"verbose":     []string{"Show Name Season 1 Episode 2 Ep Name"},
		"season_only": []string{"Show.Name.S01.Source.Quality.Etc-Group"},
		/*
			"no_season_multi_ep": []string{
				"Show.Name.E02-03",
				"Show.Name.E02.2010",
			},
		*/
		"no_season_general": []string{
			"Show.Name.E23.Test",
			"Show.Name.Part.3.Source.Quality.Etc-Group",
			"Show.Name.Part.1.and.Part.2.Blah-Group",
		},
		/*
			"no_season": []string{
				"Show Name - 01 - Ep Name",
				"01 - Ep Name",
			},
		*/
		"bare": []string{"Show.Name.102.Source.Quality.Etc-Group"},
	}
	for _, regex := range NameRegexes {
		for _, str := range tests[regex.Name] {
			_, matched := regexNamedMatch(&regex.Regex, str)
			Expect(matched).Should(BeTrue(), "Expected to match %s with the %s regex", str, regex.Name)
		}
	}
}
