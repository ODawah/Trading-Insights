package models

import "gorm.io/gorm"

type WatchItem struct {
	gorm.Model
	Ticker string `json:"symbol" gorm:"unique"`
	UserID uint
}
