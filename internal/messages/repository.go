package messages

import "reminder/pkg/database"

type Repository struct {
	db *database.Db
}

func (r *Repository) GetAllByTaskID(taskID int) ([]*Message, error) {
	var messages []*Message
	res := r.db.DB.Where("task_id = ?", taskID).Find(&messages)
	return messages, res.Error
}

func (r *Repository) Create(msg *Message) error {
	return r.db.DB.Create(msg).Error
}

func (r *Repository) DeleteAllByTaskId(taskId int) error {
	return r.db.DB.Model(&Message{}).Where("task_id = ?", taskId).Delete(&Message{}).Error
}

func NewRepository(db *database.Db) *Repository {
	return &Repository{db}
}
