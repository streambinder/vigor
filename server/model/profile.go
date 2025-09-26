package model

import (
	"encoding/json"
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

type profileData struct {
	Goals       []string `json:"goals"`
	Injuries    []Injury `json:"injuries"`
	Limitations []string `json:"limitations"`
}

type Injury struct {
	Description string `json:"description"`
	Year        int    `json:"year"`
}

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

func (p *Profile) Goals() []string {
	data, err := p.data()
	if err != nil {
		return nil
	}
	return data.Goals
}

func (p *Profile) Injuries() []Injury {
	data, err := p.data()
	if err != nil {
		return nil
	}
	return data.Injuries
}

func (p *Profile) Limitations() []string {
	data, err := p.data()
	if err != nil {
		return nil
	}
	return data.Limitations
}
