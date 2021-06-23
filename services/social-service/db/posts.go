package db

import (
	"database/sql"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/social-service/model"
	"math"
)

const qPostBy = `
	SELECT id, post_type, owner, subject, is_edited, pictures, attrs, is_deleted, created_at
	FROM posts `

// PostByUser returns a post by its id.
func (pg *PGClient) PostById(postId string) (*model.Post, error) {
	post:= model.Post{}
	if err := pg.DB.Get(&post, qPostBy + " WHERE id = $1", postId); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPostNotFound
		}
		return nil, err
	}
	return &post, nil
}



func (pg *PGClient) GetPosts(params *DatabaseParams) (*[]model.Post, *uint, error) {
	results := []model.Post{}
	var total uint
	q := qPostBy + paramsWhere(params) + paramsOrderBy(params)
	if params.Pagination != nil {
		q += chassis.Paginate(params.Pagination.Page, params.Pagination.PerPage)
	}

	if err := pg.DB.Select(&results, q); err != nil {
		return nil, &total, err
	}
	if err := chassis.GetTotalResultsFromQuery(q, &total, pg.DB); err != nil {
		return nil, nil, err
	}

	return &results, &total, nil
}



const qCreatePost = `
INSERT INTO
	Posts (id, post_type, owner, subject, is_edited, pictures, attrs)
VALUES (:id, :post_type, :owner, :subject, :is_edited, :pictures, :attrs)
ON CONFLICT DO NOTHING
RETURNING created_at
`

// CreatePost creates a new post
func (pg *PGClient) CreatePost(post *model.Post) error {
	tx, err := pg.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()
	post.Id = chassis.NewID(model.PostTypeIDPrefixes[post.PostType])
	rows, err := tx.NamedQuery(qCreatePost, post)
	if err != nil {
		return err
	}
	if rows.Next() {
		err = rows.Scan(&post.CreatedAt)
		if err != nil {
			return err
		}
	}
	return err
}

const qUpdatePost = `
UPDATE posts
 SET is_edited=:is_edited, is_deleted=:is_deleted,
      pictures=:pictures, attrs=:attrs
 WHERE id = :id`

// UpdatePost updates the post details in the database.
func (pg *PGClient) UpdatePost(post *model.Post) error {
	tx, err := pg.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	check := &model.Post{}
	if err = tx.Get(check, qPostBy+`id = $1`, post.Id); err != nil {
		if err == sql.ErrNoRows {
			return ErrPostNotFound
		}
		return err
	}

	// Check read-only fields.
	if post.PostType != check.PostType || post.Subject != check.Subject ||
		post.Owner != check.Owner || post.CreatedAt != check.CreatedAt {
		return ErrReadOnlyField
	}

	result, err := tx.NamedExec(qUpdatePost, post)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrPostNotFound
	}

	return err
}


const qMarkPostAsDeleted = `
UPDATE posts 
SET is_deleted = true
WHERE id = $1
`
// DeletePost mark a post as deleted, but keep its data.
func (pg *PGClient) DeletePost(postId string) error {
	tx, err := pg.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	result, err := tx.Exec(qMarkPostAsDeleted, postId)
	// Try to delete the cart item.
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrPostNotFound
	}
	return err
}


const qAvgReviewRankBySubject = `
SELECT COALESCE(AVG((attrs->'rank')::NUMERIC), 0) as rank
FROM posts 
WHERE post_type = 'review' AND subject = $1 `


// AvgReviewRank returns the average rank of a given subject
func (pg *PGClient) AvgReviewRank(subject string) (*float64, error) {
	result := struct {
		Rank float64 `db:"rank"`
	}{}
	if err := pg.DB.Get(&result, qAvgReviewRankBySubject, subject); err != nil && err != sql.ErrNoRows{
		return nil, err
	}
	roundedValue := math.Round(result.Rank*10)/10
	return &roundedValue, nil
}