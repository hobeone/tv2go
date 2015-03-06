package db

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/hobeone/tv2go/quality"
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
	return nil
}

// AfterFind updates all times to UTC because SQLite driver sets everything to local
func (s *Show) AfterFind() error {
	s.LastIndexerUpdate = s.LastIndexerUpdate.UTC()
	s.CreatedAt = s.CreatedAt.UTC()
	s.UpdatedAt = s.UpdatedAt.UTC()
	return nil
}

func (h *Handle) NextAirdateForShow(dbshow *Show) *time.Time {
	var ep Episode
	err := h.db.Where("show_id = ? and air_date >= ? and status IN (?,?)", dbshow.ID, time.Now(), types.WANTED, types.UNAIRED).Order("air_date asc").First(&ep).Error

	if err != nil {
		return nil
	}
	return &ep.AirDate
}

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
		return errors.New("Status must be set")
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
	err := tx.AutoMigrate(&Show{}, &Episode{}, &quality.QualityGroup{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Don't check errors on these because they'll usually already exist
	tx.Model(&Episode{}).AddIndex(
		"idx_show_season_ep", "show_id", "season", "episode",
	)
	tx.Model(&Show{}).AddUniqueIndex("idx_show_name", "name")
	tx.Model(&quality.QualityGroup{}).AddUniqueIndex("idx_quality_group_name", "name")
	tx.Commit()
	RunMigrations(&db)
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

// AddShow adds the given show to the Database
func (h *Handle) AddShow(s *Show) error {
	return h.db.Create(s).Error
}

// GetAllShows returns all shows in the database.
func (h *Handle) GetAllShows() ([]Show, error) {
	var shows []Show
	err := h.db.Preload("QualityGroup").Find(&shows).Error
	return shows, err
}

// GetShowEpisodes returns all of the given show's episodes
func (h *Handle) GetShowEpisodes(s *Show) ([]Episode, error) {
	var episodes []Episode
	err := h.db.Model(s).Related(&episodes).Error
	return episodes, err
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

	if err != nil || len(eps) == 0 {
		return nil, err
	}

	return &eps[0], nil
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

// SaveShow saves the show (and any episodes) to the database
func (h *Handle) SaveShow(s *Show) error {
	if h.writeUpdates {
		return h.db.Save(s).Error
	}
	return nil
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

func (h *Handle) GetQualityGroupFromStringOrDefault(name string) *quality.QualityGroup {
	qual := &quality.QualityGroup{}
	err := h.db.Where("name = ?", name).Find(qual).Error
	if err == nil {
		return qual
	}
	err = h.db.Where("default = ?", true).Find(qual).Error
	if err == nil {
		return qual
	}
	h.db.FirstOrInit(qual, quality.DefaultQualityGroup)
	return qual
}

// Migrations

type DBVersion struct {
	ID        int64
	Version   int64
	UpdatedAt time.Time
	CreatedAt time.Time
}

func RunMigrations(dbh *gorm.DB) {
	tx := dbh.Begin()
	err := Migration_001_AddBaseData(tx)
	if err != nil {
		tx.Rollback()
	}
	tx.Commit()
}

func Migration_001_AddBaseData(tx *gorm.DB) error {
	baseQualities := []quality.QualityGroup{
		quality.QualityGroup{
			Name:    "HDALL",
			Default: true,
			Qualities: []quality.Quality{
				quality.HDTV,
				quality.HDWEBDL,
				quality.HDBLURAY,
				quality.FULLHDWEBDL,
				quality.FULLHDTV,
				quality.FULLHDBLURAY,
			},
		},
	}

	for _, qg := range baseQualities {
		err := tx.Save(&qg).Error
		if err != nil {
			return err
		}
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
	basedir, err := filepath.Abs("")
	if err != nil {
		t.Fatalf("Error finding base directory: %s", err.Error())
	}
	shows := []Show{
		{
			Name:            "show1",
			IndexerID:       1,
			Location:        basedir + "/testdata/show1",
			DefaultEpStatus: types.WANTED,
			QualityGroupID:  1,
			Episodes: []Episode{
				{
					Name:    "show1episode1",
					Season:  1,
					Episode: 1,
					AirDate: time.Date(2006, time.January, 1, 0, 0, 0, 0, time.UTC),
					Status:  types.WANTED,
					Quality: quality.UNKNOWN,
				},
				{
					Name:    "show1episode2",
					Season:  1,
					Episode: 2,
					Status:  types.WANTED,
					Quality: quality.UNKNOWN,
				},
			},
		},
		{
			Name:      "show2",
			IndexerID: 2,
			Location:  basedir + "/testdata/show2",
			Episodes: []Episode{
				{
					Name:     "show2episode1",
					Season:   1,
					Episode:  1,
					AirDate:  time.Date(2001, time.January, 1, 0, 0, 0, 0, time.UTC),
					Status:   types.WANTED,
					Quality:  quality.UNKNOWN,
					Location: "testdata/show1",
				},
				{
					Name:    "show2episode2",
					Season:  2,
					Episode: 1,
					AirDate: time.Date(2002, time.February, 1, 0, 0, 0, 0, time.UTC),
					Status:  types.WANTED,
					Quality: quality.UNKNOWN,
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
