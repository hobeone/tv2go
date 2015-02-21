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

	"github.com/gholt/brimtime"
	"github.com/golang/glog"
)

// NameRegexes is a list of Regular Expressions to try in order when trying to
// extract information from a filename.
var NameRegexes = []*regexp.Regexp{
	regexp.MustCompile(`^(?i)((?P<series_name>.+?)[. _-]+)?(\()?s(?P<season_num>\d+)[. _-]*e(?P<ep_num>\d+)(\))?(([. _-]*e|-)(?P<extra_ep_num>\d+)(\))?)*[. _-]*((?P<extra_info>.+?)(-(?P<release_group>[^- ]+([. _-]\[.*\])?))?)?$`),
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

// Return named matches in a map
func regexNamedMatch(re *regexp.Regexp, str string) (map[string]string, bool) {
	res := re.FindStringSubmatch(str)
	if res == nil {
		return nil, false
	}

	result := make(map[string]string, len(re.SubexpNames()))
	for i, name := range re.SubexpNames() {
		if res[i] != "" {
			result[name] = res[i]
		}
	}

	return result, true
}

func (np *NameParser) parseString(name string) (*ParseResult, error) {
	var matchResults []ParseResult
	for i, r := range NameRegexes {
		if matches, ok := regexNamedMatch(r, name); ok {
			pr := ParseResult{
				OriginalName: name,
				RegexUsed:    r.String(),
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
				sn, err := strconv.ParseInt(m, 10, 64)
				if err != nil {
					glog.Errorf("Error converting %s to int from string: %s", m, pr.OriginalName)
					continue
				}
				pr.SeasonNumber = sn
				pr.Score++
			}
			if m, ok := matches["ep_num"]; ok {
				// Maybe handle Roman numberals like SickRage?
				en, err := strconv.ParseInt(m, 10, 64)
				if err != nil {
					glog.Errorf("Error converting %s to int from string: %s", m, pr.OriginalName)
					continue
				}
				if extraEp, ok := matches["extra_ep_num"]; ok {
					extraEpCvt, err := strconv.ParseInt(extraEp, 10, 64)
					if err != nil {
						glog.Errorf("Error converting %s to int from string: %s", extraEp, pr.OriginalName)
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
					glog.Errorf("Error converting %s to int from string: %s", m, pr.OriginalName)
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
	sort.Sort(byScore(matchResults))
	if len(matchResults) > 0 {
		return &matchResults[0], nil
	}
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
