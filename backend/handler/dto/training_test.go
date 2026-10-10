package dto

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/streambinder/vigor/model"
)

func TestNewPartnerInfo(t *testing.T) {
	partner := model.Partner{
		ID:         uuid.New(),
		TrainingID: uuid.New(),
		UserID:     uuid.New(),
		CreatedAt:  time.Date(2026, 1, 2, 3, 4, 5, 0, time.FixedZone("X", 3600)),
	}
	partner.User.Profile.FirstName = "Ada"
	partner.User.Profile.LastName = "Lovelace"

	info := NewPartnerInfo(partner)
	if info.ID != partner.ID.String() || info.TrainingID != partner.TrainingID.String() || info.UserID != partner.UserID.String() {
		t.Errorf("ids = %+v", info)
	}
	if info.FirstName != "Ada" || info.LastName != "Lovelace" {
		t.Errorf("names = %+v", info)
	}
	if info.CreatedAt != "2026-01-02T02:04:05Z" {
		t.Errorf("created = %q", info.CreatedAt)
	}
}
