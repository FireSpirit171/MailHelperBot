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

	mailBot := bot.New(token)
	mailBot.Start()
}
