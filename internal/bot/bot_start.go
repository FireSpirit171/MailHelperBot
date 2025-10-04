package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

type Bot struct {
	api *tgbotapi.BotAPI
}

func New(token string) *Bot {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	return &Bot{api: bot}
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	log.Printf("Authorized on account %s", b.api.Self.UserName)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			handleCommand(b.api, update.Message)
		}
	}
}
