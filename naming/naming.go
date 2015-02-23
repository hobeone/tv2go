package naming

import (
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hobeone/go-pcre"

	"github.com/gholt/brimtime"
	"github.com/golang/glog"
)

type NameRegex struct {
	Name  string
	Regex pcre.Regexp
}

// NameRegexes is a list of Regular Expressions to try in order when trying to
// extract information from a filename.
var NameRegexes = []NameRegex{

	{
		Name: "standard_repeat",
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
		Regex: pcre.MustCompile(`(?i)^((?P<series_name>.+?)[. _-]+)?`+ //  Show_Name and separator
			`(?P<air_date>(\d+[. _-]\d+[. _-]\d+)|(\d+\w+[. _-]\w+[. _-]\d+))`+
			`[. _-]*((?P<extra_info>.+?)`+ //  Source_Quality_Etc-
			`((?<![. _-])(?<!WEB)`+ //  Make sure this is really the release group
			`-(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$`, pcre.CASELESS), //  Group
	},
	{
		Name: "scene_sports_format",
		Regex: pcre.MustCompile(`^(?P<series_name>.*?(UEFA|MLB|ESPN|WWE|MMA|UFC|TNA|EPL|NASCAR|NBA|NFL|NHL|NRL|PGA|SUPER LEAGUE|FORMULA|FIFA|NETBALL|MOTOGP).*?)[. _-]+`+
			`((?P<series_num>\d{1,3})[. _-]+)?`+
			`(?P<air_date>(\d+[. _-]\d+[. _-]\d+)|(\d+\w+[. _-]\w+[. _-]\d+))[. _-]+`+
			`((?P<extra_info>.+?)((?<![. _-])`+
			`(?<!WEB)-(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$`, pcre.CASELESS),
	},
	{
		Name: "stupid",
		Regex: pcre.MustCompile(`(?i)(?P<release_group>.+?)-\w+?[\. ]?`+ //  tpz-abc
			`(?!264)`+ //  don't count x264
			`(?P<season_num>\d{1,2})`+ //  1
			`(?P<ep_num>\d{2})$`, pcre.CASELESS), //  02
	},
	{
		Name: "verbose",
		Regex: pcre.MustCompile(`(?i)^(?P<series_name>.+?)[. _-]+`+ //  Show Name and separator
			`season[. _-]+`+ //  season and separator
			`(?P<season_num>\d+)[. _-]+`+ //  1
			`episode[. _-]+`+ //  episode and separator
			`(?P<ep_num>\d+)[. _-]+`+ //  02 and separator
			`(?P<extra_info>.+)$`, pcre.CASELESS), //  Source_Quality_Etc-
	},
	{
		Name: "season_only",
		Regex: pcre.MustCompile(`(?i)^((?P<series_name>.+?)[. _-]+)?`+ //  Show_Name and separator
			`s(eason[. _-])?`+ //  S01/Season 01
			`(?P<season_num>\d+)[. _-]*`+ //  S01 and optional separator
			`[. _-]*((?P<extra_info>.+?)`+ //  Source_Quality_Etc-
			`((?<![. _-])(?<!WEB)`+ //  Make sure this is really the release group
			`-(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$`, pcre.CASELESS), //  Group
	},
	{
		Name: "no_season_multi_ep",
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
		Regex: pcre.MustCompile(`(?i)^(?P<series_name>.+?)[. _-]+`+ //  Show_Name and separator
			`(?P<season_num>\d{1,2})`+ //  1
			`(?P<ep_num>\d{2})`+ //  02 and separator
			`([. _-]+(?P<extra_info>(?!\d{3}[. _-]+)[^-]+)`+ //  Source_Quality_Etc-
			`(-(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$`, pcre.CASELESS), //  Group
	},
}

var (
	mediaExtensions = []string{
		"avi", "mkv", "mpg", "mpeg", "wmv",
		"ogm", "mp4", "iso", "img", "divx",
		"m2ts", "m4v", "ts", "flv", "f4v",
		"mov", "rmvb", "vob", "dvr-ms", "wtv",
		"ogv", "3gp", "webm",
	}

	sampleRegex = regexp.MustCompile(`(?i)(^|[\W_])(sample\d*)[\W_]`)
	extrasRegex = regexp.MustCompile(`(?i)extras?$`)
)

// IsMediaExtension checks if the given string matches a known Media file
// extension.
func IsMediaExtension(extension string) bool {
	extension = strings.TrimLeft(extension, ".")
	extension = strings.ToLower(extension)
	for _, ext := range mediaExtensions {
		if ext == extension {
			return true
		}
	}
	return false
}

func stripExtension(fname string) string {
	extension := filepath.Ext(fname)
	return fname[0 : len(fname)-len(extension)]
}

// IsMediaFile checks if the given string is a media file
func IsMediaFile(filename string) bool {
	// ignore samples
	if sampleRegex.MatchString(filename) {
		return false
	}
	// ignore Mac resource fork files
	if strings.HasPrefix(filename, "._") {
		return false
	}

	extension := filepath.Ext(filename)
	name := stripExtension(filename)

	if extrasRegex.MatchString(name) {
		return false
	}

	return IsMediaExtension(extension)
}

/*
*
* Name Parser
 */

type ParseResult struct {
	OriginalName           string
	SeriesName             string
	SeasonNumber           int64
	EpisodeNumbers         []int64
	ExtraInfo              string
	ReleaseGroup           string
	AirDate                time.Time
	AbsoluteEpisodeNumbers []int64
	Show                   string
	Score                  int
	Quality                string
	Version                string
	RegexUsed              string
}

type byScore []ParseResult

func (a byScore) Len() int           { return len(a) }
func (a byScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byScore) Less(i, j int) bool { return a[i].Score < a[j].Score }

type NameParser struct {
	FileName string
}

func NewNameParser(filename string) *NameParser {
	return &NameParser{
		FileName: filename,
	}
}

var knownMatches = []string{
	"series_name",
	"season_num",
	"ep_num",
	"ep_ab_num",
	"extra_ep_num",
	"extra_info",
	"release_group",
	"air_date",
	"series_num",
}

// Return named matches in a map
func regexNamedMatch(re *pcre.Regexp, str string) (map[string]string, bool) {
	m := re.MatcherString(str, 0)
	if !m.Matches() {
		return nil, false
	}

	result := make(map[string]string, len(knownMatches))
	for _, name := range knownMatches {
		extract, err := m.NamedString(name)
		if err == nil && extract != "" {
			result[name] = extract
		}
	}
	return result, true
}

func (np *NameParser) parseString(name string) (*ParseResult, error) {
	var matchResults []ParseResult
	for i, r := range NameRegexes {
		glog.Infof("Trying to match %s with regex %s", name, r.Name)
		if matches, ok := regexNamedMatch(&r.Regex, name); ok {
			glog.Infof("Matched %s with regex %s", name, r.Name)
			pr := ParseResult{
				OriginalName: name,
				RegexUsed:    r.Name,
				Score:        0 - i,
			}

			if m, ok := matches["series_name"]; ok {
				pr.SeriesName = m
				//pr.SeriesName = cleanSeriesName(pr.SeriesName)
				pr.Score++
			}
			if _, ok := matches["series_num"]; ok {
				pr.Score++
			}
			if m, ok := matches["season_num"]; ok {
				m = strings.TrimLeft(m, "0")
				glog.Infof("Converting Season '%s' to int", m)
				sn, err := strconv.ParseInt(m, 10, 64)
				if err != nil {
					glog.Errorf("Error converting season_num '%s' to int from string: %s", m, pr.OriginalName)
					continue
				}
				pr.SeasonNumber = sn
				pr.Score++
			}
			if m, ok := matches["ep_num"]; ok {
				m = strings.TrimLeft(m, "0")
				// Maybe handle Roman numberals like SickRage?
				en, err := strconv.ParseInt(m, 10, 64)
				if err != nil {
					glog.Errorf("Error converting ep_num '%s' to int from string: %s", m, pr.OriginalName)
					continue
				}
				if extraEp, ok := matches["extra_ep_num"]; ok {
					m = strings.TrimLeft(m, "0")
					extraEpCvt, err := strconv.ParseInt(extraEp, 10, 64)
					if err != nil {
						glog.Errorf("Error converting extra_ep_num '%s' to int from string: %s", extraEp, pr.OriginalName)
					} else {
						pr.EpisodeNumbers = []int64{en, extraEpCvt}
					}
				} else {
					pr.EpisodeNumbers = []int64{en}
				}
			}
			if m, ok := matches["ep_ab_num"]; ok {
				en, err := strconv.ParseInt(m, 10, 64)
				if err != nil {
					glog.Errorf("Error converting ep_ab_num '%s' to int from string: %s", m, pr.OriginalName)
					continue
				}
				// Handle extra ab number
				pr.AbsoluteEpisodeNumbers = []int64{en}
			}
			if m, ok := matches["air_date"]; ok {
				year, month, day := brimtime.TranslateYMD(m, []string{"", "D", "MD", "YMD"})
				if year != 0 && month != 0 {
					pr.AirDate = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
					pr.Score++
				} else {
					continue
				}
			}
			if _, ok := matches["extra_info"]; ok {
				// rando crap here
			}
			if m, ok := matches["release_group"]; ok {
				pr.ReleaseGroup = m
				pr.Score++
			}
			if m, ok := matches["version"]; ok {
				pr.Version = m
			}
			matchResults = append(matchResults, pr)
		}
	}

	// There's a whole mess of other logic that goes on in Sickbeard, but it's
	// super gross and we'll skip it for now:
	sort.Sort(sort.Reverse(byScore(matchResults)))
	if len(matchResults) > 0 {
		best := &matchResults[0]
		glog.Infof("Chose best match with regex %s, score %d", best.RegexUsed, best.Score)
		return &matchResults[0], nil
	}
	glog.Warningf("Couldn't match %s with any regex", name)
	return &ParseResult{}, fmt.Errorf("Couldn't parse string %s", name)
}

// Parse tries to extract show and episode information from a file path.
func (np *NameParser) Parse(name string) ParseResult {
	dirName, fileName := filepath.Split(name)
	fileName = stripExtension(fileName)
	dirNameBase := filepath.Base(dirName)

	fileNameResult, _ := np.parseString(fileName)
	dirNameResult, _ := np.parseString(dirNameBase)
	finalRes, _ := np.parseString(name)

	combineResults(finalRes, fileNameResult, dirNameResult, "AirDate")
	combineResults(finalRes, fileNameResult, dirNameResult, "AbsoluteEpisodeNumbers")
	combineResults(finalRes, fileNameResult, dirNameResult, "SeasonNumber")
	combineResults(finalRes, fileNameResult, dirNameResult, "EpisodeNumbers")
	combineResults(finalRes, dirNameResult, fileNameResult, "SeriesName")
	combineResults(finalRes, dirNameResult, fileNameResult, "ExtraInfo")
	combineResults(finalRes, dirNameResult, fileNameResult, "ReleaseGroup")
	combineResults(finalRes, dirNameResult, fileNameResult, "Version")
	// TODO: set this
	combineResults(finalRes, fileNameResult, dirNameResult, "Quality")
	return *finalRes
}

// From src/pkg/encoding/json.
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func combineResults(finalRes, fileRes, dirRes *ParseResult, field string) error {
	r := reflect.Indirect(reflect.ValueOf(finalRes)).FieldByName(field)
	fVal := reflect.Indirect(reflect.ValueOf(fileRes)).FieldByName(field)
	dVal := reflect.Indirect(reflect.ValueOf(dirRes)).FieldByName(field)

	if !r.IsValid() || !fVal.IsValid() || !dVal.IsValid() {
		return fmt.Errorf("Invalid field name given: %s", field)
	}

	if r.CanSet() {
		if !isEmptyValue(fVal) {
			r.Set(fVal)
		} else {
			r.Set(dVal)
		}
	}

	return nil
}
