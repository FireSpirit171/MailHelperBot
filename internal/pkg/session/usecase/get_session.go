package usecase

import "mail_helper_bot/internal/pkg/session/domain"

func (m *MemoryStorage) GetSession(chatID int64) (*domain.UserSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, exists := m.sessions[chatID]
	if !exists {
		return nil, nil
	}
	return session, nil
}
