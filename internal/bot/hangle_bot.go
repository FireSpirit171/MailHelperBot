package bot

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleCommand(bot *Bot, msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		handleStartCommand(bot, msg)
	case "login":
		handleLoginCommand(bot, msg)
	case "status":
		handleStatusCommand(bot, msg)
	case "logout":
		handleLogoutCommand(bot, msg)
	default:
		reply := tgbotapi.NewMessage(msg.Chat.ID, "Неизвестная команда 🤔")
		_, err := bot.Api.Send(reply)
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}

func handleMessage(bot *Bot, msg *tgbotapi.Message) {
	session, err := bot.oauth.GetUserSession(msg.Chat.ID)
	if err != nil {
		log.Printf("Error getting session: %v", err)
		return
	}

	if session != nil && session.AccessToken != "" {
		reply := tgbotapi.NewMessage(msg.Chat.ID,
			"Вы уже авторизованы! Используйте /status для проверки статуса или /logout для выхода.")
		bot.Api.Send(reply)
		return
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID,
		"Для авторизации используйте команду /login")
	bot.Api.Send(reply)
}

func handleStartCommand(bot *Bot, msg *tgbotapi.Message) {
	text := `Привет! Я бот.

Доступные команды:
/login - Авторизация через Mail.ru
/status - Проверить статус авторизации
/logout - Выйти из аккаунта`

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	_, err := bot.Api.Send(reply)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func handleLoginCommand(bot *Bot, msg *tgbotapi.Message) {
	session, err := bot.oauth.GetUserSession(msg.Chat.ID)
	if err != nil {
		log.Printf("Error getting session: %v", err)
		return
	}

	if session != nil && session.AccessToken != "" {
		reply := tgbotapi.NewMessage(msg.Chat.ID,
			"Вы уже авторизованы! Используйте /status для проверки статуса.")
		bot.Api.Send(reply)
		return
	}

	authURL, _, err := bot.oauth.GenerateAuthURL(msg.Chat.ID)
	if err != nil {
		log.Printf("Error generating auth URL: %v", err)
		reply := tgbotapi.NewMessage(msg.Chat.ID,
			"Ошибка при создании ссылки авторизации. Попробуйте позже.")
		bot.Api.Send(reply)
		return
	}

	text := `<b>Для авторизации выполните следующие шаги:</b>

1. Нажмите на ссылку ниже
2. Разрешите доступ приложению к вашему аккаунту Mail.ru
3. После успешной авторизации вы автоматически получите сообщение в этом чате`

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ParseMode = "HTML"

	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🔐 Авторизоваться через Mail.ru", authURL),
		),
	)
	reply.ReplyMarkup = markup

	_, err = bot.Api.Send(reply)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func handleStatusCommand(bot *Bot, msg *tgbotapi.Message) {
	session, err := bot.oauth.GetUserSession(msg.Chat.ID)
	if err != nil {
		log.Printf("Error getting session: %v", err)
		return
	}

	if session == nil || session.AccessToken == "" {
		reply := tgbotapi.NewMessage(msg.Chat.ID,
			"Вы не авторизованы. Используйте /login для авторизации.")
		bot.Api.Send(reply)
		return
	}

	_, err = bot.oauth.GetUserInfo(session.AccessToken)
	if err != nil {
		newToken, err := bot.oauth.RefreshToken(session.RefreshToken)
		if err != nil {
			bot.oauth.storage.DeleteSession(msg.Chat.ID)
			reply := tgbotapi.NewMessage(msg.Chat.ID,
				"Сессия устарела. Пожалуйста, авторизуйтесь снова с помощью /login")
			bot.Api.Send(reply)
			return
		}

		userInfo, _ := bot.oauth.GetUserInfo(newToken.AccessToken)
		bot.oauth.SaveUserSession(msg.Chat.ID, newToken, userInfo)
		session.AccessToken = newToken.AccessToken
	}

	text := fmt.Sprintf("✅ Вы авторизованы!\n\n👤 Имя: %s\n📧 Email: %s",
		session.Name, session.Email)

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	bot.Api.Send(reply)
}

func handleLogoutCommand(bot *Bot, msg *tgbotapi.Message) {
	err := bot.oauth.storage.DeleteSession(msg.Chat.ID)
	if err != nil {
		log.Printf("Error deleting session: %v", err)
		reply := tgbotapi.NewMessage(msg.Chat.ID,
			"Ошибка при выходе из аккаунта.")
		bot.Api.Send(reply)
		return
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID,
		"✅ Вы успешно вышли из аккаунта.")
	bot.Api.Send(reply)
}
