package postgres_storage

import (
	"database/sql"
	"fmt"
	"mail_helper_bot/internal/pkg/session/domain"
	"time"

	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

// ==================== Методы для работы с сессиями ====================

func (p *PostgresStorage) SaveSession(chatID int64, session *domain.UserSession) error {
	fmt.Printf("Saving session to DB: chatID=%d, name=%s, email=%s\n", chatID, session.Name, session.Email)

	_, err := p.db.Exec(`
		INSERT INTO user_sessions (chat_id, name, email, access_token, refresh_token, token_expires_at, is_logged_in)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (chat_id) DO UPDATE
		SET name = EXCLUDED.name,
		    email = EXCLUDED.email,
		    access_token = EXCLUDED.access_token,
		    refresh_token = EXCLUDED.refresh_token,
		    token_expires_at = EXCLUDED.token_expires_at,
		    is_logged_in = EXCLUDED.is_logged_in,
		    updated_at = now()
	`, chatID, session.Name, session.Email, session.AccessToken, session.RefreshToken,
		session.TokenExpiresAt, session.IsLoggedIn)

	return err
}

func (p *PostgresStorage) GetSession(chatID int64) (*domain.UserSession, error) {
	row := p.db.QueryRow(`
		SELECT chat_id, name, email, access_token, refresh_token, token_expires_at, is_logged_in, created_at, updated_at
		FROM user_sessions
		WHERE chat_id = $1 AND is_logged_in = true
	`, chatID)

	s := &domain.UserSession{}
	err := row.Scan(&s.ChatID, &s.Name, &s.Email, &s.AccessToken, &s.RefreshToken,
		&s.TokenExpiresAt, &s.IsLoggedIn, &s.CreatedAt, &s.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (p *PostgresStorage) UpdateTokens(chatID int64, accessToken, refreshToken string, expiresAt *time.Time) error {
	_, err := p.db.Exec(`
		UPDATE user_sessions 
		SET access_token = $2,
		    refresh_token = $3,
		    token_expires_at = $4,
		    updated_at = now()
		WHERE chat_id = $1
	`, chatID, accessToken, refreshToken, expiresAt)

	return err
}

func (p *PostgresStorage) Logout(chatID int64) error {
	_, err := p.db.Exec(`
		UPDATE user_sessions 
		SET is_logged_in = false,
		    access_token = NULL,
		    refresh_token = NULL,
		    token_expires_at = NULL,
		    updated_at = now()
		WHERE chat_id = $1
	`, chatID)

	return err
}

func (p *PostgresStorage) IsLoggedIn(chatID int64) (bool, error) {
	var isLoggedIn bool
	err := p.db.QueryRow(`
		SELECT is_logged_in 
		FROM user_sessions 
		WHERE chat_id = $1
	`, chatID).Scan(&isLoggedIn)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return isLoggedIn, nil
}

// ==================== Методы для работы с OAuth состояниями ====================

func (p *PostgresStorage) SaveState(state string, chatID int64) error {
	expiry := time.Now().Add(10 * time.Minute)
	_, err := p.db.Exec(`
		INSERT INTO oauth_states (state, chat_id, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (state) DO UPDATE
		SET chat_id=$2, expires_at=$3
	`, state, chatID, expiry)
	return err
}

func (p *PostgresStorage) GetChatIDByState(state string) (int64, error) {
	row := p.db.QueryRow(`
		SELECT chat_id, expires_at
		FROM oauth_states
		WHERE state=$1
	`, state)

	var chatID int64
	var expiresAt time.Time
	err := row.Scan(&chatID, &expiresAt)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	if time.Now().After(expiresAt) {
		return 0, fmt.Errorf("state expired")
	}
	return chatID, nil
}

func (p *PostgresStorage) DeleteState(state string) error {
	_, err := p.db.Exec(`DELETE FROM oauth_states WHERE state = $1`, state)
	return err
}

func (p *PostgresStorage) CleanupExpiredStates() error {
	_, err := p.db.Exec(`DELETE FROM oauth_states WHERE expires_at < now()`)
	return err
}

// ==================== Методы для работы с расшаренными папками ====================

func (p *PostgresStorage) SaveSharedFolder(chatID int64, folderName, folderPath, publicURL string) error {
	_, err := p.db.Exec(`
		INSERT INTO shared_folders (chat_id, folder_name, folder_path, public_url)
		VALUES ($1, $2, $3, $4)
	`, chatID, folderName, folderPath, publicURL)

	return err
}

func (p *PostgresStorage) GetSharedFolders(chatID int64) ([]*domain.SharedFolder, error) {
	rows, err := p.db.Query(`
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
		folder := &domain.SharedFolder{}
		err := rows.Scan(&folder.ID, &folder.ChatID, &folder.FolderName,
			&folder.FolderPath, &folder.PublicURL, &folder.CreatedAt)
		if err != nil {
			return nil, err
		}
		folders = append(folders, folder)
	}

	return folders, nil
}
