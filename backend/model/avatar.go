package model

import (
	"time"

	"github.com/google/uuid"
)

type Avatar struct {
	UserID      uuid.UUID `gorm:"type:uuid;primaryKey"`
	User        *User     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Data        []byte    `gorm:"type:bytea;not null" json:"-"`
	ContentType string    `gorm:"not null" json:"-"`
	UpdatedAt   time.Time `json:"-"`
}
