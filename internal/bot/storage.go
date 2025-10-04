package bot

import (
	"fmt"
	"sync"
	"time"
)

type Storage interface {
	SaveSession(chatID int64, session *UserSession) error
	GetSession(chatID int64) (*UserSession, error)
	DeleteSession(chatID int64) error
	SaveState(state string, chatID int64) error
	GetChatIDByState(state string) (int64, error)
	CleanupExpiredStates() error
	DeleteState(state string) error
}

type MemoryStorage struct {
	sessions    map[int64]*UserSession
	stateToChat map[string]int64
	stateExpiry map[string]time.Time
	mu          sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	storage := &MemoryStorage{
		sessions:    make(map[int64]*UserSession),
		stateToChat: make(map[string]int64),
		stateExpiry: make(map[string]time.Time),
	}

	go storage.cleanupWorker()

	return storage
}

func (m *MemoryStorage) cleanupWorker() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		m.CleanupExpiredStates()
	}
}

func (m *MemoryStorage) SaveSession(chatID int64, session *UserSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[chatID] = session
	return nil
}

func (m *MemoryStorage) GetSession(chatID int64) (*UserSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, exists := m.sessions[chatID]
	if !exists {
		return nil, nil
	}
	return session, nil
}

func (m *MemoryStorage) DeleteSession(chatID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, chatID)
	return nil
}

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
