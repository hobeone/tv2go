package db

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/hobeone/tv2go/quality"
	"github.com/hobeone/tv2go/types"
)

// Show is a TV Show
type Show struct {
	ID                int64  `gorm:"column:id; primary_key:yes"`
	Name              string `sql:"not null"`
	Description       string
	Indexer           string `sql:"not null"`
	IndexerID         int64  `gorm:"column:indexer_key"` // id to use when looking up with the indexer
	Episodes          []Episode
	Location          string // Location of Show on disk
	Network           string
	Genre             string // pipe seperated
	Classification    string
	Runtime           int64 // in minutes
	QualityGroup      quality.QualityGroup
	QualityGroupID    int64
	Airs              string // Hour of the day
	Status            string
	FlattenFolders    bool
	Paused            bool
	StartYear         int
	AirByDate         bool
	Language          string
	Subtitles         bool
	ImdbID            string
	Sports            bool
	Anime             bool
	Scene             bool
	DefaultEpStatus   types.EpisodeStatus
	LastIndexerUpdate time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// BeforeSave validates a show before writing it to the database
func (s *Show) BeforeSave() error {
	if s.Name == "" {
		return fmt.Errorf("Show Name can not be empty")
	}
	if s.IndexerID == 0 {
		return fmt.Errorf("IndexerID can not be unset")
	}
	if s.Indexer == "" {
		return fmt.Errorf("Indexer must be set")
	}
	if s.DefaultEpStatus == types.UNKNOWN {
		s.DefaultEpStatus = types.IGNORED
	}
	return nil
}

// AfterFind updates all times to UTC because SQLite driver sets everything to local
func (s *Show) AfterFind() error {
	s.LastIndexerUpdate = s.LastIndexerUpdate.UTC()
	s.CreatedAt = s.CreatedAt.UTC()
	s.UpdatedAt = s.UpdatedAt.UTC()
	if s.DefaultEpStatus == types.UNKNOWN {
		s.DefaultEpStatus = types.IGNORED
	}
	return nil
}

// NextAirdateForShow returns the date of the next episode for this show.
func (h *Handle) NextAirdateForShow(dbshow *Show) *time.Time {
	var ep Episode
	err := h.db.Where("show_id = ? and air_date >= ? and status IN (?,?)", dbshow.ID, time.Now(), types.WANTED, types.UNAIRED).Order("air_date asc").First(&ep).Error

	if err != nil {
		return nil
	}
	return &ep.AirDate
}

// AddShow adds the given show to the Database
func (h *Handle) AddShow(s *Show) error {
	return h.db.Create(s).Error
}

// SaveShow saves the show (and any episodes) to the database
func (h *Handle) SaveShow(s *Show) error {
	if h.writeUpdates {
		return h.db.Save(s).Error
	}
	return nil
}

// GetAllShows returns all shows in the database.
func (h *Handle) GetAllShows() ([]Show, error) {
	var shows []Show
	err := h.db.Preload("QualityGroup").Find(&shows).Error
	return shows, err
}

// GetShowByID returs the show with the given ID or an error if it doesn't
// exist.
func (h *Handle) GetShowByID(showID int64) (*Show, error) {
	var show Show

	err := h.db.Preload("QualityGroup").Find(&show, showID).Error
	if err != nil {
		return nil, err
	}
	err = h.db.Model(&show).Related(&show.Episodes).Error
	if err != nil {
		return nil, err
	}
	return &show, err
}

// GetShowByName returns the show with the given name (and it's episodes) or an
// error if not found.
func (h *Handle) GetShowByName(name string) (*Show, error) {
	var show Show
	err := h.db.Preload("Episodes").Preload("QualityGroup").Where("name = ?", name).Find(&show).Error
	return &show, err
}

// GetShowByNameIgnoreCase returns the show with the given name (and it's episodes) or an
// error if not found.
func (h *Handle) GetShowByNameIgnoreCase(name string) (*Show, error) {
	var show Show
	err := h.db.Preload("Episodes").Preload("QualityGroup").Where("name = ? COLLATE NOCASE", name).Find(&show).Error
	return &show, err
}

//GetShowByAllNames tries to find a show that matches the given name using
//alternate spelling and associated show names from various SceneException
//providers.
//
//It returns the Show, a season override if that given name maps to a
//particular season of the show (common for Anime) and an error if one occured.
func (h *Handle) GetShowByAllNames(name string) (*Show, int64, error) {
	glog.Infof("Trying to match provider result %s", name)

	glog.Infof("Trying to find an exact match in the database for %s", name)
	dbshow, err := h.GetShowByName(name)
	if err == nil {
		glog.Infof("Matched name %s to show %s", name, dbshow.Name)
		return dbshow, -1, nil
	}
	glog.Infof("Couldn't find show with exact name %s in database.", name)

	dbshow, season, err := h.GetShowFromNameException(name)
	if err == nil {
		glog.Infof("Matched provider result %s to show %s", name, dbshow.Name)
		return dbshow, season, nil
	}
	glog.Infof("Couldn't find a match scene name %s", name)

	return nil, -1, fmt.Errorf("Couldn't find a match for show %s", name)
}

// GetShowByIndexerAndID returns the show with the given indexer and indexerid
// or an error if it doesn't exist.
func (h *Handle) GetShowByIndexerAndID(indexer string, indexerID int64) (*Show, error) {
	var show Show

	err := h.db.Preload("QualityGroup").Where("indexer = ? AND indexer_key = ?", indexer, indexerID).Find(&show).Error
	if err != nil {
		return nil, err
	}
	err = h.db.Model(&show).Related(&show.Episodes).Error
	if err != nil {
		return nil, err
	}
	return &show, err
}

// GetShowEpisodes returns all of the given show's episodes
func (h *Handle) GetShowEpisodes(s *Show) ([]Episode, error) {
	var episodes []Episode
	err := h.db.Model(s).Related(&episodes).Error
	return episodes, err
}
