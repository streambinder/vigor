package database

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/streambinder/vigor/model"
	"gorm.io/gorm"
)

type (
	modelHealthSleepDaily    = model.HealthSleepDaily
	modelHealthRecoveryDaily = model.HealthRecoveryDaily
)

// Health daily tables are RANGE-partitioned by date on postgres so each
// metric group ages out through a metadata-only partition drop, with no
// external purge job. Retention is per metric group: sleep is only read
// over a 30-day baseline window, recovery (HRV, resting HR) needs a
// 60-day reference window. On other dialects (sqlite tests) the models
// migrate as plain tables and retention is a no-op.
type healthTable struct {
	name          string
	retentionDays int
	columns       string
}

var healthTables = []healthTable{
	{
		name:          "health_sleep_daily",
		retentionDays: 30,
		columns: `user_id uuid NOT NULL,
			date date NOT NULL,
			sleep_hours double precision,
			synced_at timestamptz NOT NULL DEFAULT now(),
			PRIMARY KEY (user_id, date)`,
	},
	{
		name:          "health_recovery_daily",
		retentionDays: 60,
		columns: `user_id uuid NOT NULL,
			date date NOT NULL,
			resting_hr integer,
			hrv_rmssd double precision,
			synced_at timestamptz NOT NULL DEFAULT now(),
			PRIMARY KEY (user_id, date)`,
	},
}

func healthPartitionName(table string, day time.Time) string {
	return fmt.Sprintf("%s_%s", table, day.Format("20060102"))
}

// bootHealthDaily creates the partitioned health tables and runs the
// first retention sweep.
func bootHealthDaily(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" {
		return db.AutoMigrate(&modelHealthSleepDaily{}, &modelHealthRecoveryDaily{})
	}
	for _, table := range healthTables {
		if err := db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (%s) PARTITION BY RANGE (date)`, table.name, table.columns)).Error; err != nil {
			return fmt.Errorf("health %s table: %w", table.name, err)
		}
		if err := db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s PARTITION OF %s DEFAULT`, table.name+"_default", table.name)).Error; err != nil {
			return fmt.Errorf("health %s default partition: %w", table.name, err)
		}
	}
	for _, table := range healthTables {
		MaintainHealthDaily(db, table.name)
	}
	return nil
}

// MaintainHealthDaily creates the partitions spanning the table's whole
// retention window plus tomorrow, and drops partitions older than the
// retention cutoff. It runs at boot and before each health sync, so
// retention needs no external job.
func MaintainHealthDaily(db *gorm.DB, tableName string) {
	if db.Dialector.Name() != "postgres" {
		return
	}
	var table *healthTable
	for i := range healthTables {
		if healthTables[i].name == tableName {
			table = &healthTables[i]
			break
		}
	}
	if table == nil {
		return
	}

	today := utcDay(time.Now())
	for i := -table.retentionDays; i <= 1; i++ {
		day := today.AddDate(0, 0, i)
		name := healthPartitionName(table.name, day)
		stmt := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s PARTITION OF %s FOR VALUES FROM ('%s') TO ('%s')`,
			name, table.name, day.Format("2006-01-02"), day.AddDate(0, 0, 1).Format("2006-01-02"))
		if err := db.Exec(stmt).Error; err != nil {
			log.Warn().Err(err).Str("partition", name).Msg("health partition creation failed")
		}
	}

	cutoff := today.AddDate(0, 0, -table.retentionDays)
	var partitions []string
	if err := db.Raw(`SELECT inhrelid::regclass::text
			FROM pg_inherits
			WHERE inhparent = ?::regclass`, table.name).
		Scan(&partitions).Error; err != nil {
		log.Warn().Err(err).Str("table", table.name).Msg("health partition listing failed")
		return
	}
	for _, part := range partitions {
		name := part
		if len(name) < 8 {
			continue
		}
		day, err := time.Parse("20060102", name[len(name)-8:])
		if err != nil {
			continue
		}
		if !day.Before(cutoff) {
			continue
		}
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", name)).Error; err != nil {
			log.Warn().Err(err).Str("partition", name).Msg("health partition drop failed")
		}
	}
	if err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE date < ?", table.name+"_default"),
		cutoff.Format("2006-01-02")).Error; err != nil {
		log.Warn().Err(err).Str("table", table.name).Msg("health default partition sweep failed")
	}
}

// MaintainHealth runs the retention sweep for every health table.
func MaintainHealth(db *gorm.DB) {
	for _, table := range healthTables {
		MaintainHealthDaily(db, table.name)
	}
}
