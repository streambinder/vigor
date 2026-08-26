// Package daily provides day-keyed volatile snapshots on the main database.
//
// Homepage data that is expensive to build (LLM probes, heavy aggregations)
// lives here: one row per user and calendar day, durable across backend
// restarts and discarded once it ages out. On Postgres every snapshot type
// is a table inside the daily schema, RANGE-partitioned by day so that
// retention is a metadata-only DROP PARTITION instead of a DELETE scan;
// other dialects (sqlite unit tests) fall back to a plain daily_<name> table.
package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// retentionDays bounds how far back daily snapshots are kept before the
// day-rollover maintenance drops their partition.
const retentionDays = 14

// Table allowlist: identifiers are interpolated into DDL, so they must be
// compile-time constants, never user input.
const (
	TableReadiness = "readiness"
)

var tables = map[string]bool{
	TableReadiness: true,
}

type row struct {
	Payload string `gorm:"column:payload"`
}

func tableName(db *gorm.DB, table string) (string, error) {
	if !tables[table] {
		return "", fmt.Errorf("unknown daily table %q", table)
	}
	if db.Dialector.Name() == "postgres" {
		return "daily." + table, nil
	}
	return "daily_" + table, nil
}

// dayDate normalizes an instant to the calendar day seen from loc.
func dayDate(at time.Time, loc *time.Location) time.Time {
	day, err := time.ParseInLocation("2006-01-02", at.In(loc).Format("2006-01-02"), time.UTC)
	if err != nil {
		return time.Now().UTC().Truncate(24 * time.Hour)
	}
	return day
}

func partitionName(table string, day time.Time) string {
	return fmt.Sprintf("%s_%s", table, day.Format("20060102"))
}

// utcDay normalizes an instant to the UTC calendar day; partition boundaries
// and the maintenance stamp share it so every table touch is idempotent.
func utcDay(at time.Time) time.Time {
	return at.UTC().Truncate(24 * time.Hour)
}

// Boot prepares the daily schema on startup: namespaced partitioned tables,
// the day partitions around today and the retention sweep.
func bootDaily(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" {
		return nil
	}
	if err := db.Exec("CREATE SCHEMA IF NOT EXISTS daily").Error; err != nil {
		return fmt.Errorf("daily schema: %w", err)
	}
	for table := range tables {
		if err := db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS daily.%s (
				user_id uuid NOT NULL,
				day date NOT NULL,
				payload jsonb NOT NULL,
				updated_at timestamptz NOT NULL DEFAULT now(),
				PRIMARY KEY (user_id, day)
			) PARTITION BY RANGE (day)`, table)).Error; err != nil {
			return fmt.Errorf("daily %s table: %w", table, err)
		}
		if err := db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS daily.%s
			PARTITION OF daily.%s DEFAULT`, table+"_default", table)).Error; err != nil {
			return fmt.Errorf("daily %s default partition: %w", table, err)
		}
	}
	for table := range tables {
		maintain(DB, table)
	}
	return nil
}

// maintain runs once per UTC day per table: it creates the partitions
// spanning yesterday through tomorrow (user timezones sit on both sides of
// UTC) and drops partitions older than retention. Rows that landed in the
// default partition because theirs was missing are swept with a tiny DELETE.
func maintain(db *gorm.DB, table string) {
	if db.Dialector.Name() != "postgres" {
		return
	}
	for i := -1; i <= 1; i++ {
		day := utcDay(time.Now()).AddDate(0, 0, i)
		name := partitionName(table, day)
		stmt := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS daily.%s
			PARTITION OF daily.%s FOR VALUES FROM (?) TO (?)`,
			name, table)
		if err := db.Exec(stmt, day.Format("2006-01-02"), day.AddDate(0, 0, 1).Format("2006-01-02")).Error; err != nil {
			log.Warn().Err(err).Str("partition", name).Msg("daily partition creation failed")
		}
	}

	cutoff := utcDay(time.Now().AddDate(0, 0, -retentionDays))
	var partitions []string
	if err := db.Raw(`SELECT inhrelid::regclass::text
			FROM pg_inherits
			WHERE inhparent = ?`, "daily."+table).
		Scan(&partitions).Error; err != nil {
		log.Warn().Err(err).Str("table", table).Msg("daily partition listing failed")
		return
	}
	for _, part := range partitions {
		name := part
		if idx := len("daily."); len(name) > idx && name[:idx] == "daily." {
			name = name[idx:]
		}
		if len(name) < 8 {
			continue
		}
		day, err := time.Parse("20060102", name[len(name)-8:])
		if err != nil {
			// default partition or foreign naming: leave alone
			continue
		}
		if !day.Before(cutoff) {
			continue
		}
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS daily.%s", name)).Error; err != nil {
			log.Warn().Err(err).Str("partition", name).Msg("daily partition drop failed")
		}
	}
	if err := db.Exec(fmt.Sprintf("DELETE FROM daily.%s WHERE day < ?", table+"_default"),
		cutoff.Format("2006-01-02")).Error; err != nil {
		log.Warn().Err(err).Str("table", table).Msg("daily default partition sweep failed")
	}
}

// Load returns the snapshot payload for (table, user, day) as raw JSON, or
// nil when no snapshot exists.
func DailyLoad(table string, userID uuid.UUID, at time.Time, loc *time.Location) (json.RawMessage, error) {
	name, err := tableName(DB, table)
	if err != nil {
		return nil, err
	}
	maintain(DB, table)
	var r row
	err = DB.Table(name).
		Select("payload").
		Where("user_id = ? AND day = ?", userID, dayDate(at, loc).Format("2006-01-02")).
		Take(&r).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("daily %s load: %w", table, err)
	}
	return json.RawMessage(r.Payload), nil
}

// Save upserts the snapshot for (table, user, day), overwriting any previous
// payload stored for the same day.
func DailySave(table string, userID uuid.UUID, at time.Time, loc *time.Location, payload any) error {
	name, err := tableName(DB, table)
	if err != nil {
		return err
	}
	maintain(DB, table)
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("daily %s marshal: %w", table, err)
	}
	stmt := fmt.Sprintf(`INSERT INTO %s (user_id, day, payload, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (user_id, day) DO UPDATE SET payload = EXCLUDED.payload, updated_at = EXCLUDED.updated_at`,
		name)
	if err := DB.Exec(stmt, userID, dayDate(at, loc).Format("2006-01-02"), string(raw), time.Now().UTC()).Error; err != nil {
		return fmt.Errorf("daily %s save: %w", table, err)
	}
	return nil
}
