package messages

import "reminder/pkg/database"

// Repository предоставляет методы для работы с сообщениями в базе данных
type Repository struct {
	db *database.Db
}

// GetAllByTaskID возвращает все сообщения, связанные с задачей
func (r *Repository) GetAllByTaskID(taskID int) ([]*Message, error) {
	var messages []*Message
	res := r.db.DB.Where("task_id = ?", taskID).Find(&messages)
	return messages, res.Error
}

// Create создает новое сообщение в базе данных
func (r *Repository) Create(msg *Message) error {
	return r.db.DB.Create(msg).Error
}

// DeleteAllByTaskId удаляет все сообщения, связанные с задачей
func (r *Repository) DeleteAllByTaskId(taskId int) error {
	return r.db.DB.Model(&Message{}).Where("task_id = ?", taskId).Delete(&Message{}).Error
}

// NewRepository создает новый экземпляр Repository
func NewRepository(db *database.Db) *Repository {
	return &Repository{db}
}
