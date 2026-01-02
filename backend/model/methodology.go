package model

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
)

// MethodologyWork defines score constraints for a movement family within a methodology.
type MethodologyWork struct {
	Min int `json:"min"`
	Max int `json:"max,omitempty"` // 0 = no upper limit (use capability)
}

// Methodology defines a training methodology with associated movement families and score ranges.
type Methodology struct {
	ID          string         `gorm:"type:varchar(64);primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(128);not null" json:"name"`
	Description string         `gorm:"type:text;not null" json:"description"`
	Work        datatypes.JSON `gorm:"type:jsonb;column:work;not null" json:"-"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// GetWork returns the work families map from JSONB field.
func (m *Methodology) GetWork() map[string]MethodologyWork {
	var work map[string]MethodologyWork
	if err := json.Unmarshal(m.Work, &work); err != nil {
		return nil
	}
	return work
}

// SetWork serializes the work map to JSONB for storage.
func (m *Methodology) SetWork(work map[string]MethodologyWork) error {
	data, err := json.Marshal(work)
	if err != nil {
		return err
	}
	m.Work = datatypes.JSON(data)
	return nil
}
