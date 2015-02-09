package db

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/jinzhu/gorm"

	//import sqlite3 driver
	_ "github.com/mattn/go-sqlite3"
)

// Type system:
// AnimeShow
// AnimeEpisode
// TvShow
// TvEpisode
// Movie
// MusicAlbum
// MusicTrack
// FanArt

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

type Indexer struct {
	ID   int64 `gorm:"column:id; primary_key:yes"`
	Name string
}

// Show is a TV Show
type Show struct {
	ID              int64  `gorm:"column:id; primary_key:yes"`
	Name            string `sql:"not null"`
	Indexer         Indexer
	Location        string
	Network         string
	Genre           string
	Classification  string
	Runtime         int64
	Quality         int64 // convert to foreign key
	Airs            string
	Status          string
	FlattenFolders  bool
	Paused          bool
	StartYear       int
	AirByDate       bool
	Lang            string
	Subtitles       bool
	ImdbID          string
	Sports          bool
	Anime           bool
	Scene           bool
	DefaultEpStatus int64 // convert to enum
	CreatedAt       time.Time
	UpdatedAt       time.Time
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
	ShowID              Show
	Name                string
	Season              int64
	Episode             int64
	Description         string
	Airdate             int64
	HasNFO              bool
	HasTBN              bool
	Status              int64
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

// Handle controls access to the database and makes sure only one
// operation is in process at a time.
type Handle struct {
	DB           gorm.DB
	writeUpdates bool
	syncMutex    sync.Mutex
}

func setupDB(db gorm.DB) error {
	tx := db.Begin()
	err := tx.AutoMigrate(&Show{}, &Episode{}).Error
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
	return &Handle{DB: db}
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
	return nil
}
