package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID         uint        `gorm:"primaryKey"`
	Name       string      `json:"name"`
	Email      string      `json:"email" gorm:"unique"`
	Purchases  []Purchase  `gorm:"foreignKey:UserID"`
	WatchItems []WatchItem `gorm:"foreignKey:UserID"`
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
