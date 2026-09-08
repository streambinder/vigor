package model

import (
	"time"

	"github.com/google/uuid"
)

// Proficiency represents a user's demonstrated level for a muscle group.
// Multiple records per user+muscle are allowed for historical tracking.
type Proficiency struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index:idx_proficiency_user" json:"user_id"`
	TrainingID uuid.UUID `gorm:"type:uuid;not null;index:idx_proficiency_training" json:"training_id"`
	Muscle     string    `gorm:"type:varchar(64);not null" json:"muscle"`
	Value      float64   `gorm:"not null" json:"value"`
	CreatedAt  time.Time `json:"created_at"`

	User     User     `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	Training Training `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
}
