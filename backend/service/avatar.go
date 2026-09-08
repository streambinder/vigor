package service

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash/fnv"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/database"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	maxAvatarSize      = 256 * 1024
	maxAvatarDimension = 512

	defaultAvatarDimension = 128
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
	// Find reports a missing row via RowsAffected instead of an error, so the
	// caller keeps explicit control over the not-found case.
	res := database.DB.Where("user_id = ?", userID).Find(&avatar)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &avatar, nil
}

// defaultAvatarCache memoizes generated placeholders, keyed by user ID.
var defaultAvatarCache sync.Map // uuid.UUID -> cachedDefaultAvatar

type cachedDefaultAvatar struct {
	data []byte
	etag string
}

// DefaultAvatar returns a deterministic placeholder PNG for users without a
// custom avatar, derived from the user ID so it is stable without stored
// state. The returned etag is stable for the same user ID.
func DefaultAvatar(userID uuid.UUID) ([]byte, string) {
	if cached, ok := defaultAvatarCache.Load(userID); ok {
		entry := cached.(cachedDefaultAvatar)
		return entry.data, entry.etag
	}

	img := image.NewRGBA(image.Rect(0, 0, defaultAvatarDimension, defaultAvatarDimension))
	h := fnv.New64a()
	_, _ = h.Write(userID[:])
	top, bottom := defaultAvatarColors(h.Sum64())

	for y := range defaultAvatarDimension {
		t := float64(y) / float64(defaultAvatarDimension-1)
		c := lerpColor(top, bottom, t)
		for x := range defaultAvatarDimension {
			img.Set(x, y, c)
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, ""
	}
	data := buf.Bytes()
	sum := sha256.Sum256(data)
	entry := cachedDefaultAvatar{data: data, etag: fmt.Sprintf(`"default-%x"`, sum[:8])}
	defaultAvatarCache.Store(userID, entry)
	return entry.data, entry.etag
}

// defaultAvatarColors picks two hues from a stable palette slot so each user
// gets a recognizable, pleasant gradient.
func defaultAvatarColors(seed uint64) (color.RGBA, color.RGBA) {
	hue := float64(seed % 360)
	return hsvToRGB(hue, 0.55, 0.75), hsvToRGB(hue+40, 0.65, 0.45)
}

func hsvToRGB(h, s, v float64) color.RGBA {
	c := v * s
	x := c * (1 - absFloat(modFloat(h/60, 2)-1))
	m := v - c

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return color.RGBA{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
		A: 255,
	}
}

func lerpColor(a, b color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(a.R)*(1-t) + float64(b.R)*t),
		G: uint8(float64(a.G)*(1-t) + float64(b.G)*t),
		B: uint8(float64(a.B)*(1-t) + float64(b.B)*t),
		A: 255,
	}
}

func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func modFloat(x, y float64) float64 {
	for x >= y {
		x -= y
	}
	return x
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
