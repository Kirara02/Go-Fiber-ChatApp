package domain

import "gorm.io/gorm"

type RoomType string

const (
	RoomTypeDM    RoomType = "dm"
	RoomTypeGroup RoomType = "group"
)

type Room struct {
	gorm.Model
	Name string `json:"name"`
	// HAPUS IsPrivate dan ganti dengan Type
	// IsPrivate   bool          `json:"isPrivate" gorm:"default:false"`
	Type        RoomType      `json:"type" gorm:"type:varchar(10);not null;index"`
	OwnerID     *uint         `json:"ownerId,omitempty" gorm:"index"`
	RoomImage   *string       `json:"room_image"`
	Users       []*User       `json:"users,omitempty" gorm:"many2many:user_rooms;"`
	Messages    []ChatMessage `json:"messages,omitempty" gorm:"foreignKey:RoomID"`
	LastMessage ChatMessage   `json:"-" gorm:"-"`
}
