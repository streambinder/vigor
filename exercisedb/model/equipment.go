package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Equipment represents exercise equipment.
type Equipment struct {
	ID   uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name string    `gorm:"not null;uniqueIndex:idx_equipment_name" json:"name"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
