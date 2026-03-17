package service

import (
	"testing"
	"time"
)

func TestParseTimezone(t *testing.T) {
	tests := []struct {
		name      string
		timezone  string
		wantErr   bool
		wantName  string
	}{
		{
			name:     "valid US timezone",
			timezone: "America/Los_Angeles",
			wantErr:  false,
			wantName: "America/Los_Angeles",
		},
		{
			name:     "valid EU timezone",
			timezone: "Europe/London",
			wantErr:  false,
			wantName: "Europe/London",
		},
		{
			name:     "valid UTC",
			timezone: "UTC",
			wantErr:  false,
			wantName: "UTC",
		},
		{
			name:     "empty string should error",
			timezone: "",
			wantErr:  true,
		},
		{
			name:     "invalid timezone should error",
			timezone: "Invalid/Timezone",
			wantErr:  true,
		},
		{
			name:     "garbage should error",
			timezone: "garbage",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := ParseTimezone(tt.timezone)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTimezone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && loc.String() != tt.wantName {
				t.Errorf("ParseTimezone() location = %v, want %v", loc.String(), tt.wantName)
			}
		})
	}
}

func TestTimezoneAttributionConsistency(t *testing.T) {
	// test that same UTC timestamp is attributed to correct calendar day across timezones
	utcTime := time.Date(2026, 3, 16, 2, 30, 0, 0, time.UTC) // 2:30 AM UTC on March 16

	tests := []struct {
		timezone string
		wantDate string
	}{
		{"UTC", "2026-03-16"},
		{"America/Los_Angeles", "2026-03-15"}, // PST is UTC-8, so 2:30 AM UTC = 6:30 PM March 15 PST
		{"America/New_York", "2026-03-15"},    // EST is UTC-5, so 2:30 AM UTC = 9:30 PM March 15 EST
		{"Europe/London", "2026-03-16"},       // GMT is UTC+0
		{"Asia/Tokyo", "2026-03-16"},          // JST is UTC+9, so 2:30 AM UTC = 11:30 AM March 16 JST
	}

	for _, tt := range tests {
		t.Run(tt.timezone, func(t *testing.T) {
			loc, err := time.LoadLocation(tt.timezone)
			if err != nil {
				t.Fatalf("failed to load timezone %s: %v", tt.timezone, err)
			}

			// convert UTC time to local timezone and format as date
			localDate := utcTime.In(loc).Format("2006-01-02")
			if localDate != tt.wantDate {
				t.Errorf("timezone %s: got date %s, want %s", tt.timezone, localDate, tt.wantDate)
			}
		})
	}
}

func TestExplicitUTCStorage(t *testing.T) {
	// verify that time.Now().UTC() returns a time in UTC location
	now := time.Now().UTC()

	if now.Location().String() != "UTC" {
		t.Errorf("time.Now().UTC() location = %v, want UTC", now.Location().String())
	}

	// verify offset is 0
	_, offset := now.Zone()
	if offset != 0 {
		t.Errorf("time.Now().UTC() offset = %d seconds, want 0", offset)
	}
}
