package db

import (
	"errors"
	"time"

	"github.com/golang/glog"
	"github.com/hobeone/tv2go/quality"
	"github.com/hobeone/tv2go/types"
	"github.com/jinzhu/gorm"
)

// Episode represents an individual episode of a Show
type Episode struct {
	ID                  int64 `gorm:"column:id; primary_key:yes"`
	Show                Show
	ShowId              int64
	Name                string
	Season              int64
	Episode             int64
	Description         string
	AirDate             time.Time
	HasNFO              bool `gorm:"column:has_nfo"`
	HasTBN              bool `gorm:"column:has_tbn"`
	Status              types.EpisodeStatus
	Quality             quality.Quality
	Location            string
	FileSize            int64
	ReleaseName         string
	SceneSeason         int64
	SceneEpisode        int64
	AbsoluteNumber      int64
	SceneAbsoluteNumber int64
	Version             int64
	ReleaseGroup        string
}

// BeforeSave performs validation on the record before saving
func (e *Episode) BeforeSave() error {
	//if e.Name == "" {
	//	return errors.New("Episode name can not be empty")
	//}
	// Season 0 is used for Specials
	//	if e.Season == 0 {
	//		return errors.New("Season must be set")
	//	}
	if e.Episode == 0 {
		return errors.New("Episode must be set")
	}
	if e.Status == types.UNKNOWN {
		e.Status = types.IGNORED
	}
	return nil
}

// AfterFind fixes the SQLite driver sets everything to local
func (e *Episode) AfterFind() error {
	e.AirDate = e.AirDate.UTC()
	return nil
}

// AirDateString returns the episode's airdate as a YYYY-MM-DD date if set.
// Otherwise it returns the empty string.
func (e *Episode) AirDateString() string {
	if !e.AirDate.IsZero() {
		return e.AirDate.Format("2006-01-02")
	}
	return ""
}

// GetEpisodeByID returns an episode with the given ID or an error if it
// doesn't exist.
func (h *Handle) GetEpisodeByID(episodeid int64) (*Episode, error) {
	var ep Episode
	err := h.db.Preload("Show").Find(&ep, episodeid).Error
	return &ep, err
}

// GetEpisodeByShowSeasonAndNumber does exactly what it says.
func (h *Handle) GetEpisodeByShowSeasonAndNumber(showid, season, number int64) (*Episode, error) {
	var eps []Episode

	err := h.db.Where("show_id = ? and season = ? and episode = ?", showid, season, number).Find(&eps).Error

	if err != nil {
		return nil, err
	}
	if len(eps) == 0 {
		return nil, gorm.RecordNotFound
	}

	return &eps[0], nil
}

// SaveEpisode saves the given episode to the database
func (h *Handle) SaveEpisode(e *Episode) error {
	if h.writeUpdates {
		return h.db.Save(e).Error
	}
	return nil
}

// SaveEpisodes save the list of episodes to the database.  This is done in a
// transaction which will be much faster if the number of episodes is large.
func (h *Handle) SaveEpisodes(eps []*Episode) error {
	if h.writeUpdates {
		tx := h.db.Begin()
		for _, e := range eps {
			err := tx.Save(e).Error
			if err != nil {
				glog.Errorf("Error saving episodes to the database: %s", err.Error())
				tx.Rollback()
				return err
			}
		}
		tx.Commit()
	}
	return nil
}
