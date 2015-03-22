package db

import "github.com/hobeone/tv2go/quality"

// GetQualityGroups returns all quality groups.
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
