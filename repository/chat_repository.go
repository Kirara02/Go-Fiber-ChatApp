package repository

import (
	"main/domain"

	"gorm.io/gorm"
)

type ChatRepository interface {
	CreateMessage(message *domain.ChatMessage) error
	GetMessagesByRoomID(roomID uint, limit int, offset int) ([]domain.ChatMessage, error)
}

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) CreateMessage(message *domain.ChatMessage) error {
	return r.db.Create(message).Error
}

func (r *chatRepository) GetMessagesByRoomID(roomID uint, limit int, offset int) ([]domain.ChatMessage, error) {
	var messages []domain.ChatMessage

	err := r.db.
		Preload("User").
		Where("room_id = ?", roomID).
		Order("created_at asc").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	return messages, nil
}
