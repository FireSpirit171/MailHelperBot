package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"mail_helper_bot/internal/pkg/group/domain"
	"strings"
)

func (b *Bot) handleBotAddedToGroup(msg *tgbotapi.Message) {
	chat := msg.Chat
	user := msg.From

	member, err := b.Api.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: chat.ID,
			UserID: user.ID,
		},
	})
	if err != nil {
		log.Printf("Error getting chat member: %v", err)
		return
	}

	if !b.isUserAdmin(member) {
		reply := tgbotapi.NewMessage(chat.ID,
			"❌ Только администратор группы может настроить бота.")
		b.Api.Send(reply)
		return
	}

	// Генерируем путь к папке в облаке
	cloudFolderPath := b.mediaProcessor.GenerateCloudFolderPath(chat.ID, chat.Title)

	// Сохраняем информацию о группе
	group := &domain.GroupSession{
		GroupID:         chat.ID,
		GroupTitle:      chat.Title,
		OwnerChatID:     user.ID,
		MediaType:       "photos", // по умолчанию
		CloudFolderPath: cloudFolderPath,
	}

	if err := b.groupRepo.SaveGroupSession(group); err != nil {
		log.Printf("Error saving group session: %v", err)
		reply := tgbotapi.NewMessage(chat.ID,
			"❌ Ошибка при сохранении настроек группы.")
		b.Api.Send(reply)
		return
	}

	// Отправляем сообщение с выбором типа медиа
	b.sendMediaTypeSelection(chat.ID)
}

func (b *Bot) isUserAdmin(member tgbotapi.ChatMember) bool {
	return member.Status == "creator" || member.Status == "administrator"
}

func (b *Bot) sendMediaTypeSelection(chatID int64) {
	text := `📁 Бот добавлен в группу!

Выберите тип медиа для автоматической выгрузки в облако:`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📷 Только фото", "media_type:photos"),
			tgbotapi.NewInlineKeyboardButtonData("🎥 Только видео", "media_type:videos"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📷🎥 Все медиа", "media_type:all"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	b.Api.Send(msg)
}

func (b *Bot) handleMediaTypeSelection(chatID int64, data string, messageID int) {
	parts := strings.Split(data, ":")
	if len(parts) != 2 {
		return
	}

	mediaType := parts[1]
	validTypes := map[string]string{
		"photos": "📷 Только фото",
		"videos": "🎥 Только видео",
		"all":    "📷🎥 Все медиа",
	}

	if _, valid := validTypes[mediaType]; !valid {
		return
	}

	// Обновляем настройки группы
	group, err := b.groupRepo.GetGroupSession(chatID)
	if err != nil || group == nil {
		return
	}

	group.MediaType = mediaType
	if err := b.groupRepo.SaveGroupSession(group); err != nil {
		log.Printf("Error updating group media type: %v", err)
		return
	}

	// Проверяем авторизацию владельца и создаем папку с публичной ссылкой
	session, err := b.oauth.GetUserSession(group.OwnerChatID)
	if err == nil && session != nil && session.AccessToken != "" {
		// Создаем папку в облаке
		err := b.mediaProcessor.CreateCloudFolder(session.AccessToken, group.CloudFolderPath)
		if err != nil {
			log.Printf("Error creating cloud folder: %v", err)
		} else {
			// Создаем публичную ссылку
			publicURL, err := b.mediaProcessor.CreatePublicLink(session.AccessToken, group.CloudFolderPath)
			if err != nil {
				log.Printf("Error creating public link: %v", err)
			} else {
				group.PublicURL = publicURL
				b.groupRepo.SaveGroupSession(group)
			}
		}
	}

	// Обновляем сообщение
	text := fmt.Sprintf("✅ Настройки сохранены!\n\nГруппа: %s\nТип медиа: %s\n\n☁️ Облачная папка: %s",
		group.GroupTitle, validTypes[mediaType], group.CloudFolderPath)

	if group.PublicURL != "" {
		text += fmt.Sprintf("\n\n🔗 Публичная ссылка:\n%s", group.PublicURL)
	}

	editMsg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	b.Api.Send(editMsg)

	// Отправляем инструкцию
	instruction := `📖 Инструкция:

Теперь бот будет автоматически загружать все новые медиафайлы указанного типа из этой группы прямо в ваше облако Mail.ru.

Для просмотра статуса и ссылки используйте команду /group_status`

	msg := tgbotapi.NewMessage(chatID, instruction)
	b.Api.Send(msg)
}

func (b *Bot) handleGroupStatus(msg *tgbotapi.Message) {
	group, err := b.groupRepo.GetGroupSession(msg.Chat.ID)
	if err != nil || group == nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID,
			"❌ Эта группа не настроена для выгрузки медиа.")
		b.Api.Send(reply)
		return
	}

	groupStats, err := b.groupRepo.GetGroupMediaStats(msg.Chat.ID)
	if err != nil {
		log.Printf("Error getting media stats: %v", err)
		groupStats.PhotosCount, groupStats.VideosCount = 0, 0
	}

	mediaTypeText := map[string]string{
		"photos": "📷 Только фото",
		"videos": "🎥 Только видео",
		"all":    "📷🎥 Все медиа",
	}

	text := fmt.Sprintf(`📊 Статус группы: %s

Тип медиа: %s
Загружено фото: %d
Загружено видео: %d
☁️ Облачная папка: %s`,
		group.GroupTitle,
		mediaTypeText[group.MediaType],
		groupStats.PhotosCount,
		groupStats.VideosCount,
		group.CloudFolderPath)

	// Добавляем публичную ссылку, если она есть
	if group.PublicURL != "" {
		text += fmt.Sprintf("\n\n🔗 Публичная ссылка:\n%s", group.PublicURL)
		text += "\n\n📤 Поделитесь этой ссылкой с друзьями для просмотра медиа!"
	} else {
		// Пытаемся создать публичную ссылку, если её еще нет
		session, err := b.oauth.GetUserSession(msg.Chat.ID)
		if err == nil && session != nil && session.AccessToken != "" {
			publicURL, err := b.mediaProcessor.CreatePublicLink(session.AccessToken, group.CloudFolderPath)
			if err == nil && publicURL != "" {
				group.PublicURL = publicURL
				b.groupRepo.SaveGroupSession(group)
				text += fmt.Sprintf("\n\n🔗 Публичная ссылка:\n%s", publicURL)
				text += "\n\n📤 Поделитесь этой ссылкой с друзьями для просмотра медиа!"
			}
		}
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ParseMode = "HTML"
	b.Api.Send(reply)
}

func (b *Bot) handleMyGroups(msg *tgbotapi.Message) {
	groups, err := b.groupRepo.GetUserGroups(msg.Chat.ID)
	if err != nil {
		log.Printf("Error getting user groups: %v", err)
		return
	}

	if len(groups) == 0 {
		reply := tgbotapi.NewMessage(msg.Chat.ID,
			"🤷‍♂️ Вы не управляете ни одной группой с этим ботом.")
		b.Api.Send(reply)
		return
	}

	text := "📋 Ваши группы:\n\n"
	for i, group := range groups {
		mediaTypeText := map[string]string{
			"photos": "📷",
			"videos": "🎥",
			"all":    "📷🎥",
		}

		groupStats, _ := b.groupRepo.GetGroupMediaStats(group.GroupID)
		text += fmt.Sprintf("%d. %s %s\n   ☁️ В облаке: 📷%d 🎥%d",
			i+1, mediaTypeText[group.MediaType], group.GroupTitle,
			groupStats.PhotosCount, groupStats.VideosCount)

		if group.PublicURL != "" {
			text += fmt.Sprintf("\n   🔗 Ссылка: %s", group.PublicURL)
		}
		text += "\n\n"
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ParseMode = "HTML"
	b.Api.Send(reply)
}
