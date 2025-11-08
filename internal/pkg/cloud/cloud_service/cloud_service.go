package cloud_service

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mail_helper_bot/internal/pkg/http_client"
	"net/http"
	"os"
	"strings"
)

const (
	//baseAPIURL = "https://openapi.cloud.mail.ru"
	baseAPIURL = "http://mock-api:8082"
)

type CloudService struct {
	client *http_client.LoggedClient
}

type CloudFolder struct {
	Name      string
	Path      string
	PublicURL string
}

type UploadResponse struct {
	Email  string `json:"email"`
	Body   string `json:"body"`
	Time   int64  `json:"time"`
	Status int    `json:"status"`
}

type ShareResponse struct {
	Email string `json:"email"`
	Body  struct {
		Url string `json:"url"`
	} `json:"body"`
	Time   int64 `json:"time"`
	Status int   `json:"status"`
}

func NewCloudService() *CloudService {
	logServerURL := os.Getenv("LOG_SERVER_URL")
	return &CloudService{
		client: http_client.NewLoggedClient(logServerURL),
	}
}

// CreateFolder создает папку в облаке
func (cs *CloudService) CreateFolder(accessToken, folderPath string) error {
	url := fmt.Sprintf("%s/api/v1/private/mkdir/%s", baseAPIURL, folderPath)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Добавляем токен авторизации
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := cs.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create folder: status=%d, body=%s", resp.StatusCode, string(body))
	}

	return nil
}

// UploadFile загружает файл в облако
func (cs *CloudService) UploadFile(accessToken, localFilePath, cloudPath string) error {
	// Читаем файл
	fileData, err := os.ReadFile(localFilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// Вычисляем хеш файла
	hasher := sha1.New()
	hasher.Write(fileData)
	fileHash := strings.ToUpper(hex.EncodeToString(hasher.Sum(nil)))

	// Получаем информацию о файле
	fileInfo, err := os.Stat(localFilePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	// Подготавливаем данные для загрузки
	uploadData := map[string]interface{}{
		"hash":          fileHash,
		"size":          fileInfo.Size(),
		"path":          cloudPath,
		"overwrite":     true,
		"last_modified": fileInfo.ModTime().Unix(),
	}

	jsonData, err := json.Marshal(uploadData)
	if err != nil {
		return fmt.Errorf("failed to marshal upload data: %v", err)
	}

	// Создаем запрос
	url := fmt.Sprintf("%s/api/v1/private/add", baseAPIURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос
	resp, err := cs.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to upload file: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// Теперь нужно загрузить сам файл
	err = cs.uploadFileContent(accessToken, fileData, cloudPath)
	if err != nil {
		return fmt.Errorf("failed to upload file content: %v", err)
	}

	return nil
}

// uploadFileContent загружает содержимое файла
func (cs *CloudService) uploadFileContent(accessToken string, fileData []byte, cloudPath string) error {
	// Для загрузки файлов может потребоваться специальный endpoint
	// Это примерная реализация, может потребоваться корректировка
	uploadURL := fmt.Sprintf("%s/api/v1/private/upload", baseAPIURL)

	req, err := http.NewRequest("PUT", uploadURL, bytes.NewReader(fileData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-Cloud-Path", cloudPath)

	resp, err := cs.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// CreatePublicLink создает публичную ссылку на папку
func (cs *CloudService) CreatePublicLink(accessToken, folderPath string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/private/share/%s", baseAPIURL, folderPath)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := cs.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to create public link: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var shareResp ShareResponse
	if err := json.Unmarshal(body, &shareResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	return shareResp.Body.Url, nil
}

// RemovePublicLink удаляет публичную ссылку
func (cs *CloudService) RemovePublicLink(accessToken, folderPath string) error {
	url := fmt.Sprintf("%s/api/v1/private/unshare/%s", baseAPIURL, folderPath)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := cs.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to remove public link: status=%d, body=%s", resp.StatusCode, string(body))
	}

	return nil
}
