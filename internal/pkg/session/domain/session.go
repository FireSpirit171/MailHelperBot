package domain

import "time"

type UserSession struct {
	ChatID         int64
	Name           string
	Email          string
	AccessToken    string
	RefreshToken   string
	TokenExpiresAt *time.Time
	IsLoggedIn     bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type SharedFolder struct {
	ID         int
	ChatID     int64
	FolderName string
	FolderPath string
	PublicURL  string
	CreatedAt  time.Time
}
