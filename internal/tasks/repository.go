package tasks

import (
	"reminder/pkg/database"
	"time"
)

type Repository struct {
	db *database.Db
}

// GetAllByUserId Возвращает все задачи пользователя
func (r *Repository) GetAllByUserId(userId int64) ([]*Task, error) {
	var tasks []*Task
	res := r.db.DB.
		Where("user_id = ?", userId).
		Order("execution_date ASC").
		Preload("User").
		Find(&tasks)

	return tasks, res.Error
}

func (r *Repository) GetByUserAndInterval(userId int64, start, end time.Time) ([]*Task, error) {
	var tasks []*Task
	res := r.db.DB.
		Where("user_id = ? AND execution_date BETWEEN ? AND ?", userId, start, end).
		Order("execution_date ASC").
		Preload("User").
		Find(&tasks)

	return tasks, res.Error
}

func (r *Repository) GetByInterval(start, end time.Time) ([]*Task, error) {
	var tasks []*Task
	res := r.db.DB.Where("execution_date BETWEEN ? AND ?", start, end).Find(&tasks)
	return tasks, res.Error
}

func (r *Repository) Create(t *Task) (*Task, error) {
	res := r.db.DB.Create(t)
	return t, res.Error
}

func (r *Repository) SetDate(taskId int, newDate *time.Time) error {
	res := r.db.DB.Model(&Task{}).Where(taskId).Update("execution_date", newDate)
	return res.Error
}

func (r *Repository) Delete(t *Task) error {
	res := r.db.DB.Delete(t)
	return res.Error
}

func NewRepository(db *database.Db) *Repository {
	return &Repository{db}
}
