package service

import (
	"bytes"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm/clause"
)

const (
	maxAvatarSize      = 256 * 1024
	maxAvatarDimension = 512
)

var (
	ErrAvatarTooLarge    = errors.New("avatar exceeds maximum size of 256 KB")
	ErrAvatarInvalidType = errors.New("only PNG and JPEG images are allowed")
	ErrAvatarInvalidData = errors.New("invalid image data")
	ErrAvatarNotSquare   = errors.New("avatar must be square")
	ErrAvatarTooLargeDim = errors.New("avatar dimensions exceed 512x512")
)

var allowedAvatarTypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
}

func GetAvatar(userID uuid.UUID) (*model.Avatar, error) {
	var avatar model.Avatar
	if err := database.DB.First(&avatar, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &avatar, nil
}

func SetAvatar(userID uuid.UUID, data []byte) error {
	if len(data) > maxAvatarSize {
		return ErrAvatarTooLarge
	}

	// detect actual content type from bytes, ignore client-provided header
	contentType := http.DetectContentType(data)
	if !allowedAvatarTypes[contentType] {
		return ErrAvatarInvalidType
	}

	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return ErrAvatarInvalidData
	}
	if cfg.Width != cfg.Height {
		return ErrAvatarNotSquare
	}
	if cfg.Width > maxAvatarDimension {
		return ErrAvatarTooLargeDim
	}

	return database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"data", "content_type", "updated_at"}),
	}).Create(&model.Avatar{
		UserID:      userID,
		Data:        data,
		ContentType: contentType,
	}).Error
}
