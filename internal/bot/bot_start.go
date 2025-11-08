package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"mail_helper_bot/internal/pkg/group/repository"
	"mail_helper_bot/internal/pkg/media"
	"mail_helper_bot/internal/pkg/oauth/oauth_service"
	"os"
	"path/filepath"
	"strings"
)

type Bot struct {
	Api            *tgbotapi.BotAPI
	oauth          *oauth_service.OAuthService
	storage        oauth_service.Storage
	groupRepo      repository.GroupRepository
	bufferPath     string
	mediaProcessor *media.MediaProcessor
}

func New(token string, storage oauth_service.Storage, groupRepo repository.GroupRepository) *Bot {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	// Создаем папку для буфера
	bufferPath := "./buffer"
	if err := os.MkdirAll(bufferPath, 0755); err != nil {
		log.Fatalf("failed to create buffer directory: %v", err)
	}

	return &Bot{
		Api:            bot,
		storage:        storage,
		groupRepo:      groupRepo,
		bufferPath:     bufferPath,
		mediaProcessor: media.NewMediaProcessor("./buffer"),
	}
}

// todo: окак(вынести в New)
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
	// Обработка команд
	log.Println("handle message:", msg)
	if msg.IsCommand() {
		log.Println("handle command:", msg)
		b.handleCommand(msg)
		return
	}

	// Обработка добавления бота в группу
	if (msg.Chat.IsGroup() || msg.Chat.IsSuperGroup()) &&
		strings.Contains(msg.Text, "@"+b.Api.Self.UserName) {
		log.Println("handle add:", msg)
		b.handleBotAddedToGroup(msg)
		return
	}
	// Обработка медиафайлов
	if b.containsMedia(msg) {
		log.Println("handle media:", msg)
		b.handleMediaMessage(msg)
	}
}

func (b *Bot) handleCommand(msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		handleStartCommand(b, msg)
	case "login":
		handleLoginCommand(b, msg)
	case "status":
		handleStatusCommand(b, msg)
	case "logout":
		handleLogoutCommand(b, msg)
	case "group_status":
		b.handleGroupStatus(msg)
	case "my_groups":
		b.handleMyGroups(msg)
	case "upload":
		handleUploadCommand(b, msg)
	default:
		reply := tgbotapi.NewMessage(msg.Chat.ID, "Неизвестная команда 🤔")
		b.Api.Send(reply)
	}
}

func (b *Bot) handleCallback(query *tgbotapi.CallbackQuery) {
	data := query.Data
	chatID := query.Message.Chat.ID
	messageID := query.Message.MessageID

	if strings.HasPrefix(data, "media_type:") {
		b.handleMediaTypeSelection(chatID, data, messageID)
	}

	// Подтверждаем обработку callback
	callback := tgbotapi.NewCallback(query.ID, "")
	b.Api.Request(callback)
}

func (b *Bot) handleChatMemberUpdate(update *tgbotapi.ChatMemberUpdated) {
	// Обрабатываем добавление/удаление бота из чата
	if update.NewChatMember.User.ID == b.Api.Self.ID {
		if update.NewChatMember.Status == "member" {
			// Бот добавлен в группу
			msg := &tgbotapi.Message{
				Chat: &update.Chat,
				From: &update.From,
				Text: "bot added",
			}
			b.handleBotAddedToGroup(msg)
		} else if update.NewChatMember.Status == "left" || update.NewChatMember.Status == "kicked" {
			// Бот удален из группы
			b.groupRepo.DeleteGroupSession(update.Chat.ID)
		}
	}
}

// Проверка наличия медиа в сообщении
func (b *Bot) containsMedia(msg *tgbotapi.Message) bool {
	return msg.Photo != nil || msg.Video != nil || msg.Document != nil
}

func (b *Bot) createGroupBufferFolder(groupID int64) (string, error) {
	groupPath := filepath.Join(b.bufferPath, string(rune(groupID)))
	if err := os.MkdirAll(groupPath, 0755); err != nil {
		return "", err
	}
	return groupPath, nil
}
