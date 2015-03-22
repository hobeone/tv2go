package db

import (
	"time"

	"github.com/golang/glog"
)

// LastPollTime stores the last time we polled a particular
// provider.
type LastPollTime struct {
	Name          string
	LastRefreshed time.Time
}

// AfterFind updates all times to UTC because SQLite driver sets everything to local
func (s *LastPollTime) AfterFind() error {
	s.LastRefreshed = s.LastRefreshed.UTC()
	return nil
}

// SetLastPollTime records the current time for the given name.
func (h *Handle) SetLastPollTime(name string) error {
	dbhistory := &LastPollTime{
		Name: name,
	}

	err := h.db.Where(dbhistory).FirstOrInit(dbhistory).Error
	if err != nil {
		glog.Errorf("Couldn't find or create poll history: %s", err)
	}
	dbhistory.LastRefreshed = time.Now()
	return h.db.Save(dbhistory).Error
}

// GetLastPollTime returns the last time recorded for the given name.  If there
// is no record it returns a zero value time.Time
func (h *Handle) GetLastPollTime(name string) time.Time {
	se := &LastPollTime{}
	err := h.db.Where("name = ?", name).Order("last_refreshed desc").First(se).Error
	if err != nil {
		return time.Time{}
	}
	return se.LastRefreshed
}
