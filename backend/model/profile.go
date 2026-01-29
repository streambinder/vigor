package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Profile stores user's physical attributes and fitness-related metadata.
type Profile struct {
	FirstName string         `json:"first_name" flutter:"required"`
	LastName  string         `json:"last_name" flutter:"required"`
	Birthdate time.Time      `json:"birthdate" flutter:"required"`
	Gender    string         `json:"gender" flutter:"required"`
	Language  string         `gorm:"default:'english'" json:"language" flutter:"required"`
	Height    float64        `json:"height" flutter:"required"`
	Weight    float64        `json:"weight" flutter:"required"`
	Data      datatypes.JSON `gorm:"type:jsonb" json:"data"`

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	UserID uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	User   *User     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

type profileData struct {
	Goals       []string     `json:"goals" flutter:"required"`
	Injuries    []Injury     `json:"injuries"`
	Limitations []string     `json:"limitations"`
	Conditions  []string     `json:"conditions"`
	Preferences *Preferences `json:"preferences,omitempty"`
}

// Preferences stores user's favorite exercises and equipment.
type Preferences struct {
	Exercises []string `json:"exercises,omitempty"`
	Equipment []string `json:"equipment,omitempty"`
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
func (p *Profile) Goals() []string {
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

// Conditions extracts the user's body conditions from profile data.
func (p *Profile) Conditions() []string {
	data, err := p.data()
	if err != nil {
		return nil
	}
	return data.Conditions
}

// FavoriteExercises extracts the user's favorite exercises from profile data.
func (p *Profile) FavoriteExercises() []string {
	data, err := p.data()
	if err != nil || data.Preferences == nil {
		return nil
	}
	return data.Preferences.Exercises
}

// FavoriteEquipment extracts the user's favorite equipment from profile data.
func (p *Profile) FavoriteEquipment() []string {
	data, err := p.data()
	if err != nil || data.Preferences == nil {
		return nil
	}
	return data.Preferences.Equipment
}
