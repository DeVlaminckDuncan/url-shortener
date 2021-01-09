package store

// User contains a user's account information
type User struct {
	ID        string `json:"id,omitempty" xorm:"pk not null unique"`
	FirstName string `json:"firstName" binding:"required" xorm:"not null"`
	LastName  string `json:"lastName" binding:"required" xorm:"not null"`
	Username  string `json:"username" binding:"required" xorm:"not null unique"`
	Email     string `json:"email" binding:"required" xorm:"not null unique"`
	Password  string `json:"password" binding:"required" xorm:"not null"`
}
