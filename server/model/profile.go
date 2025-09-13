package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Profile struct {
	Birthdate time.Time      `json:"birthdate"`
	Language  string         `json:"language"` // ISO 639-1:2002
	Height    float64        `json:"height"`
	Weight    float64        `json:"weight"`
	Data      datatypes.JSON `gorm:"type:jsonb" json:"data"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	User   *User     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

func (p *Profile) Age() int {
	now := time.Now()
	years := now.Year() - p.Birthdate.Year()
	if now.YearDay() < p.Birthdate.YearDay() {
		years--
	}
	return years
}
