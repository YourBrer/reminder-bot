package tasks

import (
	"fmt"
	"reminder/internal/users"
	"time"

	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	ExecutionDate *time.Time
	Description   string
	UserId        uint
	ChatId        uint
	User          users.User `gorm:"foreignkey:UserId"`
}

// GetUserTime возвращает время напоминания со смещением по часовому поясу, заданному для пользователя
func (t *Task) GetUserTime() *time.Time {
	if t.ExecutionDate == nil || t.User.Location == "" {
		fmt.Println("не установлена дата оповещения или локаль у пользователя")
		return t.ExecutionDate
	}

	ul, err := time.LoadLocation(t.User.Location)
	if err != nil {
		fmt.Println("ошибка получения локали")
		return t.ExecutionDate
	}
	ut := t.ExecutionDate.In(ul)
	return &ut
}

func NewTask(description string, userId, chatId uint) *Task {
	return &Task{
		Description: description,
		UserId:      userId,
		ChatId:      chatId,
	}
}
