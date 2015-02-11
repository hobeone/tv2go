package tvdb

import (
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/hobeone/tv2go/db"
	tvd "github.com/nemith/tvdb"
)

func ConvertTvdbEpisodeToDbEpisode(episode tvd.Episode) db.Episode {
	var dbep db.Episode
	dbep.Name = episode.EpisodeName
	dbep.AirDate = episode.FirstAired.UTC()
	dbep.Description = episode.Overview
	dbep.Season = int64(episode.SeasonNumber)
	dbep.Episode = int64(episode.EpisodeNumber)
	if episode.AbsoluteNumber.Valid {
		dbep.AbsoluteNumber = int64(episode.AbsoluteNumber.Value)
	}
	return dbep
}

func GetShowById(tvdbid int64) (*tvd.Series, []tvd.Episode, error) {
	glog.Infof("Getting showid %d from tvdbid", tvdbid)
	t := tvd.NewClient("90D7DF3AE9E4841E")
	series, eps, err := t.SeriesAllByID(int(tvdbid), "en")
	if err != nil {
		return series, eps, err
	}

	return series, eps, nil
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
		Name:  ts.Name,
		Genre: strings.Join(ts.Genre, "|"),
		//Classification: ts.Status,
		Status:    ts.Status,
		StartYear: ts.FirstAired.Year(),
		IndexerID: int64(ts.ID),
		Network:   ts.Network,
		Language:  ts.Language,
	}
	return s
}

func UpdateDBShow(dbshow db.Show, dbeps []db.Episode) (db.Show, []db.Episode, error) {
	ts, eps, err := GetShowById(dbshow.IndexerID)
	if err != nil {
		return dbshow, dbeps, err
	}

	dbshow.Name = ts.Name
	dbshow.Genre = strings.Join(ts.Genre, "|")
	dbshow.Status = ts.Status
	dbshow.StartYear = ts.FirstAired.Year()
	dbshow.Network = ts.Network
	dbshow.LastIndexerUpdate = time.Now()

	for _, episode := range eps {
		glog.Infof("Updating Season %d, Episode %d for '%s (tvdb id: %d)'", episode.SeasonNumber, episode.EpisodeNumber, dbshow.Name, dbshow.IndexerID)
		epToUpdate := db.Episode{}
		for _, dbep := range dbeps {
			if dbep.Season == int64(episode.SeasonNumber) && dbep.Episode == int64(episode.EpisodeNumber) {
				glog.Infof("Found existing episode for Season %d, Episode %d for '%s (tvdb id: %d)'", episode.SeasonNumber, episode.EpisodeNumber, dbshow.Name, dbshow.IndexerID)
				epToUpdate = dbep
			}
		}
		epToUpdate.Name = episode.EpisodeName
		epToUpdate.AirDate = episode.FirstAired.UTC()
		epToUpdate.Description = episode.Overview
		epToUpdate.Season = int64(episode.SeasonNumber)
		epToUpdate.Episode = int64(episode.EpisodeNumber)
		if episode.AbsoluteNumber.Valid {
			epToUpdate.AbsoluteNumber = int64(episode.AbsoluteNumber.Value)
		}
		if epToUpdate.ID == 0 {
			dbeps = append(dbeps, epToUpdate)
		}
	}

	dbshow.Episodes = dbeps
	return dbshow, dbeps, nil
}
