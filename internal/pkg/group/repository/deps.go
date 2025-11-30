package repository

import (
	"mail_helper_bot/internal/pkg/group/domain"
)

type GroupRepository interface {
	SaveGroupSession(group *domain.GroupSession) error
	GetGroupSession(groupID int64) (*domain.GroupSession, error)
	DeleteGroupSession(groupID int64) error
	GetUserGroups(ownerID int64) ([]*domain.GroupSession, error)

	SaveProcessedMedia(media *domain.ProcessedMedia) error
	IsMediaProcessed(mediaID string, groupID int64) (bool, error)
	GetGroupProcessedMedia(groupID int64) ([]*domain.ProcessedMedia, error)
	GetGroupMediaStats(groupID int64) (stats *domain.GroupStats, err error)
	MarkHistoryProcessed(groupID int64) error
}
