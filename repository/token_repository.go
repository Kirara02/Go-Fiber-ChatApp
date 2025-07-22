package repository

import (
	"main/domain"
	"gorm.io/gorm"
)

type TokenRepository interface {
	CreateInvalidatedToken(token *domain.InvalidatedToken) error
	IsTokenInvalidated(tokenString string) (bool, error)
}

type tokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) CreateInvalidatedToken(token *domain.InvalidatedToken) error {
	return r.db.Create(token).Error
}

func (r *tokenRepository) IsTokenInvalidated(tokenString string) (bool, error) {
	var token domain.InvalidatedToken
	err := r.db.Where("token = ?", tokenString).First(&token).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}