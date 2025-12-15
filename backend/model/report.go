package model

import (
	"time"

	"github.com/google/uuid"
)

// Report stores user-submitted exercise/training issues for manual dev review.
// Unlike feedback, reports are not fed to the LLM.
type Report struct {
	ID      uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Content string    `gorm:"not null" json:"content"`

	TrainingID *uuid.UUID `gorm:"type:uuid" json:"training_id"`
	Training   *Training  `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	ActivityID *string    `gorm:"type:uuid" json:"activity_id"`
	Activity   *Activity  `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	User       User       `gorm:"constraint:OnDelete:CASCADE;" json:"-"`

	CreatedAt time.Time `json:"created_at"`
}
