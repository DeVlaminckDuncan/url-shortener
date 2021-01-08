package store

// UserToken contains the security tokens associated with a user
type UserToken struct {
	UserID string `xorm:"not null"`
	Token  []byte `xorm:"not null"`
}
