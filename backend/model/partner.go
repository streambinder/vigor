package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Partner struct {
	ID         uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	TrainingID uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_partner_training_user" json:"training_id"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_partner_training_user" json:"user_id"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	Training Training `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	User     User     `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
}
