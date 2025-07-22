package models

import (
	"gorm.io/gorm"
	"time"
)

type CurrencyHistory struct {
	gorm.Model
	ID        uint           `gorm:"primaryKey"`
	Timestamp time.Time      `json:"timestamp"`
	Source    string         `json:"source"`
	Quotes    []QuoteHistory `gorm:"foreignKey:CurrencyHistoryID" json:"quotes"`
}

type QuoteHistory struct {
	gorm.Model
	CurrencyHistoryID uint    // Foreign key
	Currency          string  `json:"currency"`
	Rate              float64 `json:"rate"`
}
