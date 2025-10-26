package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Gym struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name      string         `gorm:"not null;uniqueIndex:idx_user_gym_name" json:"name"`
	Equipment pq.StringArray `gorm:"type:text[]" json:"equipment"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_gym_name" json:"-"`
	User   User      `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
}
