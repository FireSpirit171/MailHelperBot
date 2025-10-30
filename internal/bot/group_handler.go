package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"mail_helper_bot/internal/pkg/group/domain"
	"strings"
)

// Обработчик добавления бота в группу
func (b *Bot) handleBotAddedToGroup(msg *tgbotapi.Message) {
	chat := msg.Chat
	user := msg.From

	// Проверяем, является ли пользователь администратором
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

	// Проверяем, что пользователь является создателем или администратором
	if !b.isUserAdmin(member) {
		reply := tgbotapi.NewMessage(chat.ID,
			"❌ Только администратор группы может настроить бота.")
		b.Api.Send(reply)
		return
	}

	// Создаем папку для группы в буфере
	if _, err := b.createGroupBufferFolder(chat.ID); err != nil {
		log.Printf("Error creating group buffer folder: %v", err)
		reply := tgbotapi.NewMessage(chat.ID,
			"❌ Ошибка при создании папки для группы.")
		b.Api.Send(reply)
		return
	}

	// Сохраняем информацию о группе
	group := &domain.GroupSession{
		GroupID:    chat.ID,
		GroupTitle: chat.Title,
		OwnerID:    user.ID,
		MediaType:  "photos", // по умолчанию
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

// Проверка прав администратора
func (b *Bot) isUserAdmin(member tgbotapi.ChatMember) bool {
	return member.Status == "creator" || member.Status == "administrator"
}

// Отправка клавиатуры выбора типа медиа
func (b *Bot) sendMediaTypeSelection(chatID int64) {
	text := `📁 Бот добавлен в группу!

Выберите тип медиа для выгрузки в локальную папку:`

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

// Обработчик выбора типа медиа
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

	// Обновляем сообщение
	text := fmt.Sprintf("✅ Настройки сохранены!\n\nГруппа: %s\nТип медиа: %s\n\nПапка: buffers/%d",
		group.GroupTitle, validTypes[mediaType], group.GroupID)

	editMsg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	b.Api.Send(editMsg)

	// Отправляем инструкцию
	instruction := `📖 Инструкция:

Теперь бот будет автоматически сохранять все новые медиафайлы указанного типа из этой группы в локальную папку.

Для просмотра статуса используйте команду /group_status`

	msg := tgbotapi.NewMessage(chatID, instruction)
	b.Api.Send(msg)
}

// Команда для просмотра статуса группы
func (b *Bot) handleGroupStatus(msg *tgbotapi.Message) {
	group, err := b.groupRepo.GetGroupSession(msg.Chat.ID)
	if err != nil || group == nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID,
			"❌ Эта группа не настроена для выгрузки медиа.")
		b.Api.Send(reply)
		return
	}

	photosCount, videosCount, err := b.groupRepo.GetGroupMediaStats(msg.Chat.ID)
	if err != nil {
		log.Printf("Error getting media stats: %v", err)
		photosCount, videosCount = 0, 0
	}

	mediaTypeText := map[string]string{
		"photos": "📷 Только фото",
		"videos": "🎥 Только видео",
		"all":    "📷🎥 Все медиа",
	}

	text := fmt.Sprintf(`📊 Статус группы: %s

Тип медиа: %s
Сохранено фото: %d
Сохранено видео: %d
Локальная папка: buffers/%d`,
		group.GroupTitle,
		mediaTypeText[group.MediaType],
		photosCount,
		videosCount,
		group.GroupID)

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	b.Api.Send(reply)
}

// Команда для управления группами
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

		photosCount, videosCount, _ := b.groupRepo.GetGroupMediaStats(group.GroupID)
		text += fmt.Sprintf("%d. %s %s\n   📁 buffers/%d | 📷%d 🎥%d\n\n",
			i+1, mediaTypeText[group.MediaType], group.GroupTitle,
			group.GroupID, photosCount, videosCount)
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	b.Api.Send(reply)
}
