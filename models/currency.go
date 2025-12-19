package models

import "time"

type Snapshot struct {
	Base      string             `json:"base"` // e.g., "GBP"
	Result    map[string]float64 `json:"result"`
	Timestamp time.Time          `json:"timestamp"` // e.g., "2023-10-01T12:00:00Z"
}

type Currency struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement;index:currency_pk,priority:1"`
	Ticker      string    `json:"ticker" gorm:"not null;index:idx_currency_time,priority:1"`
	Base        string    `json:"base"`
	Rate        float64   `json:"rate"`
	FetchedTime time.Time `json:"timestamp" gorm:"not null;index:idx_currency_time,priority:2,sort:desc;primaryKey;index:currency_pk,priority:2"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
