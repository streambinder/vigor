package database

import (
	"os"

	"gorm.io/driver/postgres"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	Metrics   *gorm.DB
	DB        *gorm.DB
	Knowledge *gorm.DB
)

func Init() error {
	cfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}

	// metrics db (sqlite, optional)
	if path := os.Getenv("METRICS_URL"); path != "" {
		db, err := gorm.Open(sqlite.Open(path+"?mode=ro"), cfg)
		if err != nil {
			return err
		}
		Metrics = db
	}

	// main db (postgres)
	if url := os.Getenv("DATABASE_URL"); url != "" {
		db, err := gorm.Open(postgres.Open(url), cfg)
		if err != nil {
			return err
		}
		DB = db
	}

	// knowledge db (postgres)
	if url := os.Getenv("KNOWLEDGE_URL"); url != "" {
		db, err := gorm.Open(postgres.Open(url), cfg)
		if err != nil {
			return err
		}
		Knowledge = db
	}

	return nil
}
