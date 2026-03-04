package tasks

import (
	"errors"
	"reminder/internal/users"
	"time"

	"gorm.io/gorm"
)

// Task представляет задачу-напоминание в системе
type Task struct {
	gorm.Model
	ExecutionDate *time.Time
	Description   string
	UserId        uint
	ChatId        uint
	User          users.User `gorm:"foreignkey:UserId"`
}

// GetUserTime возвращает время напоминания со смещением по часовому поясу, заданному для пользователя
func (t *Task) GetUserTime() (*time.Time, error) {
	if t.ExecutionDate == nil {
		return nil, errors.New("дата оповещения не установлена")
	}

	if t.User.Location == "" {
		// Возвращаем UTC если локаль не установлена
		return t.ExecutionDate, nil
	}

	ul, err := time.LoadLocation(t.User.Location)
	if err != nil {
		return t.ExecutionDate, err
	}

	ut := t.ExecutionDate.In(ul)
	return &ut, nil
}

// NewTask создает новый экземпляр Task
func NewTask(description string, userId, chatId uint) *Task {
	return &Task{
		Description: description,
		UserId:      userId,
		ChatId:      chatId,
	}
}
