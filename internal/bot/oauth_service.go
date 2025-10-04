package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type OAuthService struct {
	config  *OAuthConfig
	storage Storage
}

func NewOAuthService(clientID, clientSecret, redirectURI string, storage Storage) *OAuthService {
	if redirectURI == "" {
		redirectURI = "urn:ietf:wg:oauth:2.0:oob"
	}

	return &OAuthService{
		config: &OAuthConfig{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURI:  redirectURI,
		},
		storage: storage,
	}
}

func (s *OAuthService) GenerateAuthURL(chatID int64) (string, string, error) {
	state, err := GenerateState()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate state: %v", err)
	}

	err = s.storage.SaveState(state, chatID)
	if err != nil {
		return "", "", fmt.Errorf("failed to save state: %v", err)
	}

	params := url.Values{}
	params.Add("client_id", s.config.ClientID)
	params.Add("response_type", "code")
	params.Add("scope", "userinfo mail.imap")
	params.Add("redirect_uri", s.config.RedirectURI)
	params.Add("state", state)
	params.Add("prompt_force", "1")

	authURL := fmt.Sprintf("https://oauth.mail.ru/login?%s", params.Encode())
	return authURL, state, nil
}

func (s *OAuthService) ValidateState(state string) (int64, error) {
	return s.storage.GetChatIDByState(state)
}

func (s *OAuthService) ExchangeCodeForToken(code, state string) (*TokenResponse, error) {
	data := url.Values{}
	data.Add("grant_type", "authorization_code")
	data.Add("code", code)
	data.Add("redirect_uri", s.config.RedirectURI)

	req, err := http.NewRequest("POST", "https://oauth.mail.ru/token",
		strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(s.config.ClientID, s.config.ClientSecret)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oauth error: %s", string(body))
	}

	var tokenResp TokenResponse
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return nil, err
	}

	s.storage.DeleteState(state)

	return &tokenResp, nil
}

func (s *OAuthService) RefreshToken(refreshToken string) (*TokenResponse, error) {
	data := url.Values{}
	data.Add("client_id", s.config.ClientID)
	data.Add("grant_type", "refresh_token")
	data.Add("refresh_token", refreshToken)

	req, err := http.NewRequest("POST", "https://oauth.mail.ru/token",
		strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh token error: %s", string(body))
	}

	var tokenResp TokenResponse
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

func (s *OAuthService) GetUserInfo(accessToken string) (*UserInfo, error) {
	req, err := http.NewRequest("GET", "https://oauth.mail.ru/userinfo", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("access_token", accessToken)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo error: %s", string(body))
	}

	var userInfo UserInfo
	err = json.Unmarshal(body, &userInfo)
	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func (s *OAuthService) SaveUserSession(chatID int64, tokenResp *TokenResponse, userInfo *UserInfo) error {
	session := &UserSession{
		ChatID:       chatID,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		Email:        userInfo.Email,
		Name:         userInfo.Name,
	}
	return s.storage.SaveSession(chatID, session)
}

func (s *OAuthService) GetUserSession(chatID int64) (*UserSession, error) {
	return s.storage.GetSession(chatID)
}
