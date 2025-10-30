package bot

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"mail_helper_bot/internal/pkg/group/domain"
)

// Обработчик медиафайлов
func (b *Bot) handleMediaMessage(msg *tgbotapi.Message) {
	// Проверяем, есть ли настройки для этой группы
	group, err := b.groupRepo.GetGroupSession(msg.Chat.ID)
	if err != nil || group == nil {
		return
	}
	log.Println("handle media")
	// Проверяем тип медиа
	var mediaType string
	var fileID string
	var fileName string

	switch {
	case msg.Photo != nil && len(msg.Photo) > 0 && (group.MediaType == "photos" || group.MediaType == "all"):
		// Берем последнее (самое качественное) фото
		photo := msg.Photo[len(msg.Photo)-1]
		fileID = photo.FileID
		mediaType = "photo"
		fileName = fmt.Sprintf("photo_%d.jpg", time.Now().Unix())

	case msg.Video != nil && (group.MediaType == "videos" || group.MediaType == "all"):
		fileID = msg.Video.FileID
		mediaType = "video"
		fileName = fmt.Sprintf("video_%d.mp4", time.Now().Unix())
		if msg.Video.FileName != "" {
			fileName = msg.Video.FileName
		}

	case msg.Document != nil && group.MediaType == "all":
		// Проверяем, является ли документ изображением или видео
		mimeType := msg.Document.MimeType
		if strings.HasPrefix(mimeType, "image/") {
			mediaType = "photo"
			fileName = fmt.Sprintf("image_%d%s", time.Now().Unix(), filepath.Ext(msg.Document.FileName))
		} else if strings.HasPrefix(mimeType, "video/") {
			mediaType = "video"
			fileName = fmt.Sprintf("video_%d%s", time.Now().Unix(), filepath.Ext(msg.Document.FileName))
		} else {
			return // Пропускаем другие типы документов
		}
		fileID = msg.Document.FileID
		if msg.Document.FileName != "" {
			fileName = msg.Document.FileName
		}

	default:
		return
	}

	// Проверяем, не обрабатывали ли мы уже это медиа
	mediaID := fmt.Sprintf("%s_%s", mediaType, fileID)
	processed, err := b.groupRepo.IsMediaProcessed(mediaID, group.GroupID)
	if err != nil {
		log.Printf("Error checking media processing: %v", err)
		return
	}
	if processed {
		return
	}

	// Получаем информацию о файле
	file, err := b.Api.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		log.Printf("Error getting file info: %v", err)
		return
	}

	// Скачиваем файл в папку группы
	_, err = b.downloadFileToGroup(file, group.GroupID, fileName)
	if err != nil {
		log.Printf("Error downloading file: %v", err)
		return
	}

	// Помечаем как обработанное
	processedMedia := &domain.ProcessedMedia{
		MediaID:   mediaID,
		GroupID:   group.GroupID,
		FileID:    fileID,
		FileName:  fileName,
		MediaType: mediaType,
	}
	if err := b.groupRepo.SaveProcessedMedia(processedMedia); err != nil {
		log.Printf("Error saving processed media: %v", err)
	}

	log.Printf("Successfully saved media: %s to group: %d", fileName, group.GroupID)
}

// Скачивание файла в папку группы
func (b *Bot) downloadFileToGroup(file tgbotapi.File, groupID int64, fileName string) (string, error) {
	url := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", b.Api.Token, file.FilePath)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Создаем папку группы, если не существует
	groupPath := b.getGroupBufferPath(groupID)
	if err := os.MkdirAll(groupPath, 0755); err != nil {
		return "", err
	}

	// Создаем файл в папке группы
	localPath := filepath.Join(groupPath, fileName)
	out, err := os.Create(localPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return localPath, err
}

// Получает путь к папке группы
func (b *Bot) getGroupBufferPath(groupID int64) string {
	return filepath.Join(b.bufferPath, fmt.Sprintf("%d", groupID))
}
