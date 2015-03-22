package db

import (
	"fmt"
	"time"

	"github.com/golang/glog"
)

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
