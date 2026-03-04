package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"reminder/configs"
	"reminder/internal/messages"
	"reminder/internal/tasks"
	"reminder/internal/users"
	"reminder/pkg/bot"
	"reminder/pkg/database"
	"reminder/pkg/webServer"
	"syscall"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func main() {
	// Загружаем конфигурацию
	config, err := configs.NewAppConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Создаем бота
	b, err := gotgbot.NewBot(config.BotToken, nil)
	if err != nil {
		log.Fatalf("Критическая ошибка создания бота: %v", err)
	}

	// Подключаемся к БД
	db, err := database.NewDatabase(config.DSN)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Ошибка закрытия БД: %v", err)
		}
	}()

	// Инициализируем репозитории
	tasksRepo := tasks.NewRepository(db)
	usersRepo := users.NewRepository(db)
	msgRepo := messages.NewRepository(db)

	// Создаем обертку бота
	botWrapper := bot.NewWrapper(b, tasksRepo, usersRepo, config.TgWebAppUrl, 30*time.Second, msgRepo)

	// Создаем веб-сервер
	server := webServer.NewWebServer(tasksRepo, msgRepo, b, config)

	// Контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем веб-сервер в горутине
	go func() {
		if err := server.Run(ctx); err != nil {
			log.Printf("Ошибка веб-сервера: %v", err)
		}
	}()

	// Запускаем бота в горутине
	go botWrapper.Run(ctx)

	log.Println("Бот успешно запущен")

	// Ожидаем сигнал завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	<-sigChan
	log.Println("Получен сигнал завершения, останавливаем приложение...")

	// Останавливаем все сервисы
	cancel()

	// Даем время на корректное завершение
	time.Sleep(2 * time.Second)

	log.Println("Приложение остановлено")
}
