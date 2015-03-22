package db

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/hobeone/tv2go/naming"
)

// NameException stores alternate names of shows to use when parsing input files.
type NameException struct {
	ID        int64
	Source    string
	Indexer   string
	IndexerID int64
	Name      string
	Season    int64
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

//GetShowFromNameException will try to match (ignoring case) the given name to
//a known Name Exception.  If found it will then try to match that to a Show by
//matching the indexer and indexerid.
func (h *Handle) GetShowFromNameException(name string) (*Show, int64, error) {
	ne := &[]NameException{}
	err := h.db.Where("name = ? COLLATE NOCASE", name).Find(ne).Error
	if err != nil {
		ne = &[]NameException{}
		sceneName := naming.FullSanitizeSceneName(name)
		glog.Infof("searching for name '%s' with scene name '%s'", name, sceneName)

		err = h.db.Where("name = ? COLLATE NOCASE", name).Find(ne).Error
		if err != nil {
			return nil, -1, err
		}
	}
	for _, exp := range *ne {
		show, err := h.GetShowByIndexerAndID(exp.Indexer, exp.IndexerID)
		if err == nil {
			return show, exp.Season, nil
		}
	}
	return nil, -1, fmt.Errorf("Couldn't find matching show for %s", name)
}

//SaveNameExceptions saves all the given exceptions, totally replacing the
//exceptions for the given source.
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
