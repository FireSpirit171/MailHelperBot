package bot

import (
	"fmt"
	"log"
	"mail_helper_bot/internal/pkg/group/domain"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleBotSettings обрабатывает команду настройки бота в группе
func (b *Bot) handleBotSettings(msg *tgbotapi.Message) {
	// Проверяем, что команда вызвана в группе
	if !msg.Chat.IsGroup() && !msg.Chat.IsSuperGroup() {
		reply := tgbotapi.NewMessage(msg.Chat.ID,
			"❌ Эта команда работает только в группах и супергруппах.")
		b.Api.Send(reply)
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
		reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ Ошибка при проверке прав.")
		b.Api.Send(reply)
		return
	}

	if !b.isUserAdmin(member) {
		reply := tgbotapi.NewMessage(msg.Chat.ID,
			"❌ Только администратор группы может настраивать бота.")
		b.Api.Send(reply)
		return
	}

	// 🔥 ПРОВЕРЯЕМ АВТОРИЗАЦИЮ АДМИНИСТРАТОРА
	session, err := b.oauth.GetUserSession(msg.From.ID)
	if err != nil || session == nil || session.AccessToken == "" {
		log.Printf("User %d is not authorized for group settings", msg.From.ID)

		// Отправляем сообщение в группу
		groupMsg := fmt.Sprintf(`🔐 Требуется авторизация

Администратор %s, для настройки бота необходимо авторизоваться.

Пожалуйста:

1. Перейдите в личный чат с ботом @%s
2. Используйте команду /login для авторизации
3. После авторизации вернитесь и используйте /bot_settings снова

После авторизации вы сможете настроить тип медиа для загрузки.`,
			msg.From.FirstName,
			b.Api.Self.UserName)

		reply := tgbotapi.NewMessage(msg.Chat.ID, groupMsg)
		b.Api.Send(reply)

		// Дополнительно отправляем сообщение в личный чат
		b.sendAuthRequiredMessage(msg.From.ID, msg.Chat.Title)
		return
	}

	// Проверяем, есть ли уже настройки для этой группы
	group, err := b.groupRepo.GetGroupSession(msg.Chat.ID)
	if err != nil || group == nil {
		// Группа не настроена - начинаем настройку
		b.startGroupSetupFromCommand(msg)
		return
	}

	// Группа уже настроена - показываем текущие настройки и предлагаем изменить
	b.showCurrentSettingsWithOptions(msg.Chat.ID, group)
}

// startGroupSetupFromCommand начинает настройку группы из команды /bot_settings
func (b *Bot) startGroupSetupFromCommand(msg *tgbotapi.Message) {
	cloudFolderPath := b.mediaProcessor.GenerateCloudFolderPath(msg.Chat.ID, msg.Chat.Title)

	group := &domain.GroupSession{
		GroupID:         msg.Chat.ID,
		GroupTitle:      msg.Chat.Title,
		OwnerChatID:     msg.From.ID,
		MediaType:       "photos", // по умолчанию
		CloudFolderPath: cloudFolderPath,
	}

	if err := b.groupRepo.SaveGroupSession(group); err != nil {
		log.Printf("Error saving group session: %v", err)
		reply := tgbotapi.NewMessage(msg.Chat.ID,
			"❌ Ошибка при сохранении настроек группы.")
		b.Api.Send(reply)
		return
	}

	// Отправляем выбор типа медиа
	text := fmt.Sprintf(`⚙️ Настройка бота для группы "%s"

Выберите тип медиа для автоматической загрузки в облако:`, msg.Chat.Title)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📷 Только фото", fmt.Sprintf("media_type_settings:%d:photos", msg.Chat.ID)),
			tgbotapi.NewInlineKeyboardButtonData("🎥 Только видео", fmt.Sprintf("media_type_settings:%d:videos", msg.Chat.ID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📷🎥 Все медиа", fmt.Sprintf("media_type_settings:%d:all", msg.Chat.ID)),
		),
	)

	msgConfig := tgbotapi.NewMessage(msg.Chat.ID, text)
	msgConfig.ReplyMarkup = keyboard
	b.Api.Send(msgConfig)
}

// showCurrentSettingsWithOptions показывает текущие настройки и предлагает изменить
func (b *Bot) showCurrentSettingsWithOptions(chatID int64, group *domain.GroupSession) {
	// Получаем статистику группы
	stats, err := b.groupRepo.GetGroupMediaStats(group.GroupID)
	if err != nil {
		log.Printf("Error getting group stats: %v", err)
		stats = &domain.GroupStats{}
	}

	mediaTypeText := map[string]string{
		"photos": "📷 Только фото",
		"videos": "🎥 Только видео",
		"all":    "📷🎥 Все медиа",
	}

	text := fmt.Sprintf(`⚙️ Текущие настройки группы "%s"

📊 **Статистика:**
• Тип медиа: %s
• Загружено фото: %d
• Загружено видео: %d
• Облачная папка: %s`,
		group.GroupTitle,
		mediaTypeText[group.MediaType],
		stats.PhotosCount,
		stats.VideosCount,
		group.CloudFolderPath)

	if group.PublicURL != "" {
		text += fmt.Sprintf("\n• 🔗 Публичная ссылка: %s", group.PublicURL)
	}

	text += "\n\n🔄 Изменить настройки:"

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
