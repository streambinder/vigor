package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Profile stores user's physical attributes and fitness-related metadata.
type Profile struct {
	Birthdate time.Time      `json:"birthdate" flutter:"required"`
	Language  string         `json:"language" flutter:"required"` // ISO 639-1:2002
	Height    float64        `json:"height" flutter:"required"`
	Weight    float64        `json:"weight" flutter:"required"`
	Data      datatypes.JSON `gorm:"type:jsonb" json:"data"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	User   *User     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

type profileData struct {
	Goals       []Goal   `json:"goals" flutter:"required"`
	Injuries    []Injury `json:"injuries" flutter:"required"`
	Limitations []string `json:"limitations" flutter:"required"`
}

// Goal represents a user's fitness objective with timeline.
type Goal struct {
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date"`
}

// Injury records a user's past injury for training considerations.
type Injury struct {
	Description string `json:"description"`
	Year        int    `json:"year"`
}

// Age calculates the user's current age from their birthdate.
func (p *Profile) Age() int {
	now := time.Now()
	years := now.Year() - p.Birthdate.Year()
	if now.YearDay() < p.Birthdate.YearDay() {
		years--
	}
	return years
}

func (p *Profile) data() (profileData, error) {
	var data profileData
	err := json.Unmarshal(p.Data, &data)
	return data, err
}

// Goals extracts the user's fitness goals from profile data.
func (p *Profile) Goals() []Goal {
	data, err := p.data()
	if err != nil {
		return nil
	}
	return data.Goals
}

// Injuries extracts the user's injury history from profile data.
func (p *Profile) Injuries() []Injury {
	data, err := p.data()
	if err != nil {
		return nil
	}
	return data.Injuries
}

// Limitations extracts the user's training limitations from profile data.
func (p *Profile) Limitations() []string {
	data, err := p.data()
	if err != nil {
		return nil
	}
	return data.Limitations
}
