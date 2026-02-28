package model

import (
	"time"

	"github.com/google/uuid"
)

type SharedLink struct {
	Token      string    `gorm:"primaryKey;type:varchar(22)" json:"token"`
	TrainingID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"training_id"`
	Training   Training  `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	CreatedAt  time.Time `json:"created_at"`
}
