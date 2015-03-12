package quality

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type Quality int64

// Episode Quality Enum
const (
	UNKNOWN      Quality = 0
	SDTV         Quality = 10
	SDDVD        Quality = 100
	HDTV         Quality = 200
	RAWHDTV      Quality = 300
	FULLHDTV     Quality = 400
	HDWEBDL      Quality = 500
	FULLHDWEBDL  Quality = 600
	HDBLURAY     Quality = 700
	FULLHDBLURAY Quality = 800
)

var ALL_HD_QUALITIES = []Quality{
	HDTV,
	RAWHDTV,
	FULLHDTV,
	HDWEBDL,
	FULLHDWEBDL,
	HDBLURAY,
	FULLHDBLURAY,
}

var qualities = map[string]Quality{
	"Unknown":      UNKNOWN,
	"SD TV":        SDTV,
	"SD DVD":       SDDVD,
	"HD TV":        HDTV,
	"RawHD TV":     RAWHDTV,
	"1080p HD TV":  FULLHDTV,
	"720p WEB-DL":  HDWEBDL,
	"1080p WEB-DL": FULLHDWEBDL,
	"720p BluRay":  HDBLURAY,
	"1080p BluRay": FULLHDBLURAY,
}

func (q Quality) MarshalJSON() ([]byte, error) {
	return json.Marshal(q.String())
}

// Scan implements the sql.Scanner interface
func (q *Quality) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
	case int64:
		*q = Quality(s)
	default:
		return errors.New("Cannot scan Quality from " + reflect.ValueOf(src).String())
	}
	return nil
}

// String() function will return the english quality name
func (quality Quality) String() string {
	for k, v := range qualities {
		if v == quality {
			return k
		}
	}
	return ""
}

func QualityFromString(s string) (Quality, error) {
	if val, ok := qualities[s]; ok {
		return val, nil
	}

	return UNKNOWN, fmt.Errorf("Unknown Quality String: %s", s)
}

func QualityFromInt(i int64) (Quality, error) {
	for _, q := range qualities {
		if i == int64(q) {
			return q, nil
		}
	}
	return UNKNOWN, fmt.Errorf("'%d' doesn't map to a known Quality", i)
}

func QualityFromName(name string, anime bool) Quality {
	// Search for exact match in a file string:
	for _, qual := range qualities {
		regexStr := strings.Replace(qual.String(), " ", `\W`, -1)
		regexStr = `\W` + regexStr + `($|[\W])` // Either non-word or end of line
		regex := regexp.MustCompile(regexStr)
		if regex.MatchString(name) {
			return qual
		}
	}
	if anime {
		return guessAnimeQualityFromName(name)
	}
	return guessQualityFromName(name)
}

func checkName(name string, regexStrings ...string) bool {
	for _, regexStr := range regexStrings {
		regex := regexp.MustCompile(`(?i)` + regexStr)
		if regex.MatchString(name) {
			return true
		}
	}
	return false
}

func checkNameAll(name string, regexStrings ...string) bool {
	for _, regexStr := range regexStrings {
		regex := regexp.MustCompile(`(?i)` + regexStr)
		if !regex.MatchString(name) {
			return false
		}
	}
	return true
}

func guessAnimeQualityFromName(name string) Quality {
	dvd := checkName(name, "dvd", "dvdrip")
	bluray := checkName(name, "bluray", "blu-ray", "BD")
	sdOptions := checkName(name, "360p", "480p", "848x480", "XviD")
	hdOptions := checkName(name, "720p", "1280x720", "960x720")
	fullHD := checkName(name, "1080p", "1920x1080")
	if sdOptions && !bluray && !dvd {
		return SDTV
	} else if dvd {
		return SDDVD
	} else if hdOptions && !bluray && !fullHD {
		return HDTV
	} else if fullHD && !bluray && !hdOptions {
		return FULLHDTV
	} else if hdOptions && !bluray && !fullHD {
		return HDWEBDL
	} else if bluray && hdOptions && !fullHD {
		return HDBLURAY
	} else if bluray && fullHD && !hdOptions {
		return FULLHDBLURAY
	}
	return UNKNOWN
}

// Copied from Sickbeard/Rage
func guessQualityFromName(name string) Quality {
	if checkNameAll(name, "(pdtv|hdtv|dsr|tvrip).(xvid|x264|h.?264)") && !checkNameAll(name, "(720|1080)[pi]") && !checkName(name, "hr.ws.pdtv.x264") {
		return SDTV
	} else if checkNameAll(name, "web.dl|webrip", "xvid|x264|h.?264") && !checkNameAll(name, "(720|1080)[pi]") {
		return SDTV
	} else if checkName(name, "(dvdrip|b[r|d]rip)(.ws)?.(xvid|divx|x264)") && !checkNameAll(name, "(720|1080)[pi]") {
		return SDDVD
	} else if (checkNameAll(name, "720p", "hdtv", "[hx]264") || checkName(name, "hr.ws.pdtv.x264")) && !checkNameAll(name, "(1080)[pi]") {
		return HDTV
	} else if checkNameAll(name, "720p|1080i", "hdtv", "mpeg-?2") || checkNameAll(name, "1080[pi].hdtv", "h.?264") {
		return RAWHDTV
	} else if checkNameAll(name, "1080p", "hdtv", "x264") {
		return FULLHDTV
	} else if checkNameAll(name, "720p", "web.dl|webrip") || checkNameAll(name, "720p", "itunes", "h.?264") {
		return HDWEBDL
	} else if checkNameAll(name, "1080p", "web.dl|webrip") || checkNameAll(name, "1080p", "itunes", "h.?264") {
		return FULLHDWEBDL
	} else if checkNameAll(name, "720p", "bluray|hddvd|b[r|d]rip", "x264") {
		return HDBLURAY
	} else if checkNameAll(name, "1080p", "bluray|hddvd|b[r|d]rip", "x264") {
		return FULLHDBLURAY
	}
	return UNKNOWN
}
