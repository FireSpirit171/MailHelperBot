package main

import (
	"github.com/joho/godotenv"
	"log"
	"mail_helper_bot/internal/bot"
	"mail_helper_bot/internal/pkg/oauth/oauth_service"
	"mail_helper_bot/internal/pkg/session/usecase"
	"mail_helper_bot/internal/pkg/web_server/web_server_service"
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

	webPort := os.Getenv("WEB_PORT")
	if webPort == "" {
		webPort = "8080"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		log.Fatal("BASE_URL must be set")
	}

	redirectURI := baseURL + "/oauth/callback/"

	storage := usecase.NewMemoryStorage()

	b := bot.New(token, storage)
	oauthService := oauth_service.NewOAuthService(
		os.Getenv("MAIL_CLIENT_ID"),
		os.Getenv("MAIL_CLIENT_SECRET"),
		redirectURI,
		storage,
	)

	b.SetOAuthService(oauthService)

	webServer := web_server_service.NewWebServer(oauthService, b.Api, webPort)
	go func() {
		if err := webServer.Start(); err != nil {
			log.Fatalf("Failed to start web server: %v", err)
		}
	}()

	b.Start()
}
