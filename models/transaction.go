package models

import (
	"gorm.io/gorm"
)

type Transaction struct {
	gorm.Model
	UserID   int     `json:"user_id"`
	User     User    `gorm:"foreignKey:UserID"`
	FromTick string  `json:"from_tick"`
	ToTick   string  `json:"to_tick"`
	FromAmt  float64 `json:"from_amt"`
	ToAmt    float64 `json:"to_amt"`
	FxRate   float64 // Formula : (ToAmt / FromAmt)
}
