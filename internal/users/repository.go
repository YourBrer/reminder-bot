package users

import "reminder/pkg/database"

// Repository предоставляет методы для работы с пользователями в базе данных
type Repository struct {
	db *database.Db
}

// GetById возвращает пользователя по ID
func (r *Repository) GetById(id int64) (*User, error) {
	var user User
	res := r.db.DB.First(&user, id)

	if res.Error != nil {
		return nil, res.Error
	}

	return &user, res.Error
}

// NewRepository создает новый экземпляр Repository
func NewRepository(db *database.Db) *Repository {
	return &Repository{db}
}
