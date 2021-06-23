package model

// PostFixedValidate is a view of a post containing all the fixed
// fields that can be modified and require validation.
type PostFixedValidate struct {
	PostType  PostType `json:"post_type"`
	IsEdited  bool     `json:"is_edited"`
	IsDeleted bool     `json:"is_deleted"`
	Pictures  []string `json:"pictures"`
}

type ReplyFixedValidate struct {
	ParentId  string   `json:"parent_id"`
	IsEdited  bool     `json:"is_edited"`
	IsDeleted bool     `json:"is_deleted"`
	Pictures  []string `json:"pictures"`
}

// FixedValidate generates a view of a post from a database model
// that contains all the fixed fields that can be modified and require validation.
func GetPostFixedValidate(post *Post) *PostFixedValidate {
	return &PostFixedValidate{
		PostType:  post.PostType,
		IsEdited:  post.IsEdited,
		IsDeleted: post.IsDeleted,
		Pictures:  post.Pictures,
	}
}

func GetReplyFixedValidate(reply *Reply) *ReplyFixedValidate {
	return &ReplyFixedValidate{
		ParentId:  reply.ParentId,
		IsEdited:  reply.IsEdited,
		IsDeleted: reply.IsDeleted,
		Pictures:  reply.Pictures,
	}
}
