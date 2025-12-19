package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name         string        `json:"name"`
	Password     string        `gorm:"not null" json:"-"`
	Email        string        `gorm:"uniqueIndex;not null" json:"email"`
	Transactions []Transaction `gorm:"foreignKey:UserID"`
	WatchItems   []WatchItem   `gorm:"foreignKey:UserID"`
}

func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
