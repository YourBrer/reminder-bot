package configs

import (
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	BotToken    string
	DSN         string
	TgWebAppUrl string
}

func NewAppConfig() *AppConfig {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	return &AppConfig{
		BotToken:    os.Getenv("BOT_TOKEN"),
		DSN:         os.Getenv("DSN"),
		TgWebAppUrl: os.Getenv("TG_WEB_APP_URL"),
	}
}
