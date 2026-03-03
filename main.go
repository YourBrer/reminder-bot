package main

import (
	"reminder/configs"
	"reminder/internal/messages"
	"reminder/internal/tasks"
	"reminder/internal/users"
	"reminder/pkg/bot"
	"reminder/pkg/database"
	"reminder/pkg/webServer"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func main() {
	config := configs.NewAppConfig()

	b, err := gotgbot.NewBot(config.BotToken, nil)
	if err != nil {
		panic("Критическая ошибка создания бота: " + err.Error())
	}

	db := database.NewDatabase(config.DSN)
	tasksRepo := tasks.NewRepository(db)
	usersRepo := users.NewRepository(db)
	msgRepo := messages.NewRepository(db)
	botWrapper := bot.NewWrapper(b, tasksRepo, usersRepo, config.TgWebAppUrl, time.Second*30, msgRepo)
	server := webServer.NewWebServer(tasksRepo, msgRepo, b)

	go server.Run()
	botWrapper.Run()
}
