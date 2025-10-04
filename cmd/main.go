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

	webPort := os.Getenv("WEB_PORT")
	if webPort == "" {
		webPort = "8080"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		log.Fatal("BASE_URL must be set")
	}

	redirectURI := baseURL + "/oauth/callback/"

	storage := bot.NewMemoryStorage()

	b := bot.New(token)
	oauthService := bot.NewOAuthService(
		os.Getenv("MAIL_CLIENT_ID"),
		os.Getenv("MAIL_CLIENT_SECRET"),
		redirectURI,
		storage,
	)

	b.SetOAuthService(oauthService)

	webServer := bot.NewWebServer(oauthService, b.Api, webPort)
	go func() {
		if err := webServer.Start(); err != nil {
			log.Fatalf("Failed to start web server: %v", err)
		}
	}()

	b.Start()
}
