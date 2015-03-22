package db

import (
	"fmt"
	"path/filepath"
	"strings"
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

// Handle controls access to the database and makes sure only one
// operation is in process at a time.
type Handle struct {
	db           gorm.DB
	writeUpdates bool
	syncMutex    sync.Mutex
}

func setupDB(db gorm.DB) error {
	tx := db.Begin()
	err := tx.AutoMigrate(
		&Show{},
		&Episode{},
		&quality.QualityGroup{},
		&NameException{},
		&LastPollTime{},
	).Error
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

// Migrations

// RunMigrations will run all migrations.
//
// TODO: make this a little more sophisticated.
func RunMigrations(dbh *gorm.DB) {
	tx := dbh.Begin()
	err := Migration001AddBaseData(tx)
	if err != nil {
		tx.Rollback()
	}
	tx.Commit()
}

// Migration001AddBaseData adds the baseline quality groups.
func Migration001AddBaseData(tx *gorm.DB) error {
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
