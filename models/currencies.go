package models

import "gorm.io/gorm"

type Currencies struct {
	Base   string             `json:"base"`
	Result map[string]float64 `json:"result"`
}

type Currency struct {
	gorm.Model
	Ticker string  `json:"ticker" gorm:"PrimaryKey"`
	Base   string  `json:"base"`
	Rate   float64 `json:"rate"`
}
