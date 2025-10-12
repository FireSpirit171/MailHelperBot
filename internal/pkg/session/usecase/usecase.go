package usecase

import (
	"mail_helper_bot/internal/pkg/session/domain"
	"sync"
	"time"
)

type MemoryStorage struct {
	sessions    map[int64]*domain.UserSession
	stateToChat map[string]int64
	stateExpiry map[string]time.Time
	mu          sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	storage := &MemoryStorage{
		sessions:    make(map[int64]*domain.UserSession),
		stateToChat: make(map[string]int64),
		stateExpiry: make(map[string]time.Time),
	}

	return storage
}
