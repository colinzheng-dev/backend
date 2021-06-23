package client

import "github.com/veganbase/backend/services/social-service/model"

// Client is the service client API for the social service.
type Client interface {
	GetUpvotesCount() (*map[string]model.UpvoteQuantityInfo, error)
	GetOverallRank(itemId string) (*float64, error)
	GetUserUpvotes(sessionUser string) (*map[string]bool, error)
}
