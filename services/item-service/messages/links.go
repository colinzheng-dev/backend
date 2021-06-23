package messages

// CreateLinkRequest represents the message sent to the API to create
// a new inter-item link.
type CreateLinkRequest struct {
	LinkType string `json:"link_type"`
	Target   string `json:"target"`
}
