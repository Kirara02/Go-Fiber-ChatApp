package domain

import (
	"gorm.io/gorm"
)

// Session merepresentasikan state percakapan AI yang disimpan di database.
type Session struct {
	gorm.Model
	SessionID string `gorm:"uniqueIndex;not null"`
	Summary   string `gorm:"type:text"`
	UserID    uint   // Foreign key ke User, untuk melacak sesi milik siapa.
}