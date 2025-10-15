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

// ------------------ Сессии ------------------

func (p *PostgresStorage) SaveSession(chatID int64, session *domain.UserSession) error {
	fmt.Printf("Saving session to DB: chatID=%d, name=%s, email=%s\n", chatID, session.Name, session.Email)
	_, err := p.db.Exec(`
		INSERT INTO user_sessions (chat_id, name, email, access_token, refresh_token)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (chat_id) DO UPDATE
		SET name=$2, email=$3, access_token=$4, refresh_token=$5
	`, chatID, session.Name, session.Email, session.AccessToken, session.RefreshToken)
	return err
}

func (p *PostgresStorage) GetSession(chatID int64) (*domain.UserSession, error) {
	row := p.db.QueryRow(`
		SELECT chat_id, name, email, access_token, refresh_token
		FROM user_sessions
		WHERE chat_id=$1
	`, chatID)

	s := &domain.UserSession{}
	err := row.Scan(&s.ChatID, &s.Name, &s.Email, &s.AccessToken, &s.RefreshToken)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (p *PostgresStorage) DeleteSession(chatID int64) error {
	_, err := p.db.Exec(`DELETE FROM user_sessions WHERE chat_id=$1`, chatID)
	return err
}

// ------------------ State ------------------

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
	_, err := p.db.Exec(`DELETE FROM oauth_states WHERE state=$1`, state)
	return err
}

func (p *PostgresStorage) CleanupExpiredStates() error {
	_, err := p.db.Exec(`DELETE FROM oauth_states WHERE expires_at < NOW()`)
	return err
}
