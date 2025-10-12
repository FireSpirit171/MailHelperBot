package domain

type UserSession struct {
	ChatID       int64
	State        string
	AccessToken  string
	RefreshToken string
	Email        string
	Name         string
}
