package oauth_service

import (
	"mail_helper_bot/internal/pkg/session/domain"
	"time"
)

type Storage interface {
	// Методы для работы с сессиями
	SaveSession(chatID int64, session *domain.UserSession) error
	GetSession(chatID int64) (*domain.UserSession, error)
	UpdateTokens(chatID int64, accessToken, refreshToken string, expiresAt *time.Time) error
	Logout(chatID int64) error
	IsLoggedIn(chatID int64) (bool, error)

	// Методы для работы с OAuth состояниями
	SaveState(state string, chatID int64) error
	GetChatIDByState(state string) (int64, error)
	DeleteState(state string) error
	CleanupExpiredStates() error

	// Методы для работы с расшаренными папками
	SaveSharedFolder(chatID int64, folderName, folderPath, publicURL string) error
	GetSharedFolders(chatID int64) ([]*domain.SharedFolder, error)
}
