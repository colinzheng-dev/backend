package model

// EmailNotificationInfo is a view of user's sensitive information
// used by other services to trigger notifications via email.
type EmailNotificationInfo struct {
	Name  string `json:"name" db:"display_name"`
	Email string `json:"email" db:"email"`
}
