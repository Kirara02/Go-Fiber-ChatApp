package domain

import "gorm.io/gorm"

type ChatMessage struct {
	gorm.Model

	Type       string `json:"type" gorm:"-"`
	SenderID   uint   `json:"senderId,omitempty" gorm:"-"`
	SenderName string `json:"senderName,omitempty" gorm:"-"`
	Content    string `json:"content,omitempty" gorm:"not null"`
	RoomID     uint   `json:"roomId,omitempty" gorm:"not null;index"`
	UserID     uint   `json:"-" gorm:"not null"`
	User       User   `json:"-" gorm:"foreignKey:UserID"`
}
