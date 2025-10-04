package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleCommand(api *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		reply := tgbotapi.NewMessage(msg.Chat.ID, "Необходимо авторизоваться")
		_, err := api.Send(reply)
		if err != nil {
			return
		}
	default:
		reply := tgbotapi.NewMessage(msg.Chat.ID, "Неизвестная команда 🤔")
		_, err := api.Send(reply)
		if err != nil {
			return
		}
	}
}
