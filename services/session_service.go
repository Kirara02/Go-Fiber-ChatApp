package services

import (
	"main/domain"
	"main/repository"
)

// SessionService mendefinisikan interface untuk logika bisnis sesi.
// Untuk saat ini, method-nya akan mencerminkan repository,
// namun di sini Anda bisa menambahkan logika bisnis di masa depan.
type SessionService interface {
	GetOrCreateSession(sessionID string, userID uint) (*domain.Session, error)
	UpdateSummary(session *domain.Session, newSummary string) error
}

type sessionService struct {
	repo repository.SessionRepository
}

// NewSessionService membuat instance baru dari sessionService.
// Ini akan di-inject oleh Wire.
func NewSessionService(repo repository.SessionRepository) SessionService {
	return &sessionService{repo: repo}
}

// GetOrCreateSession memanggil repository untuk mendapatkan atau membuat sesi.
func (s *sessionService) GetOrCreateSession(sessionID string, userID uint) (*domain.Session, error) {
	return s.repo.GetOrCreateSession(sessionID, userID)
}

// UpdateSummary memanggil repository untuk memperbarui ringkasan.
func (s *sessionService) UpdateSummary(session *domain.Session, newSummary string) error {
	return s.repo.UpdateSummary(session, newSummary)
}
