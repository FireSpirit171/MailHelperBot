package bot

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"mail_helper_bot/internal/pkg/group/domain"
	user_session_domain "mail_helper_bot/internal/pkg/session/domain"
)

// handleSetupGroup обрабатывает команду принудительной настройки группы
func (b *Bot) handleSetupGroup(msg *tgbotapi.Message) {
	// Проверяем, что команда вызвана в группе
	if !msg.Chat.IsGroup() && !msg.Chat.IsSuperGroup() {
		b.sendErrorMessage(msg.Chat.ID, "❌ Эта команда работает только в группах и супергруппах.")
		return
	}

	// Проверяем права администратора
	member, err := b.Api.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: msg.Chat.ID,
			UserID: msg.From.ID,
		},
	})
	if err != nil {
		log.Printf("Error getting chat member: %v", err)
		b.sendErrorMessage(msg.Chat.ID, "❌ Ошибка при проверке прав.")
		return
	}

	if !b.isUserAdmin(member) {
		b.sendErrorMessage(msg.Chat.ID, "❌ Только администратор группы может настраивать бота.")
		return
	}

	// Проверяем авторизацию пользователя
	session, err := b.oauth.GetUserSession(msg.From.ID)
	if err != nil || session == nil || session.AccessToken == "" {
		b.sendErrorMessage(msg.Chat.ID,
			"❌ Для настройки группы необходимо авторизоваться.\n\n"+
				"Используйте команду /login в личном чате с ботом.")
		return
	}

	// Проверяем, есть ли уже запись для этой группы
	existingGroup, err := b.groupRepo.GetGroupSession(msg.Chat.ID)
	if err == nil && existingGroup != nil {
		// Запись уже существует - показываем текущие настройки
		b.sendGroupAlreadySetupMessage(msg.Chat.ID, existingGroup)
		return
	}

	// Создаем новую запись
	b.createGroupSession(msg.Chat, msg.From.ID, session)
}

// createGroupSession создает запись о группе
func (b *Bot) createGroupSession(chat *tgbotapi.Chat, userID int64, session *user_session_domain.UserSession) {
	// Генерируем путь к папке в облаке
	cloudFolderPath := b.mediaProcessor.GenerateCloudFolderPath(chat.ID, chat.Title)

	// Создаем запись о группе
	group := &domain.GroupSession{
		GroupID:         chat.ID,
		GroupTitle:      chat.Title,
		OwnerChatID:     userID,
		MediaType:       "photos", // по умолчанию
		CloudFolderPath: cloudFolderPath,
	}

	// Сохраняем в базу
	if err := b.groupRepo.SaveGroupSession(group); err != nil {
		log.Printf("Error saving group session: %v", err)
		b.sendErrorMessage(chat.ID, "❌ Ошибка при сохранении настроек группы.")
		return
	}

	// Пытаемся создать папку в облаке
	err := b.mediaProcessor.CreateCloudFolder(session.AccessToken, cloudFolderPath)
	if err != nil {
		log.Printf("Error creating cloud folder: %v", err)
		// Не прерываем выполнение, т.к. папка может быть создана позже
	}

	// Отправляем сообщение об успехе
	b.sendSetupSuccessMessage(chat.ID, group)
}

// sendSetupSuccessMessage отправляет сообщение об успешной настройке
func (b *Bot) sendSetupSuccessMessage(chatID int64, group *domain.GroupSession) {
	text := fmt.Sprintf(`✅ Группа успешно настроена!

📋 Информация о группе:
• Название: %s
• Тип медиа: 📷 Фото (по умолчанию)
• Облачная папка: %s
• Владелец: настроен

🎯 Что дальше:
• Используйте /bot_settings для изменения типа медиа
• Используйте /share для получения публичной ссылки
• Бот начнет загружать новые медиафайлы автоматически

⚙️ Изменить настройки:`,
		group.GroupTitle,
		group.CloudFolderPath)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📷 Только фото", fmt.Sprintf("media_type_settings:%d:photos", group.GroupID)),
			tgbotapi.NewInlineKeyboardButtonData("🎥 Только видео", fmt.Sprintf("media_type_settings:%d:videos", group.GroupID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📷🎥 Все медиа", fmt.Sprintf("media_type_settings:%d:all", group.GroupID)),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	b.Api.Send(msg)
}

// sendGroupAlreadySetupMessage отправляет сообщение, если группа уже настроена
func (b *Bot) sendGroupAlreadySetupMessage(chatID int64, group *domain.GroupSession) {
	// Получаем статистику группы
	stats, err := b.groupRepo.GetGroupMediaStats(group.GroupID)
	if err != nil {
		log.Printf("Error getting group stats: %v", err)
		stats = &domain.GroupStats{}
	}

	mediaTypeText := map[string]string{
		"photos": "📷 Фото",
		"videos": "🎥 Видео",
		"all":    "📷🎥 Все медиа",
	}

	text := fmt.Sprintf(`ℹ️ Группа уже настроена

📊 Текущие настройки:
• Название: %s
• Тип медиа: %s
• Загружено: 📷%d 🎥%d
• Облачная папка: %s

🔄 Изменить настройки:`,
		group.GroupTitle,
		mediaTypeText[group.MediaType],
		stats.PhotosCount,
		stats.VideosCount,
		group.CloudFolderPath)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📷 Только фото", fmt.Sprintf("media_type_settings:%d:photos", group.GroupID)),
			tgbotapi.NewInlineKeyboardButtonData("🎥 Только видео", fmt.Sprintf("media_type_settings:%d:videos", group.GroupID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📷🎥 Все медиа", fmt.Sprintf("media_type_settings:%d:all", group.GroupID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Обновить статистику", fmt.Sprintf("refresh_stats:%d", group.GroupID)),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	b.Api.Send(msg)
}
