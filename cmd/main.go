package main

import (
	"github.com/joho/godotenv"
	"log"
	"mail_helper_bot/internal/bot"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_TOKEN is not set")
	}

	// Получаем настройки веб-сервера
	webPort := os.Getenv("WEB_PORT")
	if webPort == "" {
		webPort = "8080"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		log.Fatal("BASE_URL must be set")
	}

	redirectURI := baseURL + "/oauth/callback/"

	// Инициализация хранилища
	storage := bot.NewMemoryStorage()

	// Создаем бота
	b := bot.New(token)
	// Инициализируем OAuth service с redirect_uri
	oauthService := bot.NewOAuthService(
		os.Getenv("MAIL_CLIENT_ID"),
		os.Getenv("MAIL_CLIENT_SECRET"),
		redirectURI,
		storage,
	)

	// Устанавливаем OAuth service в бота
	b.SetOAuthService(oauthService)

	// Запускаем веб-сервер в горутине
	webServer := bot.NewWebServer(oauthService, b.Api, webPort)
	go func() {
		if err := webServer.Start(); err != nil {
			log.Fatalf("Failed to start web server: %v", err)
		}
	}()

	// Запускаем бота
	b.Start()
}
