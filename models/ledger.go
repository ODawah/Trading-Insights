package models

import (
	"time"

	"gorm.io/datatypes"
)

// LedgerEntryType represents the kind of movement recorded in the ledger.
// Values are persisted as smallint in Postgres to match the requested layout.
type LedgerEntryType int16

const (
	LedgerEntryExchange   LedgerEntryType = 0
	LedgerEntryFee        LedgerEntryType = 1
	LedgerEntryAdjustment LedgerEntryType = 2
)

// UserLedgerEntry stores append-only signed movements for a user's trades.
// Table and indexes are named explicitly to mirror the requested DDL.
type UserLedgerEntry struct {
	ID         uint64          `gorm:"primaryKey;autoIncrement;type:bigserial;index:ul_user_time,priority:3,sort:desc;index:ul_user_currency_time,priority:4,sort:desc"`
	UserID     uint            `gorm:"not null;index:ul_user_time,priority:1;index:ul_user_currency_time,priority:1"`
	TradeID    string          `gorm:"type:uuid;not null;index:ul_trade"`
	Currency   string          `gorm:"type:text;not null;index:ul_user_currency_time,priority:2"`
	Amount     float64         `gorm:"type:numeric;not null"` // Signed: +inflow, -outflow
	ExecutedAt time.Time       `gorm:"type:timestamptz;not null;index:ul_user_time,priority:2,sort:desc;index:ul_user_currency_time,priority:3,sort:desc"`
	EntryType  LedgerEntryType `gorm:"type:smallint;not null"`
	Meta       datatypes.JSON  `gorm:"type:jsonb"` // Optional metadata: rate used, provider, notes, etc.
	CreatedAt  time.Time       `gorm:"autoCreateTime"`
}

func (UserLedgerEntry) TableName() string {
	return "user_ledger_entries"
}
