package messages

// APIKeyResponse is the response sent containing a new API key.
type APIKeyResponse struct {
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
}

type SSOSecret struct {
	Secret *string `json:"secret" db:"sso_secret"`
}