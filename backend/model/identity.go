package model

import (
	"time"

	"github.com/google/uuid"
)

// Identity represents an authentication method for a user.
// A user can have multiple identities (local password, Google OAuth, Apple OAuth, etc.)
type Identity struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Provider       string    `gorm:"not null;index" json:"provider"` // local, google, apple
	ProviderUserID string    `gorm:"index" json:"provider_user_id"`  // Unique ID from OAuth provider (null for local)
	PasswordHash   string    `gorm:"" json:"-"`                      // Password hash for local auth (null for OAuth)

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:UserID;references:ID" json:"-"`
}
