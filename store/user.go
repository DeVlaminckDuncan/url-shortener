package store

// User contains a user's account information
type User struct {
	ID        string `xorm:"pk not null unique"`
	FirstName string `xorm:"not null"`
	LastName  string `xorm:"not null"`
	Username  string `xorm:"not null unique"`
	Email     string `xorm:"not null unique"`
	Password  string `xorm:"not null"`
}
