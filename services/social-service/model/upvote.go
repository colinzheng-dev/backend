package model

import "time"

type Upvote struct {
	Id        string    `db:"upvote_id" json:"upvote_id"`
	UserId    string    `db:"user_id" json:"user_id"`
	ItemId    string    `db:"item_id" json:"item_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type UpvoteQuantityInfo struct {
	ItemId      string `db:"item_id" json:"item_id"`
	Quantity    int    `db:"quantity" json:"quantity"`
	UserUpvoted bool   `json:"user_upvoted,omitempty"`
}
