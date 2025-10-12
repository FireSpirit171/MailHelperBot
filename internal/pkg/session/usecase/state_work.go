package usecase

import (
	"fmt"
	"time"
)

func (m *MemoryStorage) SaveState(state string, chatID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stateToChat[state] = chatID
	m.stateExpiry[state] = time.Now().Add(10 * time.Minute)
	return nil
}

func (m *MemoryStorage) GetChatIDByState(state string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	chatID, exists := m.stateToChat[state]
	if !exists {
		return 0, fmt.Errorf("state not found")
	}

	if expiry, exists := m.stateExpiry[state]; exists && time.Now().After(expiry) {
		return 0, fmt.Errorf("state expired")
	}

	return chatID, nil
}

func (m *MemoryStorage) CleanupExpiredStates() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for state, expiry := range m.stateExpiry {
		if now.After(expiry) {
			delete(m.stateToChat, state)
			delete(m.stateExpiry, state)
		}
	}
	return nil
}

func (m *MemoryStorage) DeleteState(state string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.stateToChat, state)
	delete(m.stateExpiry, state)
	return nil
}
