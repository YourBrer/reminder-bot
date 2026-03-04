package users

// User представляет пользователя Telegram в системе
type User struct {
	ID       uint
	Username string
	Location string `gorm:"default:'Europe/Samara'"`
}
