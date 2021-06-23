package model

import (
	"bytes"
	"encoding/json"
	"github.com/lib/pq"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/social-service/model/types"
	userModel "github.com/veganbase/backend/services/user-service/model"
	"strconv"
	"time"
)

type PostFullFixed struct {
	PostType  PostType        `json:"post_type"`
	Id        string          `json:"id"`
	Subject   string          `json:"subject"`
	Owner     userModel.Info `json:"owner"`
	IsEdited  bool            `json:"is_edited"`
	IsDeleted bool            `json:"is_deleted"`
	Pictures  pq.StringArray  `json:"pictures"`
}

type PostFull struct {
	PostFullFixed
	Attrs       chassis.GenericMap `json:"-"`
	Replies     []ReplyFull        `json:"-"`
	ReplyNumber int                `json:"-"`
	CreatedAt   time.Time          `json:"-"`
}

type ReplyFullFixed struct {
	Id        string          `json:"id"`
	ParentId  string          `json:"parent_id"`
	Owner     userModel.Info `json:"owner"`
	IsEdited  bool            `json:"is_edited"`
	Pictures  pq.StringArray  `json:"pictures"`
	IsDeleted bool            `json:"is_deleted"`
}
type ReplyFull struct {
	ReplyFullFixed
	Attrs       chassis.GenericMap `json:"-"`
	Replies     []ReplyFull        `json:"-"`
	ReplyNumber int                `json:"-"`
	CreatedAt   time.Time          `json:"-"`
}

type ThreadFull struct {
	ID           string             `json:"id"`
	Subject      string             `json:"subject"`
	Author       userModel.Info     `json:"author"`
	Content      string             `json:"content"`
	Attachments  []Attachment       `json:"attachments"`
	LockReply    bool               `json:"lock_reply"`
	Participants []userModel.Info   `json:"participants"`
	Status       types.ThreadStatus `json:"status"`
	IsEdited     bool               `json:"is_edited"`
	Messages     []MessageFull      `json:"messages"`
	CreatedAt    time.Time          `json:"created_at"`
}

type MessageFull struct {
	ID          string         `json:"id"`
	Author      userModel.Info `json:"author"`
	Content     string         `json:"content"`
	Attachments []Attachment   `json:"attachments"`
	IsEdited    bool           `json:"is_edited"`
	IsDeleted   bool           `json:"is_deleted"`
	CreatedAt   time.Time      `json:"created_at"`
}

func GetPostFull(p Post, owner *userModel.Info, replies *[]ReplyFull) *PostFull {
	return &PostFull{
		PostFullFixed: PostFullFixed{
			PostType:  p.PostType,
			Id:        p.Id,
			Subject:   p.Subject,
			Owner:     *owner,
			IsEdited:  p.IsEdited,
			IsDeleted: p.IsDeleted,
			Pictures:  p.Pictures,
		},
		Attrs:       p.Attrs,
		Replies:     *replies,
		ReplyNumber: len(*replies),
		CreatedAt:   p.CreatedAt,
	}
}

func GetThreadFull(t Thread, infoMap map[string]*userModel.Info, messages *[]MessageFull) ThreadFull {
	participants := []userModel.Info{}
	for _, p := range t.Participants {
		participants = append(participants, *infoMap[p])
	}
	return ThreadFull{
		ID:           t.ID,
		Subject:      t.Subject,
		Author:       *infoMap[t.Author],
		Content:      t.Content,
		Attachments:  t.Attachments,
		LockReply:    t.LockReply,
		Participants: participants,
		Status:       t.Status,
		IsEdited:     t.IsEdited,
		Messages:     *messages,
		CreatedAt:    t.CreatedAt,
	}
}

func GetMessageFull(m Message, userInfo *userModel.Info) MessageFull {
	return MessageFull{
		ID:          m.ID,
		Author:      *userInfo,
		Content:     m.Content,
		Attachments: m.Attachments,
		IsEdited:    m.IsEdited,
		IsDeleted:   m.IsDeleted,
		CreatedAt:   m.CreatedAt,
	}
}

func GetReplyFull(r Reply, owner *userModel.Info, replies *[]ReplyFull) *ReplyFull {
	return &ReplyFull{
		ReplyFullFixed: ReplyFullFixed{
			Id:        r.Id,
			ParentId:  r.ParentId,
			Owner:     *owner,
			IsEdited:  r.IsEdited,
			Pictures:  r.Pictures,
			IsDeleted: r.IsDeleted,
		},
		Attrs:       r.Attrs,
		Replies:     *replies,
		ReplyNumber: len(*replies),
		CreatedAt:   r.CreatedAt,
	}
}

func (view *PostFull) MarshalJSON() ([]byte, error) {
	// Marshal fixed fields.
	jsonFixed, err := json.Marshal(view.PostFullFixed)
	if err != nil {
		return nil, err
	}

	// Compose into final JSON.
	var b bytes.Buffer
	b.Write(jsonFixed[:len(jsonFixed)-1])
	if len(view.Attrs) > 0 {
		// Marshall type-specific attributes.
		jsonAttrs, err := json.Marshal(view.Attrs)
		if err != nil {
			return nil, err
		}
		b.WriteByte(',')
		b.Write(jsonAttrs[1 : len(jsonAttrs)-1])
	}
	jsonReplies, err := json.Marshal(view.Replies)
	b.Write([]byte(`, "replies:": `))
	b.Write(jsonReplies)
	b.Write([]byte(`, "reply_quantity:": ` + strconv.Itoa(view.ReplyNumber)))
	b.Write([]byte(`, "created_at":`))
	createdAtJson, err := json.Marshal(view.CreatedAt)
	b.Write(createdAtJson)
	b.WriteByte('}')
	return b.Bytes(), nil
}

func (view *ReplyFull) MarshalJSON() ([]byte, error) {
	// Marshal fixed fields.
	jsonFixed, err := json.Marshal(view.ReplyFullFixed)
	if err != nil {
		return nil, err
	}

	// Compose into final JSON.
	var b bytes.Buffer
	b.Write(jsonFixed[:len(jsonFixed)-1])
	if len(view.Attrs) > 0 {
		// Marshall type-specific attributes.
		jsonAttrs, err := json.Marshal(view.Attrs)
		if err != nil {
			return nil, err
		}
		b.WriteByte(',')
		b.Write(jsonAttrs[1 : len(jsonAttrs)-1])
	}
	jsonReplies, err := json.Marshal(view.Replies)
	b.Write([]byte(`, "replies:": `))
	b.Write(jsonReplies)
	b.Write([]byte(`, "reply_quantity:": ` + strconv.Itoa(view.ReplyNumber)))
	createdAtJson, err := json.Marshal(view.CreatedAt)
	b.Write([]byte(`, "created_at":`))
	b.Write(createdAtJson)
	b.WriteByte('}')
	return b.Bytes(), nil
}
