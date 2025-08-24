package model

import "time"

type User struct {
    ID        uint      `gorm:"primaryKey"`
    Email     string    `gorm:"uniqueIndex;not null"`
    Password  string    `gorm:"not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

type RefreshToken struct {
    ID        uint      `gorm:"primaryKey"`
    UserID    uint      `gorm:"not null;index"`
    Token     string    `gorm:"uniqueIndex;not null"`
    ExpiresAt time.Time `gorm:"not null"`
    Revoked   bool      `gorm:"default:false"`
    CreatedAt time.Time
}
