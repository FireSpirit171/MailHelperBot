package main

import (
	"database/sql"
	"log"
	"mail_helper_bot/internal/bot"
	groupPostgres "mail_helper_bot/internal/pkg/group/repository"
	"mail_helper_bot/internal/pkg/oauth/oauth_service"
	"mail_helper_bot/internal/pkg/session/postgres_storage"
	"mail_helper_bot/internal/pkg/web_server/web_server_service"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// ----------------- ENV -----------------
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

	dbConnStr := os.Getenv("POSTGRES_DSN")
	if dbConnStr == "" {
		dbConnStr = "postgres://mail_bot:mail_bot_pass@db:5432/mail_helper?sslmode=disable"
	}

	// ----------------- DB -----------------
	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	// ----------------- Storage -----------------
	storage := postgres_storage.NewPostgresStorage(db)
	groupStorage := groupPostgres.NewGroupStorage(db)

	// ----------------- OAuth Service -----------------
	oauthService := oauth_service.NewOAuthService(
		os.Getenv("MAIL_CLIENT_ID"),
		os.Getenv("MAIL_CLIENT_SECRET"),
		redirectURI,
		storage,
	)

	// ----------------- Bot -----------------
	b := bot.New(token, storage, groupStorage)
	b.SetOAuthService(oauthService)

	// ----------------- Web server -----------------
	webServer := web_server_service.NewWebServer(oauthService, b.Api, webPort)
	go func() {
		if err := webServer.Start(); err != nil {
			log.Fatalf("Failed to start web server: %v", err)
		}
	}()

	b.Start()
}
