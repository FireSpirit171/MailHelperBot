package bot

import (
	"fmt"
	"log"
	"mail_helper_bot/internal/pkg/group/domain"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleShareCommand обрабатывает команду /share в группе
func (b *Bot) handleShareCommand(msg *tgbotapi.Message) {
	// Проверяем, что команда вызвана в группе
	if !msg.Chat.IsGroup() && !msg.Chat.IsSuperGroup() {
		b.sendErrorMessage(msg.Chat.ID, "❌ Эта команда работает только в группах.")
		return
	}

	// Получаем информацию о группе
	group, err := b.groupRepo.GetGroupSession(msg.Chat.ID)
	if err != nil || group == nil {
		b.sendErrorMessage(msg.Chat.ID,
			"❌ Эта группа не настроена для выгрузки медиа.\n\n"+
				"Для настройки обратитесь к администратору.")
		return
	}

	// 🔥 ПРОВЕРЯЕМ, ЯВЛЯЕТСЯ ЛИ ПОЛЬЗОВАТЕЛЬ ТЕМ, КТО ДОБАВИЛ БОТА В ГРУППУ
	if msg.From.ID != group.OwnerChatID {
		// Получаем информацию о владельце
		ownerInfo, err := b.Api.GetChat(tgbotapi.ChatInfoConfig{
			ChatConfig: tgbotapi.ChatConfig{ChatID: group.OwnerChatID},
		})

		var ownerName string
		if err == nil {
			ownerName = ownerInfo.FirstName
		} else {
			ownerName = "администратор"
		}

		b.sendErrorMessage(msg.Chat.ID,
			fmt.Sprintf("❌ Публичная ссылка доступна только пользователю, который добавил бота в группу (%s).", ownerName))
		return
	}

	// Проверяем авторизацию владельца
	session, err := b.oauth.GetUserSession(group.OwnerChatID)
	if err != nil || session == nil || session.AccessToken == "" {
		b.sendErrorMessage(msg.Chat.ID,
			"❌ Для получения публичной ссылки необходимо авторизоваться.\n\n"+
				"Перейдите в личный чат с ботом и используйте /login")
		return
	}

	// Если публичной ссылки еще нет, создаем её
	if group.PublicURL == "" {
		b.sendCreatingLinkMessage(msg.Chat.ID)

		publicURL, err := b.mediaProcessor.CreatePublicLink(session.AccessToken, group.CloudFolderPath)
		if err != nil {
			log.Printf("Error creating public link: %v", err)
			b.sendErrorMessage(msg.Chat.ID,
				"❌ Ошибка при создании публичной ссылки.\n\n"+
					"Попробуйте позже или проверьте настройки облака.")
			return
		}
		log.Printf("Public URL: %s", publicURL)
		// Сохраняем публичную ссылку
		group.PublicURL = publicURL
		if err := b.groupRepo.SaveGroupSession(group); err != nil {
			log.Printf("Error saving public URL: %v", err)
			// Не прерываем выполнение, т.к. ссылка создана
		}
	}

	// Отправляем публичную ссылку
	b.sendShareLink(msg.Chat.ID, group)
}

// sendCreatingLinkMessage отправляет сообщение о создании ссылки
func (b *Bot) sendCreatingLinkMessage(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "🔄 Создаю публичную ссылку...")
	b.Api.Send(msg)
}

// sendShareLink отправляет публичную ссылку
func (b *Bot) sendShareLink(chatID int64, group *domain.GroupSession) {
	// Получаем статистику группы
	stats, err := b.groupRepo.GetGroupMediaStats(group.GroupID)
	if err != nil {
		log.Printf("Error getting group stats: %v", err)
		stats = &domain.GroupStats{}
	}

	mediaTypeText := map[string]string{
		"photos": "📷 фото",
		"videos": "🎥 видео",
		"all":    "📷🎥 медиа",
	}

	text := fmt.Sprintf(`🔗 Публичная ссылка для группы "%s"

📊 **Статистика:**
• Загружено: 📷%d 🎥%d
• Тип контента: %s
• Облачная папка: %s

🌐 **Ссылка для доступа:**
%s

📤 Поделитесь этой ссылкой с друзьями!`,
		group.GroupTitle,
		stats.PhotosCount,
		stats.VideosCount,
		mediaTypeText[group.MediaType],
		group.CloudFolderPath,
		group.PublicURL)

	// Создаем клавиатуру с кнопкой "Поделиться"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("📤 Поделиться ссылкой", group.PublicURL),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	b.Api.Send(msg)
}

// Альтернативная версия - отправка ссылки с кнопкой копирования
func (b *Bot) sendShareLinkWithCopy(chatID int64, group *domain.GroupSession) {
	text := fmt.Sprintf(`🔗 Публичная ссылка для группы "%s"

Скопируйте ссылку ниже и поделитесь ею:\ %s\

📤 Любой, у кого есть эта ссылка, сможет просматривать загруженные медиафайлы.`,
		group.GroupTitle,
		group.PublicURL)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🔗 Открыть в браузере", group.PublicURL),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Скопировать ссылку", fmt.Sprintf("copy_link:%s", group.PublicURL)),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	b.Api.Send(msg)
}
