package bot

//
//import (
//	"fmt"
//	"log"
//	"time"
//
//	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
//)
//
//// Запуск обработки истории группы
//func (b *Bot) processGroupHistory(groupID int64, mediaType string) {
//	log.Printf("Starting history processing for group %d, media type: %s", groupID, mediaType)
//
//	// Отправляем сообщение о начале обработки
//	msg := tgbotapi.NewMessage(groupID, "🔄 Начинаю выгрузку истории медиафайлов...")
//	b.Api.Send(msg)
//
//	offset := 0
//	processedCount := 0
//	totalProcessed := 0
//
//	for {
//		// Получаем историю сообщений (пачками по 100)
//		messages, err := b.getChatHistory(groupID, offset, 100)
//		if err != nil {
//			log.Printf("Error getting chat history: %v", err)
//			b.sendHistoryStatus(groupID, totalProcessed, true, err.Error())
//			return
//		}
//
//		if len(messages) == 0 {
//			break // Достигли конца истории
//		}
//
//		// Обрабатываем сообщения в этой пачке
//		for _, message := range messages {
//			if b.shouldProcessMedia(&message, mediaType) {
//				b.handleMediaMessage(&message)
//				processedCount++
//				totalProcessed++
//
//				// Обновляем статус каждые 10 обработанных файлов
//				if processedCount >= 10 {
//					b.sendHistoryStatus(groupID, totalProcessed, false, "")
//					processedCount = 0
//					time.Sleep(1 * time.Second) // Задержка чтобы не превысить лимиты API
//				}
//			}
//		}
//
//		offset += len(messages)
//
//		// Добавляем задержку между запросами
//		time.Sleep(500 * time.Millisecond)
//	}
//
//	// Завершаем обработку
//	b.sendHistoryStatus(groupID, totalProcessed, true, "")
//	b.groupRepo.MarkHistoryProcessed(groupID)
//
//	log.Printf("History processing completed for group %d. Total processed: %d", groupID, totalProcessed)
//}
//
//// Получение истории сообщений
//func (b *Bot) getChatHistory(chatID int64, offset, limit int) ([]tgbotapi.Message, error) {
//	config := tgbotapi.ChatHistoryConfig{
//		ChatID: chatID,
//		Offset: offset,
//		Limit:  limit,
//	}
//
//	messages, err := b.Api.GetChatHistory(config)
//	if err != nil {
//		return nil, err
//	}
//
//	return messages, nil
//}
//
//// Проверка, нужно ли обрабатывать медиа в сообщении
//func (b *Bot) shouldProcessMedia(message *tgbotapi.Message, mediaType string) bool {
//	switch {
//	case message.Photo != nil && len(message.Photo) > 0:
//		return mediaType == "photos" || mediaType == "all"
//	case message.Video != nil:
//		return mediaType == "videos" || mediaType == "all"
//	case message.Document != nil:
//		if mediaType == "all" {
//			mimeType := message.Document.MimeType
//			return mimeType != "" && (mimeType[:5] == "image" || mimeType[:5] == "video")
//		}
//	}
//	return false
//}
//
//// Отправка статуса обработки
//func (b *Bot) sendHistoryStatus(chatID int64, processed int, finished bool, errorMsg string) {
//	var text string
//
//	if errorMsg != "" {
//		text = fmt.Sprintf("❌ Ошибка при выгрузке истории: %s\n\nОбработано файлов: %d", errorMsg, processed)
//	} else if finished {
//		text = fmt.Sprintf("✅ Выгрузка истории завершена!\n\nВсего обработано файлов: %d", processed)
//	} else {
//		text = fmt.Sprintf("🔄 Выгружаю историю...\n\nОбработано файлов: %d", processed)
//	}
//
//	msg := tgbotapi.NewMessage(chatID, text)
//	b.Api.Send(msg)
//}
