package models

type Session struct {
	ID        string `db:"id"`
	UserID    string `db:"user_id"`
	Token     string `db:"token"`
	ExpiresAt int64  `db:"expires_at"`
	CreatedAt int64  `db:"created_at"`
}
