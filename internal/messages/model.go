package messages

import (
	"reminder/internal/tasks"

	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	MessageID uint
	TaskID    uint
	ChatID    uint
	Task      tasks.Task `gorm:"foreignkey:TaskID"`
}

func NewMessage(msgId, taskId, chatId uint) *Message {
	return &Message{
		MessageID: msgId,
		TaskID:    taskId,
		ChatID:    chatId,
	}
}
