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
        INSERT INTO group_sessions (group_id, group_title, owner_chat_id, media_type, cloud_folder_path, public_url, history_processed)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (group_id) DO UPDATE
        SET group_title = $2, 
            media_type = $4, 
            cloud_folder_path = $5, 
            public_url = $6, 
            history_processed = $7,
            updated_at = now()
    `, group.GroupID, group.GroupTitle, group.OwnerChatID, group.MediaType, group.CloudFolderPath, group.PublicURL, group.HistoryProcessed)
	return err
}

func (g *GroupStorage) GetGroupSession(groupID int64) (*domain.GroupSession, error) {
	row := g.db.QueryRow(`
        SELECT group_id, group_title, owner_chat_id, media_type, cloud_folder_path, COALESCE(public_url, ''), history_processed, created_at, updated_at
        FROM group_sessions
        WHERE group_id = $1
    `, groupID)

	group := &domain.GroupSession{}
	err := row.Scan(&group.GroupID, &group.GroupTitle, &group.OwnerChatID, &group.MediaType,
		&group.CloudFolderPath, &group.PublicURL, &group.HistoryProcessed, &group.CreatedAt, &group.UpdatedAt)
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

func (g *GroupStorage) GetUserGroups(ownerChatID int64) ([]*domain.GroupSession, error) {
	rows, err := g.db.Query(`
        SELECT group_id, group_title, owner_chat_id, media_type, cloud_folder_path, 
               COALESCE(public_url, ''), history_processed, created_at, updated_at
        FROM group_sessions
        WHERE owner_chat_id = $1
        ORDER BY created_at DESC
    `, ownerChatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*domain.GroupSession
	for rows.Next() {
		group := &domain.GroupSession{}
		err := rows.Scan(&group.GroupID, &group.GroupTitle, &group.OwnerChatID, &group.MediaType,
			&group.CloudFolderPath, &group.PublicURL, &group.HistoryProcessed, &group.CreatedAt, &group.UpdatedAt)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func (g *GroupStorage) SaveProcessedMedia(media *domain.ProcessedMedia) error {
	_, err := g.db.Exec(`
        INSERT INTO processed_media (group_id, file_unique_id, file_name, media_type, file_size_bytes)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (file_unique_id) DO NOTHING
    `, media.GroupID, media.FileUniqueID, media.FileName, media.MediaType, media.FileSizeBytes)
	return err
}

func (g *GroupStorage) IsMediaProcessed(fileUniqueID string, groupID int64) (bool, error) {
	row := g.db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM processed_media 
            WHERE file_unique_id = $1 AND group_id = $2
        )
    `, fileUniqueID, groupID)

	var exists bool
	err := row.Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (g *GroupStorage) GetGroupProcessedMedia(groupID int64) ([]*domain.ProcessedMedia, error) {
	rows, err := g.db.Query(`
        SELECT id, group_id, file_unique_id, file_name, media_type, file_size_bytes, uploaded_at
        FROM processed_media
        WHERE group_id = $1
        ORDER BY uploaded_at DESC
    `, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var media []*domain.ProcessedMedia
	for rows.Next() {
		m := &domain.ProcessedMedia{}
		err := rows.Scan(&m.ID, &m.GroupID, &m.FileUniqueID, &m.FileName, &m.MediaType, &m.FileSizeBytes, &m.UploadedAt)
		if err != nil {
			return nil, err
		}
		media = append(media, m)
	}
	return media, nil
}

func (g *GroupStorage) GetGroupMediaStats(groupID int64) (*domain.GroupStats, error) {
	row := g.db.QueryRow(`
        SELECT 
            COUNT(CASE WHEN media_type = 'photo' THEN 1 END) as photos_count,
            COUNT(CASE WHEN media_type = 'video' THEN 1 END) as videos_count,
            COALESCE(SUM(file_size_bytes), 0) as total_size_bytes
        FROM processed_media 
        WHERE group_id = $1
    `, groupID)

	stats := &domain.GroupStats{}
	err := row.Scan(&stats.PhotosCount, &stats.VideosCount, &stats.TotalSizeBytes)
	if err != nil {
		return nil, err
	}

	// Получаем информацию о папке
	row = g.db.QueryRow(`
        SELECT COALESCE(public_url, ''), media_type
        FROM group_sessions
        WHERE group_id = $1
    `, groupID)

	err = row.Scan(&stats.PublicURL, &stats.MediaType)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return stats, nil
}

func (g *GroupStorage) MarkHistoryProcessed(groupID int64) error {
	_, err := g.db.Exec(`
        UPDATE group_sessions 
        SET history_processed = true,
            updated_at = now()
        WHERE group_id = $1
    `, groupID)
	return err
}

// Новые методы для работы с расшаренными папками
func (g *GroupStorage) SaveSharedFolder(chatID int64, folderName, folderPath, publicURL string) error {
	_, err := g.db.Exec(`
        INSERT INTO shared_folders (chat_id, folder_name, folder_path, public_url)
        VALUES ($1, $2, $3, $4)
    `, chatID, folderName, folderPath, publicURL)
	return err
}

func (g *GroupStorage) GetUserSharedFolders(chatID int64) ([]*domain.SharedFolder, error) {
	rows, err := g.db.Query(`
        SELECT id, chat_id, folder_name, folder_path, public_url, created_at
        FROM shared_folders
        WHERE chat_id = $1
        ORDER BY created_at DESC
    `, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []*domain.SharedFolder
	for rows.Next() {
		f := &domain.SharedFolder{}
		err := rows.Scan(&f.ID, &f.ChatID, &f.FolderName, &f.FolderPath, &f.PublicURL, &f.CreatedAt)
		if err != nil {
			return nil, err
		}
		folders = append(folders, f)
	}
	return folders, nil
}
