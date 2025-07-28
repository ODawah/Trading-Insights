package models

import (
	"gorm.io/gorm"
	"time"
)

type Snapshot struct {
	Base      string             `json:"base"` // e.g., "GBP"
	Result    map[string]float64 `json:"result"`
	Timestamp time.Time          `json:"timestamp"` // e.g., "2023-10-01T12:00:00Z"
}

type Currency struct {
	gorm.Model
	Ticker      string    `json:"ticker" gorm:"PrimaryKey"`
	Base        string    `json:"base"`
	Rate        float64   `json:"rate"`
	FetchedTime time.Time `json:"timestamp"`
}
