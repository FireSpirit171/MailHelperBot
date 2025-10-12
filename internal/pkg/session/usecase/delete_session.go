package usecase

func (m *MemoryStorage) DeleteSession(chatID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, chatID)
	return nil
}
