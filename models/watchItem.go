package models

import "gorm.io/gorm"

type WatchItem struct {
	gorm.Model
	Ticker string `json:"ticker"`
	UserID uint   `json:"user_id"`
	User   User
}
