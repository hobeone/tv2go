package quality

import (
	"strconv"
	"strings"

	"github.com/golang/glog"
)

// QualityGroup represents a group of acceptable qualities for a Show.
type QualityGroup struct {
	ID            int64     `json:"-"`
	Name          string    `json:"name"`
	Qualities     []Quality `sql:"-" json:"qualities"`
	QualityString string    `json:"-"` // CSV of ints
	Default       bool      `json:"default"`
}

// Includes returns true if the given Quality is included in the QualityGroup
func (qg QualityGroup) Includes(qual Quality) bool {
	for _, q := range qg.Qualities {
		if q == qual {
			return true
		}
	}
	return false
}

// Pretty hacky, but seems to work.  Qualities are constants and not DB
// objects, so we need glue to keep them together.  Groups of Qualities can be
// user defined and so fit that role.
//
// We serialize Qualities to a CSV of ints and then reconstitute them when
// loaded from the DB.
func (qg *QualityGroup) BeforeSave() error {
	strs := []string{}
	for _, value := range qg.Qualities {
		if value.String() == "" {
			glog.Warningf("Unknown Quality: '%v'", value)
			continue
		}
		strs = append(strs, strconv.FormatInt(int64(value), 10))
	}
	qg.QualityString = strings.Join(strs, ",")
	return nil
}

func (qg *QualityGroup) AfterFind() error {
	strs := strings.Split(qg.QualityString, ",")
	qg.Qualities = []Quality{}
	for _, s := range strs {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			continue
		}
		qual, err := QualityFromInt(i)
		if err != nil {
			glog.Warningf("Unknown Quality from DB: '%v'", i)
			continue
		}
		qg.Qualities = append(qg.Qualities, qual)
	}
	return nil
}

var DefaultQualityGroup = QualityGroup{
	Name:      "DEFAULT_BUILTIN",
	Qualities: ALL_HD_QUALITIES,
}
