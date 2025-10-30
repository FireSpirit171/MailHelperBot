package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"mail_helper_bot/internal/pkg/media"
	"mail_helper_bot/internal/pkg/oauth/oauth_service"
)

type Bot struct {
	Api            *tgbotapi.BotAPI
	oauth          *oauth_service.OAuthService
	storage        oauth_service.Storage
	mediaProcessor *media.MediaProcessor
}

func New(token string, storage oauth_service.Storage) *Bot {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	return &Bot{
		Api:            bot,
		storage:        storage,
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
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			handleCommand(b, update.Message)
		} else {
			handleMessage(b, update.Message)
		}
	}
}
