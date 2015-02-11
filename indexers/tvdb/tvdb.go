package tvdb

import (
	"strconv"
	"strings"
	"time"

	tvd "github.com/garfunkel/go-tvdb"
	"github.com/golang/glog"
	"github.com/hobeone/tv2go/db"
)

func GetShowById(tvdbid int64) (*tvd.Series, error) {
	glog.Infof("Getting showid %d from tvdbid", tvdbid)
	series, err := tvd.GetSeriesByID(uint64(tvdbid))
	if err != nil {
		glog.Errorf("Error getting show from tvdbid: %s", err.Error())
		return series, err
	}
	return series, nil
}

func parseYear(s string) int {
	parseddate := 0
	t, err := time.Parse("1999-03-28", s)
	if err == nil {
		parseddate = t.Year()
	}
	return parseddate
}
func parseDate(s string) *time.Time {
	parsed, err := time.Parse("1999-03-28", s)
	if err == nil {
		return &time.Time{}
	}
	return &parsed
}
func TVDBToShow(ts *tvd.Series) db.Show {
	s := db.Show{
		Name:  ts.SeriesName,
		Genre: strings.Join(ts.Genre, "|"),
		//Classification: ts.Status,
		Status:    ts.Status,
		StartYear: parseYear(ts.FirstAired),
		IndexerID: int64(ts.ID),
		Network:   ts.Network,
	}
	return s
}

func UpdateDBShow(dbshow db.Show, dbeps []db.Episode) (db.Show, []db.Episode, error) {
	ts, err := GetShowById(dbshow.IndexerID)
	if err != nil {
		return dbshow, dbeps, err
	}
	err = ts.GetDetail()
	if err != nil {
		return dbshow, dbeps, err
	}

	dbshow.Name = ts.SeriesName
	dbshow.Genre = strings.Join(ts.Genre, "|")
	dbshow.Status = ts.Status
	dbshow.StartYear = parseYear(ts.FirstAired)
	dbshow.Network = ts.Network
	dbshow.LastIndexerUpdate = time.Now()

	for seasonnum, seasoneps := range ts.Seasons {
		glog.Infof("Updating Season %d for '%s (tvdb id: %d)'", seasonnum, dbshow.Name, dbshow.IndexerID)
		for _, episode := range seasoneps {
			glog.Infof("Updating Season %d, Episode %d for '%s (tvdb id: %d)'", seasonnum, episode.EpisodeNumber, dbshow.Name, dbshow.IndexerID)
			epToUpdate := db.Episode{}
			for _, dbep := range dbeps {
				if dbep.Season == int64(episode.SeasonNumber) && dbep.Episode == int64(episode.EpisodeNumber) {
					glog.Infof("Found existing episode for Season %d, Episode %d for '%s (tvdb id: %d)'", seasonnum, episode.EpisodeNumber, dbshow.Name, dbshow.IndexerID)
					epToUpdate = dbep
				}
			}
			epToUpdate.Name = episode.EpisodeName
			epToUpdate.AirDate = *parseDate(episode.FirstAired)
			epToUpdate.Description = episode.Overview
			epToUpdate.Season = int64(episode.SeasonNumber)
			epToUpdate.Episode = int64(episode.EpisodeNumber)
			epToUpdate.AbsoluteNumber, _ = strconv.ParseInt(episode.AbsoluteNumber, 10, 64)
			if epToUpdate.ID == 0 {
				dbeps = append(dbeps, epToUpdate)
			}
		}
	}

	dbshow.Episodes = dbeps
	return dbshow, dbeps, nil
}
