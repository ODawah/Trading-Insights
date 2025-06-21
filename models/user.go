package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID        uint     `gorm:"primaryKey"`
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	WatchList []string `json:"watch_list"`
}

func (u *User) CreateUser(db *gorm.DB) error {
	err := db.Create(&u).Error
	if err != nil {
		return err
	}
	return err
}

func FindUserByID(id int, db *gorm.DB) (User, error) {
	var user User
	if err := db.First(&user, id).Error; err != nil {
		return user, err
	}
	return user, nil
}

//func FindUserWatchList(id int, db *gorm.DB)
