package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

type Bot struct {
	Api   *tgbotapi.BotAPI
	oauth *OAuthService
}

func New(token string) *Bot {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	return &Bot{
		Api: bot,
	}
}

func (b *Bot) SetOAuthService(oauth *OAuthService) {
	b.oauth = oauth
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
