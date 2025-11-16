package media

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"mail_helper_bot/internal/pkg/cloud/cloud_service"
)

type MediaInfo struct {
	FileID          string
	Type            string // "photo" or "video"
	FileName        string
	CloudFolderPath string
}

type MediaProcessor struct {
	cloudService *cloud_service.CloudService
	botAPI       *tgbotapi.BotAPI
}

func NewMediaProcessor(botAPI *tgbotapi.BotAPI) *MediaProcessor {
	return &MediaProcessor{
		cloudService: cloud_service.NewCloudService(),
		botAPI:       botAPI,
	}
}

// GenerateCloudFolderPath генерирует путь к папке в облаке для группы
func (mp *MediaProcessor) GenerateCloudFolderPath(groupID int64, groupName string) string {
	safeName := strings.ReplaceAll(groupName, " ", "_")
	safeName = strings.ReplaceAll(safeName, "/", "_")
	safeName = strings.ReplaceAll(safeName, "\\", "_")
	safeName = strings.ReplaceAll(safeName, ":", "_")
	return fmt.Sprintf("telegram_group_%d_%s", groupID, safeName)
}

// CreateCloudFolder создает папку в облаке
func (mp *MediaProcessor) CreateCloudFolder(accessToken string, folderPath string) error {
	return mp.cloudService.CreateFolder(accessToken, folderPath)
}

// CreatePublicLink создает публичную ссылку на папку
func (mp *MediaProcessor) CreatePublicLink(accessToken string, folderPath string) (string, error) {
	return mp.cloudService.CreatePublicLink(accessToken, folderPath)
}

// ProcessSingleMedia загружает одиночный медиа файл напрямую в облако
func (mp *MediaProcessor) ProcessSingleMedia(accessToken string, mediaInfo *MediaInfo) error {
	// Получаем файл из Telegram
	fileConfig := tgbotapi.FileConfig{FileID: mediaInfo.FileID}
	file, err := mp.botAPI.GetFile(fileConfig)
	if err != nil {
		return fmt.Errorf("failed to get file from Telegram: %v", err)
	}

	// Получаем URL для скачивания файла
	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", mp.botAPI.Token, file.FilePath)

	// Скачиваем файл из Telegram
	resp, err := http.Get(fileURL)
	if err != nil {
		return fmt.Errorf("failed to download file from Telegram: %v", err)
	}
	defer resp.Body.Close()

	// Читаем содержимое файла
	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read file content: %v", err)
	}

	// Формируем полный путь к файлу в облаке
	cloudFilePath := fmt.Sprintf("%s/%s", mediaInfo.CloudFolderPath, mediaInfo.FileName)

	// Загружаем файл в облако
	err = mp.cloudService.UploadFileFromBytes(accessToken, fileData, cloudFilePath)
	if err != nil {
		return fmt.Errorf("failed to upload file to cloud: %v", err)
	}

	return nil
}
