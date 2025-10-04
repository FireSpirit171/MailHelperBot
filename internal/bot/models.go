package bot

import (
	"crypto/rand"
	"encoding/hex"
)

type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type UserInfo struct {
	ID        string `json:"id"`
	ClientID  string `json:"client_id"`
	Gender    string `json:"gender"`
	Name      string `json:"name"`
	Nickname  string `json:"nickname"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Image     string `json:"image"`
}

type UserSession struct {
	ChatID       int64
	State        string
	AccessToken  string
	RefreshToken string
	Email        string
	Name         string
}

func GenerateState() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
