package oauth_service

import "mail_helper_bot/internal/pkg/session/domain"

type Storage interface {
	SaveSession(chatID int64, session *domain.UserSession) error
	GetSession(chatID int64) (*domain.UserSession, error)
	DeleteSession(chatID int64) error
	SaveState(state string, chatID int64) error
	GetChatIDByState(state string) (int64, error)
	CleanupExpiredStates() error
	DeleteState(state string) error
	SavePublicFolder(chatID int64, publicURL string) error
	GetPublicFolders(chatID int64) ([]string, error)
}
