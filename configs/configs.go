package configs

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

// AppConfig содержит конфигурацию приложения
type AppConfig struct {
	BotToken    string
	DSN         string
	TgWebAppUrl string
	ServerPort  string
	TLSCertPath string
	TLSKeyPath  string
	UseTLS      bool
}

// NewAppConfig загружает и валидирует конфигурацию из переменных окружения
func NewAppConfig() (*AppConfig, error) {
	// Загружаем .env файл, но не паникуем если его нет (может быть в production)
	_ = godotenv.Load()

	config := &AppConfig{
		BotToken:    os.Getenv("BOT_TOKEN"),
		DSN:         os.Getenv("DSN"),
		TgWebAppUrl: os.Getenv("TG_WEB_APP_URL"),
		ServerPort:  getEnvWithDefault("SERVER_PORT", "3000"),
		TLSCertPath: os.Getenv("TLS_CERT_PATH"),
		TLSKeyPath:  os.Getenv("TLS_KEY_PATH"),
		UseTLS:      os.Getenv("USE_TLS") == "true",
	}

	// Валидация обязательных параметров
	if config.BotToken == "" {
		return nil, errors.New("BOT_TOKEN is required")
	}
	if config.DSN == "" {
		return nil, errors.New("DSN is required")
	}
	if config.TgWebAppUrl == "" {
		return nil, errors.New("TG_WEB_APP_URL is required")
	}

	// Если используется TLS, проверяем наличие сертификатов
	if config.UseTLS {
		if config.TLSCertPath == "" || config.TLSKeyPath == "" {
			return nil, errors.New("TLS_CERT_PATH and TLS_KEY_PATH are required when USE_TLS=true")
		}
	}

	return config, nil
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
