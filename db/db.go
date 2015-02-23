package db

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/hobeone/tv2go/types"
	"github.com/jinzhu/gorm"

	//import sqlite3 driver
	_ "github.com/mattn/go-sqlite3"
)

func init() {
	gorm.NowFunc = func() time.Time {
		return time.Now().UTC()
	}
}

/* From SickRage
*CREATE TABLE tv_shows
  (
     show_id             INTEGER PRIMARY KEY,
     indexer_id          NUMERIC,
     indexer             NUMERIC,
     show_name           TEXT,
     location            TEXT,
     network             TEXT,
     genre               TEXT,
     classification      TEXT,
     runtime             NUMERIC,
     quality             NUMERIC,
     airs                TEXT,
     status              TEXT,
     flatten_folders     NUMERIC,
     paused              NUMERIC,
     startyear           NUMERIC,
     air_by_date         NUMERIC,
     lang                TEXT,
     subtitles           NUMERIC,
     notify_list         TEXT,
     imdb_id             TEXT,
     last_update_indexer NUMERIC,
     dvdorder            NUMERIC,
     archive_firstmatch  NUMERIC,
     rls_require_words   TEXT,
     rls_ignore_words    TEXT,
     sports              NUMERIC,
     anime               NUMERIC,
     scene               NUMERIC,
     default_ep_status   NUMERIC
*/

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
	Quality           types.Quality
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
	DefaultEpStatus   int64 // convert to enum
	LastIndexerUpdate time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (s *Show) BeforeSave() error {
	if s.Name == "" {
		return fmt.Errorf("Name can not be empty")
	}
	if s.IndexerID == 0 {
		return fmt.Errorf("IndexerID can not be unset")
	}
	return nil
}

// SQLite driver sets everything to local
func (s *Show) AfterFind() error {
	s.LastIndexerUpdate = s.LastIndexerUpdate.UTC()
	s.CreatedAt = s.CreatedAt.UTC()
	s.UpdatedAt = s.UpdatedAt.UTC()
	return nil
}

/* Sickrage
CREATE TABLE tv_episodes
  (
     episode_id            INTEGER PRIMARY KEY,
     showid                NUMERIC,
     indexerid             NUMERIC,
     indexer               TEXT,
     name                  TEXT,
     season                NUMERIC,
     episode               NUMERIC,
     description           TEXT,
     airdate               NUMERIC,
     hasnfo                NUMERIC,
     hastbn                NUMERIC,
     status                NUMERIC,
     location              TEXT,
     file_size             NUMERIC,
     release_name          TEXT,
     subtitles             TEXT,
     subtitles_searchcount NUMERIC,
     subtitles_lastsearch  TIMESTAMP,
     is_proper             NUMERIC,
     scene_season          NUMERIC,
     scene_episode         NUMERIC,
     absolute_number       NUMERIC,
     scene_absolute_number NUMERIC,
     version               NUMERIC,
     release_group         TEXT
  );
*/
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
	Quality             types.Quality
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
	if e.Name == "" {
		return errors.New("Name can not be empty")
	}
	//	if e.Season == 0 {
	//		return errors.New("Season must be set")
	//	}
	if e.Episode == 0 {
		return errors.New("Episode must be set")
	}

	if e.Status == types.UNKNOWN {
		return errors.New("Status must be set")
	}

	if e.Quality == 0 {
		return errors.New("Quality must be set")
	}
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

// AfterFind fixes the SQLite driver sets everything to local
func (e *Episode) AfterFind() error {
	e.AirDate = e.AirDate.UTC()
	return nil
}

// Handle controls access to the database and makes sure only one
// operation is in process at a time.
type Handle struct {
	db           gorm.DB
	writeUpdates bool
	syncMutex    sync.Mutex
}

func setupDB(db gorm.DB) error {
	tx := db.Begin()
	err := tx.AutoMigrate(&Show{}, &Episode{}).Error
	tx.Model(&Episode{}).AddIndex(
		"idx_show_season_ep", "show_id", "season", "episode",
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func openDB(dbType string, dbArgs string, verbose bool) gorm.DB {
	glog.Infof("Opening database %s:%s", dbType, dbArgs)
	// Error only returns from this if it is an unknown driver.
	d, err := gorm.Open(dbType, dbArgs)
	if err != nil {
		panic(err.Error())
	}
	d.SingularTable(true)
	d.LogMode(verbose)
	// Actually test that we have a working connection
	err = d.DB().Ping()
	if err != nil {
		panic(err.Error())
	}
	return d
}

func createAndOpenDb(dbPath string, verbose bool, memory bool) *Handle {
	mode := "rwc"
	if memory {
		mode = "memory"
	}
	constructedPath := fmt.Sprintf("file:%s?mode=%s", dbPath, mode)
	db := openDB("sqlite3", constructedPath, verbose)
	err := setupDB(db)
	if err != nil {
		panic(err.Error())
	}
	return &Handle{db: db}
}

// NewDBHandle creates a new DBHandle
//	dbPath: the path to the database to use.
//	verbose: when true database accesses are logged to stdout
//	writeUpdates: when true actually write to the databse (useful for testing)
func NewDBHandle(dbPath string, verbose bool, writeUpdates bool) *Handle {
	d := createAndOpenDb(dbPath, verbose, false)
	d.writeUpdates = writeUpdates
	return d
}

// NewMemoryDBHandle creates a new in memory database.  Useful for testing.
func NewMemoryDBHandle(verbose bool, writeUpdates bool) *Handle {
	d := createAndOpenDb("in_memory_test", verbose, true)
	d.writeUpdates = writeUpdates
	return d
}

func (h *Handle) AddShow(s *Show) error {
	return h.db.Create(s).Error
}

func (h *Handle) DB() *gorm.DB {
	return &h.db
}

func (h *Handle) GetAllShows() ([]Show, error) {
	var shows []Show
	err := h.db.Find(&shows).Error
	return shows, err
}

func (h *Handle) GetShowEpisodes(s *Show) ([]Episode, error) {
	var episodes []Episode
	err := h.db.Model(s).Related(&episodes).Error
	return episodes, err
}

func (h *Handle) GetEpisodeByID(episodeid int64) (*Episode, error) {
	var ep Episode

	err := h.db.Find(&ep, episodeid).Error
	return &ep, err
}

func (h *Handle) GetEpisodeByShowSeasonAndNumber(showid, season, number int64) (*Episode, error) {
	var eps []Episode

	err := h.db.Where("show_id = ? and season = ? and episode = ?", showid, season, number).Find(&eps).Error

	if err != nil || len(eps) == 0 {
		return nil, err
	}

	return &eps[0], nil
}

func (h *Handle) GetShowSeason(showid, season int64) ([]Episode, error) {
	var episodes []Episode
	show, err := h.GetShowById(showid)
	if err != nil {
		return episodes, err
	}
	err = h.db.Where("show_id = ? AND season = ?", show.ID, season).Find(&episodes).Error
	return episodes, err
}

func (h *Handle) GetShowById(showID int64) (*Show, error) {
	var show Show

	err := h.db.Find(&show, showID).Error
	if err != nil {
		return nil, err
	}
	err = h.db.Model(&show).Related(&show.Episodes).Error
	if err != nil {
		return nil, err
	}
	return &show, err
}

func (h *Handle) SaveShow(s *Show) error {
	if h.writeUpdates {
		return h.db.Save(s).Error
	}
	return nil
}

func (h *Handle) SaveEpisode(e *Episode) error {
	if h.writeUpdates {
		return h.db.Save(e).Error
	}
	return nil
}

func (h *Handle) SaveEpisodes(eps []*Episode) error {
	if h.writeUpdates {
		tx := h.db.Begin()
		for _, e := range eps {
			defer tx.Rollback()
			return tx.Save(e).Error
		}
		tx.Commit()
	}
	return nil
}

// Testing functionality

// TestReporter is a shim interface so we don't need to include the testing
// package in the compiled binary
type TestReporter interface {
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

// LoadFixtures adds a base set of Fixtures to the given database.
func LoadFixtures(t TestReporter, d *Handle) []Show {
	shows := []Show{
		{
			Name:      "show1",
			IndexerID: 1,
			Location:  "testdata",
			Episodes: []Episode{
				{
					Name:    "show1episode1",
					Season:  1,
					Episode: 1,
					AirDate: time.Date(2006, time.January, 1, 0, 0, 0, 0, time.UTC),
					Status:  types.WANTED,
					Quality: types.NONE,
				},
				{
					Name:    "show1episode2",
					Season:  1,
					Episode: 2,
					Status:  types.WANTED,
					Quality: types.NONE,
				},
			},
		},
		{
			Name:      "show2",
			IndexerID: 2,
			Location:  "testdata",
			Episodes: []Episode{
				{
					Name:    "show2episode1",
					Season:  1,
					Episode: 1,
					AirDate: time.Date(2001, time.January, 1, 0, 0, 0, 0, time.UTC),
					Status:  types.WANTED,
					Quality: types.NONE,
				},
				{
					Name:    "show2episode2",
					Season:  2,
					Episode: 1,
					AirDate: time.Date(2002, time.February, 1, 0, 0, 0, 0, time.UTC),
					Status:  types.WANTED,
					Quality: types.NONE,
				},
			},
		},
	}
	for _, s := range shows {
		err := d.SaveShow(&s)
		if err != nil {
			t.Fatalf("Error saving show fixture to db: %s", err)
		}
	}
	return shows
}
