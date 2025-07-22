package domain

import "gorm.io/gorm"

type User struct {
	gorm.Model

	Name     string `gorm:"not null"`
	Email    string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
	Rooms    []*Room `gorm:"many2many:user_rooms;"`
}
