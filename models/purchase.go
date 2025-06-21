package models

import (
	"gorm.io/gorm"
)

type Purchase struct {
	gorm.Model
	ID     uint    `gorm:"primaryKey"`
	Ticker string  `json:"ticker"`
	Amount float64 `json:"amount"` // Only in Dollars
	UserID int
	User   User
}
