package media

import (
	"fmt"
	"io/ioutil"
	"mail_helper_bot/internal/pkg/cloud/cloud_service"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type MediaProcessor struct {
	cloudService *cloud_service.CloudService
	bufferPath   string
}

func NewMediaProcessor(bufferPath string) *MediaProcessor {
	if bufferPath == "" {
		bufferPath = "./buffer"
	}

	return &MediaProcessor{
		cloudService: cloud_service.NewCloudService(),
		bufferPath:   bufferPath,
	}
}

// ProcessChatMedia обрабатывает медиа файлы из чата
func (mp *MediaProcessor) ProcessChatMedia(accessToken string, chatID int64, chatName string) (string, error) {
	// Генерируем уникальное имя папки для чата
	folderName := mp.generateFolderName(chatID, chatName)

	// Создаем папку в облаке
	err := mp.cloudService.CreateFolder(accessToken, folderName)
	if err != nil {
		return "", fmt.Errorf("failed to create cloud folder: %v", err)
	}

	// Получаем список файлов из буферной папки
	chatBufferPath := filepath.Join(mp.bufferPath, fmt.Sprintf("%d", chatID))
	files, err := mp.getFilesFromBuffer(chatBufferPath)
	if err != nil {
		return "", fmt.Errorf("failed to get files from buffer: %v", err)
	}

	// Загружаем каждый файл в облако
	uploadedCount := 0
	for _, file := range files {
		cloudPath := filepath.Join(folderName, filepath.Base(file))
		err := mp.cloudService.UploadFile(accessToken, file, cloudPath)
		if err != nil {
			// Логируем ошибку, но продолжаем загрузку других файлов
			fmt.Printf("Failed to upload file %s: %v\n", file, err)
			continue
		}
		uploadedCount++
	}

	if uploadedCount == 0 {
		return "", fmt.Errorf("no files were uploaded")
	}

	// Создаем публичную ссылку
	publicURL, err := mp.cloudService.CreatePublicLink(accessToken, folderName)
	if err != nil {
		return "", fmt.Errorf("failed to create public link: %v", err)
	}

	// Очищаем буферную папку после успешной загрузки
	mp.cleanupBuffer(chatBufferPath)

	return publicURL, nil
}

// generateFolderName генерирует уникальное имя папки
func (mp *MediaProcessor) generateFolderName(chatID int64, chatName string) string {
	timestamp := time.Now().Format("2006-01-02_15-04")
	safeName := strings.ReplaceAll(chatName, " ", "_")
	safeName = strings.ReplaceAll(safeName, "/", "_")
	return fmt.Sprintf("chat_%d_%s_%s", chatID, safeName, timestamp)
}

// getFilesFromBuffer получает список файлов из буферной папки
func (mp *MediaProcessor) getFilesFromBuffer(bufferPath string) ([]string, error) {
	var files []string

	// Проверяем существование папки
	if _, err := os.Stat(bufferPath); os.IsNotExist(err) {
		return files, fmt.Errorf("buffer folder does not exist: %s", bufferPath)
	}

	// Читаем содержимое папки
	entries, err := ioutil.ReadDir(bufferPath)
	if err != nil {
		return files, err
	}

	// Фильтруем только медиа файлы
	allowedExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".mp4", ".avi", ".mov", ".webp"}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		for _, allowedExt := range allowedExtensions {
			if ext == allowedExt {
				files = append(files, filepath.Join(bufferPath, entry.Name()))
				break
			}
		}
	}

	return files, nil
}

// cleanupBuffer очищает буферную папку
func (mp *MediaProcessor) cleanupBuffer(bufferPath string) error {
	return os.RemoveAll(bufferPath)
}
