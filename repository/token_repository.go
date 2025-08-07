package repository

import (
	"main/domain"
	"time"

	"gorm.io/gorm"
)

type TokenRepository interface {
	CreateInvalidatedToken(token *domain.InvalidatedToken) error
	IsTokenInvalidated(tokenString string) (bool, error)
	CleanExpiredTokens() (int64, error)
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

func (r *tokenRepository) CleanExpiredTokens() (int64, error) {
	result := r.db.Where("expires_at < ?", time.Now()).Delete(&domain.InvalidatedToken{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}