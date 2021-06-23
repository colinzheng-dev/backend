package model

// Session holds the information that associates session cookies with
// users. Sessions are long-lived, and are only deleted when a user
// logs out.
type Session struct {
	Token   string `db:"token"`
	UserID  string `db:"user_id"`
	Email   string `db:"email"`
	IsAdmin bool   `db:"is_admin"`
}
