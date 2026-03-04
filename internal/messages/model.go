package messages

import (
	"reminder/internal/tasks"

	"gorm.io/gorm"
)

// Message представляет связь между Telegram сообщением и задачей
type Message struct {
	gorm.Model
	MessageID uint
	TaskID    uint
	ChatID    uint
	Task      tasks.Task `gorm:"foreignkey:TaskID"`
}

// NewMessage создает новый экземпляр Message
func NewMessage(msgId, taskId, chatId uint) *Message {
	return &Message{
		MessageID: msgId,
		TaskID:    taskId,
		ChatID:    chatId,
	}
}
