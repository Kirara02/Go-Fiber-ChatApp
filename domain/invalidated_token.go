package domain

import (
	"time"

	"gorm.io/gorm"
)

type InvalidatedToken struct {
	gorm.Model
	Token     string    `gorm:"unique;not null;index"`
	ExpiresAt time.Time `gorm:"not null"`
}