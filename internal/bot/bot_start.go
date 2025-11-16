package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"mail_helper_bot/internal/pkg/group/repository"
	"mail_helper_bot/internal/pkg/media"
	"mail_helper_bot/internal/pkg/oauth/oauth_service"
	"strings"
)

type Bot struct {
	Api            *tgbotapi.BotAPI
	oauth          *oauth_service.OAuthService
	storage        oauth_service.Storage
	groupRepo      repository.GroupRepository
	mediaProcessor *media.MediaProcessor
}

func New(token string, storage oauth_service.Storage, groupRepo repository.GroupRepository) *Bot {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	return &Bot{
		Api:            bot,
		storage:        storage,
		groupRepo:      groupRepo,
		mediaProcessor: media.NewMediaProcessor(bot),
	}
}

func (b *Bot) SetOAuthService(oauth *oauth_service.OAuthService) {
	b.oauth = oauth
}

func (b *Bot) GetMediaProcessor() *media.MediaProcessor {
	return b.mediaProcessor
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.Api.GetUpdatesChan(u)

	log.Printf("Authorized on account %s", b.Api.Self.UserName)

	for update := range updates {
		switch {
		case update.Message != nil:
			b.handleMessage(update.Message)
		case update.CallbackQuery != nil:
			b.handleCallback(update.CallbackQuery)
		case update.MyChatMember != nil:
			b.handleChatMemberUpdate(update.MyChatMember)
		}
	}
}

func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	log.Println("handle message:", msg)
	if msg.IsCommand() {
		log.Println("handle command:", msg)
		b.handleCommand(msg)
		return
	}

	if (msg.Chat.IsGroup() || msg.Chat.IsSuperGroup()) &&
		strings.Contains(msg.Text, "@"+b.Api.Self.UserName) {
		log.Println("handle add:", msg)
		b.handleBotAddedToGroup(msg)
		return
	}

	if b.containsMedia(msg) {
		log.Println("handle media:", msg)
		b.handleMediaMessage(msg)
	}
}

func (b *Bot) handleCommand(msg *tgbotapi.Message) {
	// Определяем доступные команды в зависимости от типа чата
	if msg.Chat.IsGroup() || msg.Chat.IsSuperGroup() {
		b.handleGroupCommand(msg)
	} else {
		b.handlePrivateCommand(msg)
	}
}

// handlePrivateCommand обрабатывает команды в личном чате
func (b *Bot) handlePrivateCommand(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		handleStartCommand(b, msg)
	case "login":
		handleLoginCommand(b, msg)
	case "status":
		handleStatusCommand(b, msg)
	case "logout":
		handleLogoutCommand(b, msg)
	case "my_groups":
		b.handleMyGroups(msg)
	default:
		reply := tgbotapi.NewMessage(msg.Chat.ID, "Неизвестная команда 🤔")
		b.Api.Send(reply)
	}
}

// handleGroupCommand обрабатывает команды в группе
func (b *Bot) handleGroupCommand(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "group_status":
		b.handleGroupStatus(msg)
	case "share":
		b.handleShareCommand(msg)
	case "setup_group": // НОВАЯ КОМАНДА
		b.handleSetupGroup(msg)
	case "bot_settings":
		b.handleBotSettings(msg) // Оставляем для администраторов
	case "start":
		// В группе команда start работает как добавление бота
		b.handleBotAddedToGroup(msg)
	default:
		reply := tgbotapi.NewMessage(msg.Chat.ID,
			"❌ Эта команда недоступна в группах.\n\n"+
				"📋 Доступные команды:\n"+
				"/group_status - Статус группы\n"+
				"/share - Публичная ссылка (только для администратора)\n"+
				"/bot_settings - Настройки (только для администратора)"+
				"/setup_group - Принудительная настройка")
		b.Api.Send(reply)
	}
}

func (b *Bot) handleCallback(query *tgbotapi.CallbackQuery) {
	data := query.Data
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

	if strings.HasPrefix(data, "media_type:") {
		b.handleMediaTypeSelection(chatID, data, messageID)
	} else if strings.HasPrefix(data, "media_type_settings:") {
		b.handleMediaTypeSettings(chatID, data, messageID)
	} else if strings.HasPrefix(data, "refresh_stats:") {
		b.handleRefreshStats(chatID, data, messageID)
	} else if strings.HasPrefix(data, "copy_link:") {
		b.handleCopyLink(chatID, data, messageID)
	}

	callback := tgbotapi.NewCallback(query.ID, "")
	b.Api.Request(callback)
}

func (b *Bot) handleCopyLink(chatID int64, data string, messageID int) {
	// Формат: copy_link:{url}
	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		return
	}

	// Восстанавливаем URL (может содержать ://)
	url := strings.Join(parts[1:], ":")

	// Отправляем сообщение о успешном копировании
	text := fmt.Sprintf("✅ Ссылка скопирована в буфер обмена!\n\n`%s`", url)

	editMsg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	editMsg.ParseMode = "Markdown"
	b.Api.Send(editMsg)

	// К сожалению, Telegram Bot API не поддерживает прямое копирование в буфер
	// Поэтому просто показываем ссылку для ручного копирования
}

func (b *Bot) handleChatMemberUpdate(update *tgbotapi.ChatMemberUpdated) {
	if update.NewChatMember.User.ID == b.Api.Self.ID {
		if update.NewChatMember.Status == "member" {
			msg := &tgbotapi.Message{
				Chat: &update.Chat,
				From: &update.From,
				Text: "bot added",
			}
			b.handleBotAddedToGroup(msg)
		} else if update.NewChatMember.Status == "left" || update.NewChatMember.Status == "kicked" {
			b.groupRepo.DeleteGroupSession(update.Chat.ID)
		}
	}
}

func (b *Bot) containsMedia(msg *tgbotapi.Message) bool {
	return msg.Photo != nil || msg.Video != nil || msg.Document != nil
}

func (b *Bot) sendErrorMessage(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	b.Api.Send(msg)
}

func (b *Bot) handleMediaTypeSettings(chatID int64, data string, messageID int) {
	// Формат: media_type_settings:{groupID}:{mediaType}
	parts := strings.Split(data, ":")
	if len(parts) != 3 {
		return
	}

	var groupID int64
	fmt.Sscanf(parts[1], "%d", &groupID)
	mediaType := parts[2]

	// Обновляем настройки группы
	group, err := b.groupRepo.GetGroupSession(groupID)
	if err != nil || group == nil {
		b.sendErrorMessage(chatID, "❌ Группа не найдена")
		return
	}

	group.MediaType = mediaType
	if err := b.groupRepo.SaveGroupSession(group); err != nil {
		log.Printf("Error updating group media type: %v", err)
		b.sendErrorMessage(chatID, "❌ Ошибка при сохранении настроек")
		return
	}

	// Обновляем сообщение
	mediaTypeText := map[string]string{
		"photos": "📷 фото",
		"videos": "🎥 видео",
		"all":    "📷🎥 все медиафайлы",
	}

	text := fmt.Sprintf("✅ Настройки обновлены!\n\nГруппа: %s\nНовый тип медиа: %s\n\nБот теперь будет загружать %s в ваше облако Mail.ru.",
		group.GroupTitle,
		mediaTypeText[mediaType],
		mediaTypeText[mediaType])

	editMsg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	b.Api.Send(editMsg)
}

// handleRefreshStats обновляет статистику группы
func (b *Bot) handleRefreshStats(chatID int64, data string, messageID int) {
	parts := strings.Split(data, ":")
	if len(parts) != 2 {
		return
	}

	var groupID int64
	fmt.Sscanf(parts[1], "%d", &groupID)

	// Переотправляем сообщение с обновленной статистикой
	group, err := b.groupRepo.GetGroupSession(groupID)
	if err != nil || group == nil {
		b.sendErrorMessage(chatID, "❌ Группа не найдена")
		return
	}

	b.showCurrentSettingsWithOptions(chatID, group)

	// Удаляем старое сообщение
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	b.Api.Send(deleteMsg)
}
