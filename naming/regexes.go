package naming

import "github.com/kyoh86/go-pcre"

// AllRegexes is exactly what it sounds like.
var AllRegexes = append(StandardRegexes, AnimeRegex...)

// NameRegexes is a list of Regular Expressions to try in order when trying to
// extract information from a filename.
var StandardRegexes = []NameRegex{
	// Lifted from SickRage
	{
		Name: "standard_repeat",
		TestStrings: []TestString{
			{
				String:      "Show.Name.S01E02.S01E03.Source.Quality.Etc-Group",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"extra_ep_num":  "03",
					"release_group": "Group",
					"ep_num":        "02",
					"season_num":    "01",
					"extra_info":    "Source.Quality.Etc",
					"series_name":   "Show.Name",
				},
			},
			{
				String:      "Show Name - S01E02 - S01E03 - S01E04 - Ep Name",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"series_name":  "Show Name",
					"extra_ep_num": "04",
					"extra_info":   "Ep Name",
					"season_num":   "01",
					"ep_num":       "02",
				},
			},
		},
		Regex: pcre.MustCompile(`(?i)^(?P<series_name>.+?)[. _-]+`+ //  Show_Name and separator
			`s(?P<season_num>\d+)[. _-]*`+ //  S01 and optional separator
			`e(?P<ep_num>\d+)`+ //  E02 and separator
			`([. _-]+s(?P=season_num)[. _-]*`+ //  S01 and optional separator
			`e(?P<extra_ep_num>\d+))+`+ //  E03/etc and separator
			`[. _-]*((?P<extra_info>.+?)`+ //  Source_Quality_Etc-
			`((?<![. _-])(?<!WEB)`+ //  Make sure this is really the release group
			`-(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$`, pcre.CASELESS), //  Group
	},
	{
		Name: "fov_repeat",
		TestStrings: []TestString{
			{
				String:      "Show.Name.1x02.1x03.Source.Quality.Etc-Group",
				ShouldMatch: true,
				MatchGroups: map[string]string{

					"series_name":   "Show.Name",
					"season_num":    "1",
					"extra_ep_num":  "03",
					"ep_num":        "02",
					"extra_info":    "Source.Quality.Etc",
					"release_group": "Group",
				},
			},
			{
				String:      "Show Name - 1x02 - 1x03 - 1x04 - Ep Name",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"extra_ep_num": "04",
					"series_name":  "Show Name",
					"season_num":   "1",
					"ep_num":       "02",
					"extra_info":   "Ep Name",
				},
			},
		},
		Regex: pcre.MustCompile(`(?i)^(?P<series_name>.+?)[. _-]+`+ //  Show_Name and separator
			`(?P<season_num>\d+)x`+ //  1x
			`(?P<ep_num>\d+)`+ //  02 and separator
			`([. _-]+(?P=season_num)x`+ //  1x
			`(?P<extra_ep_num>\d+))+`+ //  03/etc and separator
			`[. _-]*((?P<extra_info>.+?)`+ //  Source_Quality_Etc-
			`((?<![. _-])(?<!WEB)`+ //  Make sure this is really the release group
			`-(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$`, pcre.CASELESS), //  Group
	},
	{
		Name: "standard",
		TestStrings: []TestString{
			{
				String:      "Show.Name.S01E02.Source.Quality.Etc-Group",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"series_name":   "Show.Name",
					"season_num":    "01",
					"release_group": "Group",
					"ep_num":        "02",
					"extra_info":    "Source.Quality.Etc",
				},
			},
			{
				String:      "Show Name - S01E02 - My Ep Name",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"series_name": "Show Name",
					"season_num":  "01",
					"ep_num":      "02",
					"extra_info":  "My Ep Name",
				},
			},
			{
				String:      "Show.Name.S01.E03.My.Ep.Name",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"series_name": "Show.Name",
					"ep_num":      "03",
					"extra_info":  "My.Ep.Name",
					"season_num":  "01",
				},
			},
			{
				String:      "Show.Name.S01E02E03.Source.Quality.Etc-Group",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"ep_num":        "02",
					"extra_info":    "Source.Quality.Etc",
					"series_name":   "Show.Name",
					"season_num":    "01",
					"extra_ep_num":  "03",
					"release_group": "Group",
				},
			},
			{
				String:      "Show Name - S01E02-03 - My Ep Name",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"extra_ep_num": "03",
					"ep_num":       "02",
					"extra_info":   "My Ep Name",
					"series_name":  "Show Name",
					"season_num":   "01",
				},
			},
			{
				String:      "Show.Name.S01.E02.E03",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"extra_ep_num": "03",
					"series_name":  "Show.Name",
					"season_num":   "01",
					"ep_num":       "02",
				},
			},
		},
		Regex: pcre.MustCompile(`(?i)^((?P<series_name>.+?)[. _-]+)?`+ //  Show_Name and separator
			`(\()?s(?P<season_num>\d+)[. _-]*`+ //  S01 and optional separator
			`e(?P<ep_num>\d+)(\))?`+ //  E02 and separator
			`(([. _-]*e|-)`+ //  linking e/- char
			`(?P<extra_ep_num>(?!(1080|720|480)[pi])\d+)(\))?)*`+ //  additional E03/etc
			`[. _-]*((?P<extra_info>.+?)`+ //  Source_Quality_Etc-
			`((?<![. _-])(?<!WEB)`+ //  Make sure this is really the release group
			`-(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$`, pcre.CASELESS), //  Group
	},
	{
		Name: "fov",
		TestStrings: []TestString{
			{
				String:      "Show_Name.1x02.Source_Quality_Etc-Group",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"series_name":   "Show_Name",
					"extra_info":    "Source_Quality_Etc",
					"release_group": "Group",
					"season_num":    "1",
					"ep_num":        "02",
				},
			},
			{
				String:      "Show Name - 1x02 - My Ep Name",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"season_num":  "1",
					"series_name": "Show Name",
					"ep_num":      "02",
					"extra_info":  "My Ep Name",
				},
			},
			{
				String:      "Show_Name.1x02x03x04.Source_Quality_Etc-Group",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"series_name":   "Show_Name",
					"extra_info":    "Source_Quality_Etc",
					"extra_ep_num":  "04",
					"release_group": "Group",
					"season_num":    "1",
					"ep_num":        "02",
				},
			},
			{
				String:      "Show Name - 1x02-03-04 - My Ep Name",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"season_num":   "1",
					"ep_num":       "02",
					"series_name":  "Show Name",
					"extra_info":   "My Ep Name",
					"extra_ep_num": "04",
				},
			},
		},
		Regex: pcre.MustCompile(`(?i)^((?P<series_name>.+?)[\[. _-]+)?`+ //  Show_Name and separator
			`(?P<season_num>\d+)x`+ //  1x
			`(?P<ep_num>\d+)`+ //  02 and separator
			`(([. _-]*x|-)`+ //  linking x/- char
			`(?P<extra_ep_num>`+
			`(?!(1080|720|480)[pi])(?!(?<=x)264)`+ //  ignore obviously wrong multi-eps
			`\d+))*`+ //  additional x03/etc
			`[\]. _-]*((?P<extra_info>.+?)`+ //  Source_Quality_Etc-
			`((?<![. _-])(?<!WEB)`+ //  Make sure this is really the release group
			`-(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$`, pcre.CASELESS), //  Group
	},
	{
		Name: "scene_date_format",
		TestStrings: []TestString{
			{
				String:      "Show.Name.2010.11.23.Source.Quality.Etc-Group",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"release_group": "Group",
					"air_date":      "2010.11.23",
					"extra_info":    "Source.Quality.Etc",
					"series_name":   "Show.Name",
				},
			},
			{
				String:      "Show Name - 2010-11-23 - Ep Name",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"extra_info":  "Ep Name",
					"series_name": "Show Name",
					"air_date":    "2010-11-23",
				},
			},
		},
		Regex: pcre.MustCompile(`(?i)^((?P<series_name>.+?)[. _-]+)?`+ //  Show_Name and separator
			`(?P<air_date>(\d+[. _-]\d+[. _-]\d+)|(\d+\w+[. _-]\w+[. _-]\d+))`+
			`[. _-]*((?P<extra_info>.+?)`+ //  Source_Quality_Etc-
			`((?<![. _-])(?<!WEB)`+ //  Make sure this is really the release group
			`-(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$`, pcre.CASELESS), //  Group
	},
	{
		Name: "scene_sports_format",
		/*
			TestStrings: []TestString{
				{
					String:      "Show.Name.100.Event.2010.11.23.Source.Quality.Etc-Group",
					ShouldMatch: true,
					MatchGroups: map[string]string{},
				},
				{
					String:      "Show.Name.2010.11.23.Source.Quality.Etc-Group",
					ShouldMatch: true,
					MatchGroups: map[string]string{},
				},
				{
					String:      "Show Name - 2010-11-23 - Ep Name",
					ShouldMatch: true,
					MatchGroups: map[string]string{},
				},
			},
		*/
		Regex: pcre.MustCompile(`^(?P<series_name>.*?(UEFA|MLB|ESPN|WWE|MMA|UFC|TNA|EPL|NASCAR|NBA|NFL|NHL|NRL|PGA|SUPER LEAGUE|FORMULA|FIFA|NETBALL|MOTOGP).*?)[. _-]+`+
			`((?P<series_num>\d{1,3})[. _-]+)?`+
			`(?P<air_date>(\d+[. _-]\d+[. _-]\d+)|(\d+\w+[. _-]\w+[. _-]\d+))[. _-]+`+
			`((?P<extra_info>.+?)((?<![. _-])`+
			`(?<!WEB)-(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$`, pcre.CASELESS),
	},
	{
		Name: "stupid",
		TestStrings: []TestString{
			{
				String:      "tpz-abc102",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"season_num":    "1",
					"ep_num":        "02",
					"release_group": "tpz",
				},
			},
		},
		Regex: pcre.MustCompile(`(?i)(?P<release_group>.+?)-\w+?[\. ]?`+ //  tpz-abc
			`(?!264)`+ //  don't count x264
			`(?P<season_num>\d{1,2})`+ //  1
			`(?P<ep_num>\d{2})$`, pcre.CASELESS), //  02
	},
	{
		Name: "verbose",
		TestStrings: []TestString{
			{
				String:      "Show Name Season 1 Episode 2 Ep Name",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"season_num":  "1",
					"extra_info":  "Ep Name",
					"series_name": "Show Name",
					"ep_num":      "2",
				},
			},
		},
		Regex: pcre.MustCompile(`(?i)^(?P<series_name>.+?)[. _-]+`+ //  Show Name and separator
			`season[. _-]+`+ //  season and separator
			`(?P<season_num>\d+)[. _-]+`+ //  1
			`episode[. _-]+`+ //  episode and separator
			`(?P<ep_num>\d+)[. _-]+`+ //  02 and separator
			`(?P<extra_info>.+)$`, pcre.CASELESS), //  Source_Quality_Etc-
	},
	{
		Name: "season_only",
		TestStrings: []TestString{
			{
				String:      "Show.Name.S01.Source.Quality.Etc-Group",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"season_num":    "01",
					"release_group": "Group",
					"series_name":   "Show.Name",
					"extra_info":    "Source.Quality.Etc",
				},
			},
		},

		Regex: pcre.MustCompile(`(?i)^((?P<series_name>.+?)[. _-]+)?`+ //  Show_Name and separator
			`s(eason[. _-])?`+ //  S01/Season 01
			`(?P<season_num>\d+)[. _-]*`+ //  S01 and optional separator
			`[. _-]*((?P<extra_info>.+?)`+ //  Source_Quality_Etc-
			`((?<![. _-])(?<!WEB)`+ //  Make sure this is really the release group
			`-(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$`, pcre.CASELESS), //  Group
	},
	{
		Name: "no_season_multi_ep",
		/*
			TestStrings: []TestString{
				{
					String:      "Show.Name.E02-03",
					ShouldMatch: true,
					MatchGroups: map[string]string{},
				},

				{
					String:      "Show.Name.E02.2010",
					ShouldMatch: true,
					MatchGroups: map[string]string{},
				},
			},
		*/
		Regex: pcre.MustCompile(`(?i)^((?P<series_name>.+?)[. _-]+)?`+ //  Show_Name and separator
			`(e(p(isode)?)?|part|pt)[. _-]?`+ //  e, ep, episode, or part
			`(?P<ep_num>(\d+|[ivx]+))`+ //  first ep num
			`((([. _-]+(and|&|to)[. _-]+)|-)`+ //  and/&/to joiner
			`(?P<extra_ep_num>(?!(1080|720|480)[pi])(\d+|[ivx]+))[. _-])`+ //  second ep num
			`([. _-]*(?P<extra_info>.+?)`+ //  Source_Quality_Etc-
			`((?<![. _-])(?<!WEB)`+ //  Make sure this is really the release group
			`-(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$`, pcre.CASELESS), //  Group
	},
	{
		Name: "no_season_general",
		TestStrings: []TestString{
			{
				String:      "Show.Name.E23.Test",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"series_name": "Show.Name",
					"ep_num":      "23",
					"extra_info":  "Test",
				},
			},

			{
				String:      "Show.Name.Part.3.Source.Quality.Etc-Group",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"series_name":   "Show.Name",
					"ep_num":        "3",
					"extra_info":    "Source.Quality.Etc",
					"release_group": "Group",
				},
			},
			{
				String:      "Show.Name.Part.1.and.Part.2.Blah-Group",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"series_name":   "Show.Name",
					"extra_ep_num":  "2",
					"extra_info":    "Blah",
					"release_group": "Group",
					"ep_num":        "1",
				},
			},
		},

		Regex: pcre.MustCompile(`(?i)^((?P<series_name>.+?)[. _-]+)?`+ //  Show_Name and separator
			`(e(p(isode)?)?|part|pt)[. _-]?`+ //  e, ep, episode, or part
			`(?P<ep_num>(\d+|([ivx]+(?=[. _-]))))`+ //  first ep num
			`([. _-]+((and|&|to)[. _-]+)?`+ //  and/&/to joiner
			`((e(p(isode)?)?|part|pt)[. _-]?)`+ //  e, ep, episode, or part
			`(?P<extra_ep_num>(?!(1080|720|480)[pi])`+
			`(\d+|([ivx]+(?=[. _-]))))[. _-])*`+ //  second ep num
			`([. _-]*(?P<extra_info>.+?)`+ //  Source_Quality_Etc-
			`((?<![. _-])(?<!WEB)`+ //  Make sure this is really the release group
			`-(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$`, pcre.CASELESS), //  Group
	},
	{
		Name: "no_season",
		/*
			TestStrings: []TestString{
				{
					String:      "Show Name - 01 - Ep Name",
					ShouldMatch: true,
					MatchGroups: map[string]string{},
				},

				{
					String:      "01 - Ep Name",
					ShouldMatch: true,
					MatchGroups: map[string]string{},
				},
			},
		*/
		Regex: pcre.MustCompile(`(?i)^((?P<series_name>.+?)(?:[. _-]{2,}|[. _]))?`+ //  Show_Name and separator
			`(?P<ep_num>\d{1,3})`+ //  02
			`(?:-(?P<extra_ep_num>\d{1,3}))*`+ //  -03-04-05 etc
			`\s?of?\s?\d{1,3}?`+ //  of joiner (with or without spaces) and series total ep
			`[. _-]+((?P<extra_info>.+?)`+ //  Source_Quality_Etc-
			`((?<![. _-])(?<!WEB)`+ //  Make sure this is really the release group
			`-(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$`, pcre.CASELESS), //  Group
	},
	{
		Name: "bare",
		TestStrings: []TestString{
			{
				String:      "Show.Name.102.Source.Quality.Etc-Group",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"season_num":    "1",
					"extra_info":    "Source.Quality.Etc",
					"release_group": "Group",
					"series_name":   "Show.Name",
					"ep_num":        "02",
				},
			},
		},

		Regex: pcre.MustCompile(`(?i)^(?P<series_name>.+?)[. _-]+`+ //  Show_Name and separator
			`(?P<season_num>\d{1,2})`+ //  1
			`(?P<ep_num>\d{2})`+ //  02 and separator
			`([. _-]+(?P<extra_info>(?!\d{3}[. _-]+)[^-]+)`+ //  Source_Quality_Etc-
			`(-(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$`, pcre.CASELESS), //  Group
	},
}

var AnimeRegex = []NameRegex{
	// Also from Sickrage
	{
		Name: "anime_ultimate",
		Regex: pcre.MustCompile(`^(?:\[(?P<release_group>.+?)\][ ._-]*)`+
			`(?P<series_name>.+?)[ ._-]+`+
			`(?P<ep_ab_num>((?!(1080|720|480)[pi])|(?![hx].?264))\d{1,3})`+
			`(-(?P<extra_ab_ep_num>((?!(1080|720|480)[pi])|(?![hx].?264))\d{1,3}))?[ ._-]+?`+
			`(?:v(?P<version>[0-9]))?`+
			`(?:[\w\.]*)`+
			`(?:(?:(?:[\[\(])(?P<extra_info>\d{3,4}[xp]?\d{0,4}[\.\w\s-]*)(?:[\]\)]))|(?:\d{3,4}[xp]))`+
			`(?:[ ._]?\[(?P<crc>\w+)\])?`+
			`.*?`, pcre.CASELESS),
	},
	{
		Name: "anime_standard",
		TestStrings: []TestString{
			{
				String:      "[Group Name] Show Name.13-14",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"series_name":     "Show Name",
					"ep_ab_num":       "13",
					"extra_ab_ep_num": "14",
					"release_group":   "Group Name",
				},
			},
			{
				String:      "[Group Name] Show Name - 13-14",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"series_name":     "Show Name",
					"ep_ab_num":       "13",
					"extra_ab_ep_num": "14",
					"release_group":   "Group Name",
				},
			},

			{
				String:      "Show Name - 13-14",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"series_name":     "Show Name",
					"ep_ab_num":       "13",
					"extra_ab_ep_num": "14",
				},
			},
			{
				String:      "[Group Name] Show Name.13",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"series_name":   "Show Name",
					"ep_ab_num":     "13",
					"release_group": "Group Name",
				},
			},
			{
				String:      "[Group Name] Show Name - 13",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"series_name":   "Show Name",
					"ep_ab_num":     "13",
					"release_group": "Group Name",
				},
			},
			{
				String:      "Show Name 13",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"series_name": "Show Name",
					"ep_ab_num":   "13",
				},
			},
		},
		Regex: pcre.MustCompile(`^(\[(?P<release_group>.+?)\][ ._-]*)?`+ //  Release Group and separator
			`(?P<series_name>.+?)[ ._-]+`+ //  Show_Name and separator
			`(?P<ep_ab_num>((?!(1080|720|480)[pi])|(?![hx].?264))\d{1,3})`+ //  E01
			`(-(?P<extra_ab_ep_num>((?!(1080|720|480)[pi])|(?![hx].?264))\d{1,3}))?`+ //  E02
			`(v(?P<version>[0-9]))?`+ //  version
			`([ ._-]+\[(?P<extra_info>\d{3,4}[xp]?\d{0,4}[\.\w\s-]*)\])?`+ //  Source_Quality_Etc-
			`(\[(?P<crc>\w{8})\])?`+ //  CRC
			`.*?`, pcre.CASELESS), //  Separator and EOL
	},
	{
		Name: "anime_standard_round",
		TestStrings: []TestString{
			{
				String:      "[Stratos-Subs]_Infinite_Stratos_-_12_(1280x720_H.264_AAC)_[379759DB]",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"series_name":   "Infinite_Stratos",
					"extra_info":    "1280x720_H.264_AAC",
					"release_group": "Stratos-Subs",
					"ep_ab_num":     "12",
				},
			},
			{
				String:      "[ShinBunBu-Subs] Bleach - 02-03 (CX 1280x720 x264 AAC)",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"series_name":     "Bleach",
					"extra_ab_ep_num": "03",
					"ep_ab_num":       "02",
					"extra_info":      "CX 1280x720 x264 AAC",
					"release_group":   "ShinBunBu-Subs",
				},
			},
		},
		Regex: pcre.MustCompile(`(?i)^(\[(?P<release_group>.+?)\][ ._-]*)?`+ //  Release Group and separator
			`(?P<series_name>.+?)[ ._-]+`+ //  Show_Name and separator
			`(?P<ep_ab_num>((?!(1080|720|480)[pi])|(?![hx].?264))\d{1,3})`+ //  E01
			`(-(?P<extra_ab_ep_num>((?!(1080|720|480)[pi])|(?![hx].?264))\d{1,3}))?`+ //  E02
			`(v(?P<version>[0-9]))?`+ //  version
			`[ ._-]+\((?P<extra_info>(CX[ ._-]?)?\d{3,4}[xp]?\d{0,4}[\.\w\s-]*)\)`+ //  Source_Quality_Etc-
			`(\[(?P<crc>\w{8})\])?`+ //  CRC
			`.*?`, pcre.CASELESS), //  Separator and EOL
	},
	{
		Name: "anime_slash",
		TestStrings: []TestString{
			{
				String:      "[SGKK] Bleach 312v1 [720p/MKV]",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"series_name":   "Bleach",
					"release_group": "SGKK",
					"ep_ab_num":     "312",
					"extra_info":    "720p",
				},
			},
		},
		Regex: pcre.MustCompile(`(?i)^(\[(?P<release_group>.+?)\][ ._-]*)?`+ //  Release Group and separator
			`(?P<series_name>.+?)[ ._-]+`+ //  Show_Name and separator
			`(?P<ep_ab_num>((?!(1080|720|480)[pi])|(?![hx].?264))\d{1,3})`+ //  E01
			`(-(?P<extra_ab_ep_num>((?!(1080|720|480)[pi])|(?![hx].?264))\d{1,3}))?`+ //  E02
			`(v(?P<version>[0-9]))?`+ //  version
			`[ ._-]+\[(?P<extra_info>\d{3,4}p)`+ //  Source_Quality_Etc-
			`(\[(?P<crc>\w{8})\])?`+ //  CRC
			`.*?`, pcre.CASELESS), //  Separator and EOL
	},
	{
		Name: "anime_standard_codec",
		TestStrings: []TestString{
			{
				String:      "[Ayako]_Infinite_Stratos_-_IS_-_07_[H264][720p][EB7838FC]",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"extra_info":    "720p",
					"release_group": "Ayako",
					"series_name":   "Infinite_Stratos",
					"ep_ab_num":     "07",
				},
			},
			{
				String:      "[Ayako] Infinite Stratos - IS - 07v2 [H264][720p][44419534]",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"series_name":   "Infinite Stratos",
					"ep_ab_num":     "07",
					"extra_info":    "720p",
					"release_group": "Ayako",
				},
			},
			{
				String:      "[Ayako-Shikkaku] Oniichan no Koto Nanka Zenzen Suki Janain Dakara ne - 10 [LQ][h264][720p] [8853B21C]",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"ep_ab_num":     "10",
					"series_name":   "Oniichan no Koto Nanka Zenzen Suki Janain Dakara ne",
					"extra_info":    "720p",
					"release_group": "Ayako-Shikkaku",
				},
			},
		},
		Regex: pcre.MustCompile(`(?i)^(\[(?P<release_group>.+?)\][ ._-]*)?`+ //  Release Group and separator
			`(?P<series_name>.+?)[ ._]*`+ //  Show_Name and separator
			`([ ._-]+-[ ._-]+[A-Z]+[ ._-]+)?[ ._-]+`+ //  funny stuff, this is sooo nuts ! this will kick me in the butt one day
			`(?P<ep_ab_num>((?!(1080|720|480)[pi])|(?![hx].?264))\d{1,3})`+ //  E01
			`(-(?P<extra_ab_ep_num>((?!(1080|720|480)[pi])|(?![hx].?264))\d{1,3}))?`+ //  E02
			`(v(?P<version>[0-9]))?`+ //  version
			`([ ._-](\[\w{1,2}\])?\[[a-z][.]?\w{2,4}\])?`+ // codec
			`[ ._-]*\[(?P<extra_info>(\d{3,4}[xp]?\d{0,4})?[\.\w\s-]*)\]`+ //  Source_Quality_Etc-
			`(\[(?P<crc>\w{8})\])?`+
			`.*?`, pcre.CASELESS), //  Separator and EOL
	},
	{
		Name: "anime_codec_crc",
		Regex: pcre.MustCompile(`^(?:\[(?P<release_group>.*?)\][ ._-]*)?`+
			`(?:(?P<series_name>.*?)[ ._-]*)?`+
			`(?:(?P<ep_ab_num>(((?!(1080|720|480)[pi])|(?![hx].?264))\d{1,3}))[ ._-]*).+?`+
			`(?:\[(?P<codec>.*?)\][ ._-]*)`+
			`(?:\[(?P<crc>\w{8})\])?`+
			`.*?`, pcre.CASELESS),
	},
	{
		Name: "anime_and_normal",
		TestStrings: []TestString{
			{
				String:      "Bleach - s16e03-04 - 313-314",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"ep_num":          "03",
					"ep_ab_num":       "313",
					"series_name":     "Bleach",
					"season_num":      "16",
					"extra_ep_num":    "04",
					"extra_ab_ep_num": "314",
				},
			},
			{
				String:      "Bleach.s16e03-04.313-314",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"ep_num":          "03",
					"ep_ab_num":       "313",
					"extra_ep_num":    "04",
					"series_name":     "Bleach",
					"season_num":      "16",
					"extra_ab_ep_num": "314",
				},
			},
			{
				String:      "Bleach s16e03e04 313-314",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"ep_ab_num":       "313",
					"extra_ep_num":    "04",
					"series_name":     "Bleach",
					"season_num":      "16",
					"ep_num":          "03",
					"extra_ab_ep_num": "314",
				},
			},
		},
		Regex: pcre.MustCompile(`(?i)^(?P<series_name>.+?)[ ._-]+`+ //  start of string and series name and non optinal separator
			`[sS](?P<season_num>\d+)[. _-]*`+ //  S01 and optional separator
			`[eE](?P<ep_num>\d+)`+ //  epipisode E02
			`(([. _-]*e|-)`+ //  linking e/- char
			`(?P<extra_ep_num>\d+))*`+ //  additional E03/etc
			`([ ._-]{2,}|[ ._]+)`+ //  if "-" is used to separate at least something else has to be there(->{2,}) "s16e03-04-313-314" would make sens any way
			`((?P<ep_ab_num>((?!(1080|720|480)[pi])|(?![hx].?264))\d{1,3}))?`+ //  absolute number
			`(-(?P<extra_ab_ep_num>((?!(1080|720|480)[pi])|(?![hx].?264))\d{1,3}))?`+ //  "-" as separator and anditional absolute number, all optinal
			`(v(?P<version>[0-9]))?`+ //  the version e.g. "v2"
			`.*?`, pcre.CASELESS),
	},
	{
		Name: "anime_and_normal_x",
		TestStrings: []TestString{
			{
				String:      "Bleach - s16x03-04 - 313-314",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"ep_num":          "03",
					"ep_ab_num":       "313",
					"series_name":     "Bleach",
					"season_num":      "16",
					"extra_ep_num":    "04",
					"extra_ab_ep_num": "314",
				},
			},
			{
				String:      "Bleach.s16x03-04.313-314",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"ep_num":          "03",
					"ep_ab_num":       "313",
					"extra_ep_num":    "04",
					"series_name":     "Bleach",
					"season_num":      "16",
					"extra_ab_ep_num": "314",
				},
			},
			{
				String:      "Bleach s16x03e04 313-314",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"ep_ab_num":       "313",
					"extra_ep_num":    "04",
					"series_name":     "Bleach",
					"season_num":      "16",
					"ep_num":          "03",
					"extra_ab_ep_num": "314",
				},
			},
		},
		Regex: pcre.MustCompile(`(?i)^(?P<series_name>.+?)[ ._-]+`+ //  start of string and series name and non optinal separator
			`[sS](?P<season_num>\d+)[. _-]*`+ //  S01 and optional separator
			`[xX](?P<ep_num>\d+)`+ //  epipisode E02
			`(([. _-]*e|-)`+ //  linking e/- char
			`(?P<extra_ep_num>\d+))*`+ //  additional E03/etc
			`([ ._-]{2,}|[ ._]+)`+ //  if "-" is used to separate at least something else has to be there(->{2,}) "s16e03-04-313-314" would make sens any way
			`((?P<ep_ab_num>((?!(1080|720|480)[pi])|(?![hx].?264))\d{1,3}))?`+ //  absolute number
			`(-(?P<extra_ab_ep_num>((?!(1080|720|480)[pi])|(?![hx].?264))\d{1,3}))?`+ //  "-" as separator and anditional absolute number, all optinal
			`(v(?P<version>[0-9]))?`+ //  the version e.g. "v2"
			`.*?`, pcre.CASELESS),
	},
	{
		Name: "anime_and_normal_reverse",
		TestStrings: []TestString{
			{
				String:      "Bleach - 313-314 - s16e03-04",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"ep_num":          "03",
					"ep_ab_num":       "313",
					"series_name":     "Bleach",
					"season_num":      "16",
					"extra_ep_num":    "04",
					"extra_ab_ep_num": "314",
				},
			},
		},
		Regex: pcre.MustCompile(`(?i)^(?P<series_name>.+?)[ ._-]+`+ //  start of string and series name and non optinal separator
			`(?P<ep_ab_num>((?!(1080|720|480)[pi])|(?![hx].?264))\d{1,3})`+ //  absolute number
			`(-(?P<extra_ab_ep_num>((?!(1080|720|480)[pi])|(?![hx].?264))\d{1,3}))?`+ //  "-" as separator and anditional absolute number, all optinal
			`(v(?P<version>[0-9]))?`+ //  the version e.g. "v2"
			`([ ._-]{2,}|[ ._]+)`+ //  if "-" is used to separate at least something else has to be there(->{2,}) "s16e03-04-313-314" would make sens any way
			`[sS](?P<season_num>\d+)[. _-]*`+ //  S01 and optional separator
			`[eE](?P<ep_num>\d+)`+ //  epipisode E02
			`(([. _-]*e|-)`+ //  linking e/- char
			`(?P<extra_ep_num>\d+))*`+ //  additional E03/etc
			`.*?`, pcre.CASELESS),
	},
	{
		Name: "anime_and_normal_front",
		TestStrings: []TestString{
			{
				String:      "165.Naruto Shippuuden.s08e014",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"series_name": "Naruto Shippuuden",
					"season_num":  "08",
					"ep_num":      "014",
					"ep_ab_num":   "165",
				},
			},
		},
		Regex: pcre.MustCompile(`(?i)^(?P<ep_ab_num>((?!(1080|720|480)[pi])|(?![hx].?264))\d{1,3})`+ //  start of string and absolute number
			`(-(?P<extra_ab_ep_num>((?!(1080|720|480)[pi])|(?![hx].?264))\d{1,3}))?`+ //  "-" as separator and anditional absolute number, all optinal
			`(v(?P<version>[0-9]))?[ ._-]+`+ //  the version e.g. "v2"
			`(?P<series_name>.+?)[ ._-]+`+
			`[sS](?P<season_num>\d+)[. _-]*`+ //  S01 and optional separator
			`[eE](?P<ep_num>\d+)`+
			`(([. _-]*e|-)`+ //  linking e/- char
			`(?P<extra_ep_num>\d+))*`+ //  additional E03/etc
			`.*?`, pcre.CASELESS),
	},
	{
		Name: "anime_ep_name",
		Regex: pcre.MustCompile(`^(?:\[(?P<release_group>.+?)\][ ._-]*)`+
			`(?P<series_name>.+?)[ ._-]+`+
			`(?P<ep_ab_num>((?!(1080|720|480)[pi])|(?![hx].?264))\d{1,3})`+
			`(-(?P<extra_ab_ep_num>((?!(1080|720|480)[pi])|(?![hx].?264))\d{1,3}))?[ ._-]*?`+
			`(?:v(?P<version>[0-9])[ ._-]+?)?`+
			`(?:.+?[ ._-]+?)?`+
			`\[(?P<extra_info>\w+)\][ ._-]?`+
			`(?:\[(?P<crc>\w{8})\])?`+
			`.*?`, pcre.CASELESS),
	},
	{
		Name: "anime_bare",
		TestStrings: []TestString{
			{
				String:      "One Piece - 102",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"series_name": "One Piece",
					"ep_ab_num":   "102",
				},
			},
			{
				String:      "[ACX]_Wolf's_Spirit_001.mkv",
				ShouldMatch: true,
				MatchGroups: map[string]string{
					"release_group": "ACX",
					"series_name":   "Wolf's_Spirit",
					"ep_ab_num":     "001",
				},
			},
		},
		Regex: pcre.MustCompile(`^(\[(?P<release_group>.+?)\][ ._-]*)?`+
			`(?P<series_name>.+?)[ ._-]+`+ //  Show_Name and separator
			`(?P<ep_ab_num>((?!(1080|720|480)[pi])|(?![hx].?264))\d{1,3})`+ //  E01
			`(-(?P<extra_ab_ep_num>((?!(1080|720|480)[pi])|(?![hx].?264))\d{1,3}))?`+ //  E02
			`(v(?P<version>[0-9]))?`+ //  v2
			`.*?`, pcre.CASELESS), //  Separator and EOL
	},
}
