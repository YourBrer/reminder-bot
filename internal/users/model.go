package users

type User struct {
	ID       uint
	Username string
	Location string `gorm:"default:'Europe/Samara'"`
}
