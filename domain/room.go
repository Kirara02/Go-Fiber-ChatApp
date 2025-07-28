package domain

import "gorm.io/gorm"

type Room struct {
	gorm.Model
	Name      string        `json:"name"`
	IsPrivate bool          `json:"isPrivate" gorm:"default:false"`
	OwnerID   *uint         `json:"ownerId,omitempty" gorm:"index"`
	Users     []*User       `json:"users,omitempty" gorm:"many2many:user_rooms;"`
	Messages  []ChatMessage `json:"messages,omitempty" gorm:"foreignKey:RoomID"`
	LastMessage ChatMessage   `json:"-" gorm:"-"`
}