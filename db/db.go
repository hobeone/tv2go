package db

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/hobeone/tv2go/naming"
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

// NameException stores alternate names of shows to use when parsing input files.
type NameException struct {
	ID        int64
	Source    string
	Indexer   string
	IndexerID int64
	Name      string
	Custom    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

//BeforeSave ensures that all fields are set to non default values.
func (e *NameException) BeforeSave() error {
	if e.Indexer == "" {
		return fmt.Errorf("NameException Indexer can't be blank")
	}
	if e.Source == "" {
		return fmt.Errorf("NameException Source can't be blank")
	}
	if e.IndexerID == 0 {
		return fmt.Errorf("NameException IndexerID can't be blank")
	}
	if e.Name == "" {
		return fmt.Errorf("NameException Name can't be blank")
	}
	return nil
}

// XEMException maps alternate names to season numbers.
type XEMException struct {
	ID        int64
	Indexer   string
	IndexerID int64
	Name      string
	Season    int64
}

func (x *XEMException) BeforeSave() error {
	if x.Indexer == "" {
		return fmt.Errorf("XEMException Indexer can't be blank")
	}
	if x.IndexerID == 0 {
		return fmt.Errorf("XEMException IndexerID can't be blank")
	}
	if x.Name == "" {
		return fmt.Errorf("XEMException Name can't be blank")
	}
	return nil
}

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
		e.Status = types.IGNORED
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
	err := tx.AutoMigrate(&Show{}, &Episode{}, &quality.QualityGroup{},
		&NameException{}, &LastPollTime{}, &XEMException{}).Error
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

type logBridge struct{}

func (l logBridge) Print(v ...interface{}) {
	strs := make([]string, len(v))
	for i, val := range v {
		strs[i] = fmt.Sprintf("%v", val)
	}
	glog.Info(strings.Join(strs, " "))
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
	d.SetLogger(logBridge{})
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
	constructedPath := fmt.Sprintf("file:%s?mode=%s&loc=UTC", dbPath, mode)
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

	if err != nil {
		return nil, err
	}
	if len(eps) == 0 {
		return nil, gorm.RecordNotFound
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

// GetShowByIndexerAndID returns the show with the given indexer and indexerid or an error if it doesn't
// exist.
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

func (h *Handle) GetQualityGroups() ([]quality.QualityGroup, error) {
	groups := []quality.QualityGroup{}
	err := h.db.Find(&groups).Error
	if err != nil {
		return groups, err
	}
	return groups, nil
}

// GetQualityGroupFromStringOrDefault tries to find a matching QualityGroup
// with the given name.  If that doesn't exist it returns the first one with
// the Default bit set.  If _that_ fails it will return (and create inthe db)
// the hardcoded default.
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

func (h *Handle) GetLastPollTime(name string) time.Time {
	se := &LastPollTime{}
	err := h.db.Where("name = ?", name).Order("last_refreshed desc").First(se).Error
	if err != nil {
		return time.Time{}
	}
	return se.LastRefreshed
}

func (h *Handle) GetShowAndSeasonFromXEMName(name string) (*Show, int64, error) {
	xem := &XEMException{}
	err := h.db.Where("name = ? COLLATE NOCASE", name).Find(xem).Error
	if err != nil {
		return nil, -1, err
	}

	show, err := h.GetShowByIndexerAndID(xem.Indexer, xem.IndexerID)
	if err != nil {
		return nil, -1, err
	}
	return show, xem.Season, nil

}

func (h *Handle) GetShowFromNameException(name string) (*Show, error) {

	ne := &NameException{}
	err := h.db.Where("name = ? COLLATE NOCASE", name).Find(ne).Error
	if err != nil {
		return nil, err
	}
	show, err := h.GetShowByIndexerAndID(ne.Indexer, ne.IndexerID)
	if err != nil {
		return nil, err
	}
	return show, nil
}

func (h *Handle) SaveNameExceptions(source string, excepts []*NameException) error {
	if h.writeUpdates {
		tx := h.db.Begin()
		err := tx.Where("source = ?", source).Delete(NameException{}).Error
		if err != nil {
			glog.Errorf("Couldn't delete old name exceptions for %s: %s", source, err)
			tx.Rollback()
			return err
		}
		for _, e := range excepts {
			err := tx.Save(e).Error
			if err != nil {
				glog.Errorf("Error saving exceptions to the database: %s", err.Error())
				tx.Rollback()
				return err
			}
		}
		tx.Commit()
	}
	return nil
}

// SaveXEMException saves a list of exceptions for the given indexer,
// overwriting all exceptions for that indexer.
func (h *Handle) SaveXEMExceptions(indexer string, excepts []*XEMException) error {
	if h.writeUpdates {
		tx := h.db.Begin()
		err := tx.Where("indexer = ?", indexer).Delete(XEMException{}).Error
		if err != nil {
			glog.Errorf("Couldn't delete old XEM exceptions for %s: %s", indexer, err)
			tx.Rollback()
			return err
		}
		for _, e := range excepts {
			err := tx.Save(e).Error
			if err != nil {
				glog.Errorf("Error saving exceptions to the database: %s", err.Error())
				tx.Rollback()
				return err
			}
		}
		tx.Commit()
	}
	return nil
}

func (h *Handle) GetShowByAllNames(name string) (*Show, int64, error) {
	glog.Infof("Trying to match provider result %s", name)

	glog.Infof("Trying to find an exact match in the database for %s", name)
	dbshow, err := h.GetShowByName(name)
	if err == nil {
		glog.Infof("Matched name %s to show %s", name, dbshow.Name)
		return dbshow, -1, nil
	}
	glog.Infof("Couldn't find show with exact name %s in database.", name)

	dbshow, season, err := h.GetShowAndSeasonFromXEMName(name)
	if err == nil {
		glog.Infof("Matched name %s to show %s", name, dbshow.Name)
		return dbshow, season, nil
	}
	glog.Info("Couldn't find show with XEM Exception in database.")

	sceneName := naming.FullSanitizeSceneName(name)
	glog.Infof("Converting name '%s' to scene name '%s'", name, sceneName)

	dbshow, err = h.GetShowFromNameException(sceneName)
	if err == nil {
		glog.Infof("Matched provider result %s to show %s", sceneName, dbshow.Name)
		return dbshow, -1, nil
	}
	glog.Infof("Couldn't find a match scene name %s", sceneName)

	return nil, -1, fmt.Errorf("Couldn't find a match for show %s", name)
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
		{
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
		{
			Name: "SD",
			Qualities: []quality.Quality{
				quality.SDTV,
				quality.SDDVD,
			},
		},
		{
			Name: "HD720p",
			Qualities: []quality.Quality{
				quality.HDWEBDL,
				quality.HDBLURAY,
				quality.HDTV,
			},
		},
		{
			Name: "HD1080p",
			Qualities: []quality.Quality{
				quality.FULLHDTV,
				quality.FULLHDWEBDL,
				quality.FULLHDBLURAY,
			},
		},
		{
			Name: "ALL",
			Qualities: []quality.Quality{
				quality.SDTV,
				quality.SDDVD,
				quality.HDTV,
				quality.RAWHDTV,
				quality.FULLHDTV,
				quality.HDWEBDL,
				quality.FULLHDWEBDL,
				quality.HDBLURAY,
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
			Indexer:         "tvdb",
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
			Indexer:   "tvdb",
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
					Name:           "show2episode2",
					Season:         2,
					Episode:        1,
					AbsoluteNumber: 2,
					AirDate:        time.Date(2002, time.February, 1, 0, 0, 0, 0, time.UTC),
					Status:         types.WANTED,
					Quality:        quality.UNKNOWN,
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
