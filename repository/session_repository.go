package repository

import (
	"main/domain"

	"gorm.io/gorm"
)

// SessionRepository mendefinisikan interface untuk operasi database sesi.
type SessionRepository interface {
	GetOrCreateSession(sessionID string, userID uint) (*domain.Session, error)
	UpdateSummary(session *domain.Session, newSummary string) error
}

type sessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository membuat instance baru dari sessionRepository.
func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{db: db}
}

// GetOrCreateSession mencari sesi berdasarkan sessionID. Jika tidak ditemukan, sesi baru akan dibuat.
func (r *sessionRepository) GetOrCreateSession(sessionID string, userID uint) (*domain.Session, error) {
	var session domain.Session
	// Coba cari sesi yang sudah ada terlebih dahulu.
	err := r.db.Where(domain.Session{SessionID: sessionID}).First(&session).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Jika tidak ditemukan, buat sesi baru.
			newSession := domain.Session{
				SessionID: sessionID,
				Summary:   "Percakapan baru dimulai.",
				UserID:    userID,
			}
			if err := r.db.Create(&newSession).Error; err != nil {
				return nil, err
			}
			return &newSession, nil
		}
		// Kembalikan error lain jika terjadi.
		return nil, err
	}
	return &session, nil
}

// UpdateSummary memperbarui field ringkasan dari sebuah sesi.
func (r *sessionRepository) UpdateSummary(session *domain.Session, newSummary string) error {
	return r.db.Model(session).Update("summary", newSummary).Error
}
