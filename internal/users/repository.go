package users

import "reminder/pkg/database"

type Repository struct {
	db *database.Db
}

func (r *Repository) GetById(id int64) (*User, error) {
	var user User
	res := r.db.DB.First(&user, id)

	if res.Error != nil {
		return nil, res.Error
	}

	return &user, res.Error
}

func NewRepository(db *database.Db) *Repository {
	return &Repository{db}
}
