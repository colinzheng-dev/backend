package messages

// LoginRequest is a structure representing the request body for login
// requests.
type LoginRequest struct {
	Email    string `json:"email"`
	Site     string `json:"site"`
	Language string `json:"language"`
}
