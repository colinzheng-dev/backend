package db

import (
	"errors"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/social-service/model"
)

var ErrPostNotFound = errors.New("post not found")
var ErrReplyNotFound = errors.New("reply not found")
var ErrUpvoteNotFound = errors.New("upvote not found")
var ErrThreadNotFound = errors.New("thread not found")
var ErrMessageNotFound = errors.New("message not found")
var ErrAlreadyUpvoted = errors.New("already upvoted")
var ErrReadOnlyField = errors.New("attempt to modify read-only field")

type DatabaseParams struct {
	PostTypes  []string
	Pagination *chassis.Pagination
	Subject    *string
	Owner      *string
	SortBy     *chassis.Sorting
}

type DB interface {
	ListUserSubscriptions(userID string) ([]string, error)
	ListFollowers(targetID string) ([]string, error)
	CreateUserSubscription(userID, subscriptionID string) error
	DeleteUserSubscription(userID, subscriptionID string) error

	//UPVOTE-FEATURE
	CreateUpvote(upvote *model.Upvote) error
	ListUserUpvotes(userID string) (*[]string, error)
	ListWhoUpvoted(itemId string) ([]string, error)
	DeleteUpvote(upvoteId string) error
	UpvoteByUserAndItemId(userId, itemId string) (*model.Upvote, error)
	UpvoteQuantityByItemId(ids []string) (*[]model.UpvoteQuantityInfo, error)

	//POSTS-FEATURE ( generalization of REVIEW and Q&A features )
	CreatePost(post *model.Post) error
	GetPosts(params *DatabaseParams) (*[]model.Post, *uint, error)
	PostById(postId string) (*model.Post, error)
	UpdatePost(post *model.Post) error
	DeletePost(postId string) error

	CreateReply(reply *model.Reply) error
	ReplyById(replyId string) (*model.Reply, error)
	RepliesByParentId(parentId string) (*[]model.Reply, error)
	UpdateReply(reply *model.Reply) error
	DeleteReply(replyId string) error

	ThreadByID(threadID string) (*model.Thread, error)
	GetThreads(params *DatabaseParams) (*[]model.Thread, *uint, error)
	CreateThread(th *model.Thread) error
	UpdateThread(th *model.Thread) error
	ChangeThreadStatus(threadID, status string) error

	MessageByID(messageID string) (*model.Message, error)
	MessagesByParentID(parentId string) (*[]model.Message, error)
	CreateMessage(msg *model.Message) error
	UpdateMessage(msg *model.Message) error
	DeleteMessage(msgID string) error
	AuthorsByThreadID(ids []string) ([]string, error)

	AvgReviewRank(subject string) (*float64, error)

	SaveEvent(topic string, eventData interface{}, inTx func() error) error
}

//go:generate go-bindata -pkg db -o migrations.go migrations/...
