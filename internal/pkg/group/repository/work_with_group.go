package repository

import (
	"database/sql"
	"mail_helper_bot/internal/pkg/group/domain"
)

type GroupStorage struct {
	db *sql.DB
}

func NewGroupStorage(db *sql.DB) *GroupStorage {
	return &GroupStorage{db: db}
}

func (g *GroupStorage) SaveGroupSession(group *domain.GroupSession) error {
	_, err := g.db.Exec(`
        INSERT INTO group_sessions (group_id, group_title, owner_id, media_type, history_processed)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (group_id) DO UPDATE
        SET group_title = $2, media_type = $4, history_processed = $5
    `, group.GroupID, group.GroupTitle, group.OwnerID, group.MediaType, group.HistoryProcessed)
	return err
}

func (g *GroupStorage) GetGroupSession(groupID int64) (*domain.GroupSession, error) {
	row := g.db.QueryRow(`
        SELECT group_id, group_title, owner_id, media_type, created_at
        FROM group_sessions
        WHERE group_id = $1
    `, groupID)

	group := &domain.GroupSession{}
	err := row.Scan(&group.GroupID, &group.GroupTitle, &group.OwnerID, &group.MediaType, &group.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (g *GroupStorage) DeleteGroupSession(groupID int64) error {
	_, err := g.db.Exec(`DELETE FROM group_sessions WHERE group_id = $1`, groupID)
	return err
}

func (g *GroupStorage) GetUserGroups(ownerID int64) ([]*domain.GroupSession, error) {
	rows, err := g.db.Query(`
        SELECT group_id, group_title, owner_id, media_type, created_at
        FROM group_sessions
        WHERE owner_id = $1
        ORDER BY created_at DESC
    `, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*domain.GroupSession
	for rows.Next() {
		group := &domain.GroupSession{}
		err := rows.Scan(&group.GroupID, &group.GroupTitle, &group.OwnerID, &group.MediaType, &group.CreatedAt)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func (g *GroupStorage) SaveProcessedMedia(media *domain.ProcessedMedia) error {
	_, err := g.db.Exec(`
        INSERT INTO processed_media (media_id, group_id, file_id, file_name, media_type)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (media_id) DO NOTHING
    `, media.MediaID, media.GroupID, media.FileID, media.FileName, media.MediaType)
	return err
}

func (g *GroupStorage) IsMediaProcessed(mediaID string, groupID int64) (bool, error) {
	row := g.db.QueryRow(`
        SELECT 1 FROM processed_media 
        WHERE media_id = $1 AND group_id = $2
    `, mediaID, groupID)

	var exists int
	err := row.Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (g *GroupStorage) GetGroupProcessedMedia(groupID int64) ([]*domain.ProcessedMedia, error) {
	rows, err := g.db.Query(`
        SELECT media_id, group_id, file_id, file_name, media_type, created_at
        FROM processed_media
        WHERE group_id = $1
        ORDER BY created_at DESC
    `, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var media []*domain.ProcessedMedia
	for rows.Next() {
		m := &domain.ProcessedMedia{}
		err := rows.Scan(&m.MediaID, &m.GroupID, &m.FileID, &m.FileName, &m.MediaType, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		media = append(media, m)
	}
	return media, nil
}

func (g *GroupStorage) GetGroupMediaStats(groupID int64) (photosCount, videosCount int, err error) {
	row := g.db.QueryRow(`
        SELECT 
            COUNT(CASE WHEN media_type = 'photo' THEN 1 END) as photos_count,
            COUNT(CASE WHEN media_type = 'video' THEN 1 END) as videos_count
        FROM processed_media 
        WHERE group_id = $1
    `, groupID)

	err = row.Scan(&photosCount, &videosCount)
	if err != nil {
		return 0, 0, err
	}
	return photosCount, videosCount, nil
}

func (g *GroupStorage) MarkHistoryProcessed(groupID int64) error {
	_, err := g.db.Exec(`
        UPDATE group_sessions 
        SET history_processed = true 
        WHERE group_id = $1
    `, groupID)
	return err
}
