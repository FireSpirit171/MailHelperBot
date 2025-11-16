package bot

import (
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"mail_helper_bot/internal/pkg/group/domain"
	"mail_helper_bot/internal/pkg/media"
)

func (b *Bot) handleMediaMessage(msg *tgbotapi.Message) {
	// Проверяем, есть ли настройки для этой группы
	group, err := b.groupRepo.GetGroupSession(msg.Chat.ID)
	if err != nil || group == nil {
		return
	}

	log.Println("handle media")

	// Проверяем авторизацию владельца группы
	session, err := b.oauth.GetUserSession(msg.Chat.ID)
	if err != nil || session == nil || session.AccessToken == "" {
		log.Printf("Owner not authorized for group %d. msg.Chat.ID = %d", group.GroupID, msg.Chat.ID)
		return
	}

	// Определяем тип медиа и собираем информацию
	var mediaInfo *media.MediaInfo

	switch {
	case msg.Photo != nil && len(msg.Photo) > 0 && (group.MediaType == "photos" || group.MediaType == "all"):
		photo := msg.Photo[len(msg.Photo)-1] // Берем самое качественное фото
		mediaInfo = &media.MediaInfo{
			FileID:          photo.FileID,
			Type:            "photo",
			FileName:        fmt.Sprintf("photo_%d.jpg", time.Now().Unix()),
			CloudFolderPath: group.CloudFolderPath,
		}

	case msg.Video != nil && (group.MediaType == "videos" || group.MediaType == "all"):
		fileName := fmt.Sprintf("video_%d.mp4", time.Now().Unix())
		if msg.Video.FileName != "" {
			fileName = msg.Video.FileName
		}
		mediaInfo = &media.MediaInfo{
			FileID:          msg.Video.FileID,
			Type:            "video",
			FileName:        fileName,
			CloudFolderPath: group.CloudFolderPath,
		}

	case msg.Document != nil && group.MediaType == "all":
		mimeType := msg.Document.MimeType
		var mediaType string

		if strings.HasPrefix(mimeType, "image/") {
			mediaType = "photo"
		} else if strings.HasPrefix(mimeType, "video/") {
			mediaType = "video"
		} else {
			return
		}

		mediaInfo = &media.MediaInfo{
			FileID:          msg.Document.FileID,
			Type:            mediaType,
			FileName:        msg.Document.FileName,
			CloudFolderPath: group.CloudFolderPath,
		}

	default:
		return
	}

	// Проверяем, не обрабатывали ли мы уже это медиа
	mediaID := fmt.Sprintf("%s_%s", mediaInfo.Type, mediaInfo.FileID)
	processed, err := b.groupRepo.IsMediaProcessed(mediaID, group.GroupID)
	if err != nil {
		log.Printf("Error checking media processing: %v", err)
		return
	}
	if processed {
		return
	}

	// Если публичной ссылки еще нет, создаем её
	if group.PublicURL == "" {
		// Создаем папку в облаке если её нет
		err := b.mediaProcessor.CreateCloudFolder(session.AccessToken, group.CloudFolderPath)
		if err != nil {
			log.Printf("Error creating cloud folder: %v", err)
		}

		// Создаем публичную ссылку
		publicURL, err := b.mediaProcessor.CreatePublicLink(session.AccessToken, group.CloudFolderPath)
		if err == nil && publicURL != "" {
			group.PublicURL = publicURL
			b.groupRepo.SaveGroupSession(group)

			// Уведомляем в чате о создании публичной ссылки
			notifyMsg := fmt.Sprintf("✅ Создана публичная ссылка для папки группы:\n%s", publicURL)
			reply := tgbotapi.NewMessage(msg.Chat.ID, notifyMsg)
			b.Api.Send(reply)
		}
	}

	// Загружаем медиа напрямую в облако
	err = b.mediaProcessor.ProcessSingleMedia(session.AccessToken, mediaInfo)
	if err != nil {
		log.Printf("Error uploading media to cloud: %v", err)
		return
	}

	// Помечаем как обработанное
	processedMedia := &domain.ProcessedMedia{
		ID:           mediaID,
		GroupID:      group.GroupID,
		FileUniqueID: mediaInfo.FileID,
		FileName:     mediaInfo.FileName,
		MediaType:    mediaInfo.Type,
	}

	if err := b.groupRepo.SaveProcessedMedia(processedMedia); err != nil {
		log.Printf("Error saving processed media: %v", err)
	}

	log.Printf("Successfully uploaded media: %s to cloud folder: %s", mediaInfo.FileName, group.CloudFolderPath)
}
