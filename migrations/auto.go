package main

import (
	"log"
	"os"
	"reminder/internal/messages"
	"reminder/internal/tasks"
	"reminder/internal/users"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
	db, err := gorm.Open(postgres.Open(os.Getenv("DSN")), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&tasks.Task{}, &users.User{}, &messages.Message{})
	if err != nil {
		log.Print("Database migration failed", err)
	}
}
