package usecase

import "mail_helper_bot/internal/pkg/session/domain"

func (m *MemoryStorage) SaveSession(chatID int64, session *domain.UserSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[chatID] = session
	return nil
}
