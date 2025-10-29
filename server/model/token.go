package model

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken stores JWT refresh tokens for session management.
type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Token     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Revoked   bool      `gorm:"default:false"`
	CreatedAt time.Time
}

func (RefreshToken) TableName() string {
	return "tokens"
}
