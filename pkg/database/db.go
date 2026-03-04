package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Db представляет подключение к базе данных
type Db struct {
	*gorm.DB
}

// NewDatabase создает новое подключение к PostgreSQL базе данных
func NewDatabase(dsn string) (*Db, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &Db{db}, nil
}

// Close закрывает соединение с базой данных
func (d *Db) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
