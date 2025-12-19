package database

import (
	"fmt"

	"gorm.io/gorm"
)

// EnsureTimescale configures TimescaleDB features when the backing Postgres supports it.
// It is safe to call even when running on plain Postgres (it will return an error you can log and ignore).
func EnsureTimescale(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("nil db")
	}

	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS timescaledb;`).Error; err != nil {
		return fmt.Errorf("create timescaledb extension: %w", err)
	}

	// Ensure the primary key includes the time partitioning column to satisfy Timescale requirements.
	_ = db.Exec(`ALTER TABLE currencies DROP CONSTRAINT IF EXISTS currencies_pkey;`).Error
	_ = db.Exec(`ALTER TABLE currencies ADD PRIMARY KEY (id, fetched_time);`).Error

	// Convert the existing currencies time-series table into a hypertable.
	// Table name is "currencies" from models.Currency. Time column is "fetched_time".
	if err := db.Exec(`SELECT create_hypertable('currencies', 'fetched_time', if_not_exists => TRUE, migrate_data => TRUE);`).Error; err != nil {
		return fmt.Errorf("create hypertable currencies(fetched_time): %w", err)
	}

	// Helpful index for point-in-time lookups + history queries (if not already created by GORM tags).
	_ = db.Exec(`CREATE INDEX IF NOT EXISTS currencies_ticker_time ON currencies (ticker, fetched_time DESC);`).Error

	return nil
}
