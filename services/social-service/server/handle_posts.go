package server

import (
	"errors"
	"github.com/veganbase/backend/services/social-service/events"

	"github.com/go-chi/chi"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/social-service/db"
	"github.com/veganbase/backend/services/social-service/model"
	"net/http"
)

func (s *Server) createItemPost(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get the post ID to update.
	itemId := chi.URLParam(r, "item_id")
	if itemId == "" {
		return chassis.BadRequest(w, "missing item ID")
	}

	if _, err := s.itemSvc.ItemInfo(itemId); err != nil {
		return chassis.BadRequest(w, "error validating item: "+ err.Error())
	}
	return s.createPost(w, r, itemId, true)
}

func (s *Server) createPost(w http.ResponseWriter, r *http.Request, subject string, isItem bool) (interface{}, error) {
	// Get authentication information from context and only allow
	// authenticated users to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	// Read request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Unmarshal request data -- validates JSON request body.
	post := model.Post{}
	if err = post.UnmarshalJSON(body); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Set up post ownership. Not allowed to be inserted via request,
	// a.k.a. create posts for others
	post.Owner = authInfo.UserID
	post.Subject = subject

	if post.PostType == model.ReviewPost && isItem{
		bought, err := s.purSvc.UserBoughtItem(post.Subject, post.Owner);
		if err != nil {
			return nil, err
		}
		post.Attrs["user_bought"] = bought
	}
	// Create the post.
	if err = s.db.CreatePost(&post); err != nil {
		return nil, err
	}
	chassis.Emit(s, events.PostCreated, post)

	if post.PostType == model.ReviewPost && isItem {
		chassis.Emit(s, events.ItemRankTopic, post.Subject)
	}

	// Add post/blob associations for new post.
	if len(post.Pictures) > 0 {
		if err = s.addBlobs(post.Id, post.Pictures); err != nil {
			return nil, err
		}
	}

	return &post, nil
}

func (s *Server) patchPost(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// authenticated users to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	// Get the post ID to update.
	postId := chi.URLParam(r, "post_id")
	if postId == "" {
		return chassis.BadRequest(w, "missing post ID")
	}

	// Read patch request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Look up post value and patch it.
	post, err := s.db.PostById(postId)
	if err != nil {
		if err == db.ErrPostNotFound {
			return chassis.NotFoundWithMessage(w, err.Error())
		}
		return nil, err
	}

	// Determine permission information for update.
	if post.Owner != authInfo.UserID {
		return chassis.NotFound(w)
	}

	picsBefore := map[string]bool{}
	for _, pic := range post.Pictures {
		picsBefore[pic] = true
	}

	if err = post.Patch(body); err != nil {
		return chassis.BadRequest(w, err.Error())
	}
	picsAfter := map[string]bool{}
	for _, pic := range post.Pictures {
		picsAfter[pic] = true
	}

	post.IsEdited = true
	// Do the update.
	if err = s.db.UpdatePost(post); err != nil {
		return nil, err
	}

	chassis.Emit(s, events.PostUpdated, post)

	// Update the item/blob associations.
	if err = s.updateBlobs(post.Id, picsBefore, picsAfter); err != nil {
		return nil, err
	}

	if post.PostType == model.ReviewPost {
		//TODO: HOW TO DETECT THAT THIS UPDATE IS FOR AN ITEM REVIEW?
		chassis.Emit(s, events.ItemRankTopic, post.Subject)
	}

	return &post, nil
}

func (s *Server) deletePost(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// authenticated users to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	// Get the item ID to delete.
	postId := chi.URLParam(r, "post_id")
	if postId == "" {
		return chassis.BadRequest(w, "missing item ID")
	}

	post, err := s.db.PostById(postId)
	if err != nil {
		return nil, err
	}
	if post.Owner != authInfo.UserID {
		return chassis.NotFound(w)
	}
	if post.IsDeleted {
		return chassis.BadRequest(w, "post already deleted")
	}
	// Delete the item.
	post.IsDeleted = true
	if err = s.db.UpdatePost(post); err != nil {
		return nil, err
	}

	chassis.Emit(s, events.PostDeleted, post)

	if err = s.removeBlobs(post.Id, post.Pictures); err != nil {
		return nil, err
	}


	if post.PostType == model.ReviewPost {
		//TODO: HOW TO DETECT THAT THIS UPDATE IS FOR AN ITEM REVIEW?
		chassis.Emit(s, events.ItemRankTopic, post.Subject)
	}

	return chassis.NoContent(w)
}


func (s *Server) doPostSearchForItems(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get the post ID to update.
	subject := chi.URLParam(r, "item_id")
	if subject == "" {
		return chassis.BadRequest(w, "missing item ID")
	}
	return s.doPostsSearch(w, r, subject)
}

//doPostSearch performs a generic search with all possible parameters for posts paths.
func (s *Server) doPostsSearch(w http.ResponseWriter, r *http.Request, subject string) (interface{}, error) {

	params, err := Params(r, true)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	params.Subject = &subject

	posts, total, err := s.db.GetPosts(&params.DatabaseParams)
	if err != nil {
		return nil, errors.New("SummaryPosts: " + err.Error())
	}

	// Summary views are simple.
	if params.Format == SummaryResults {
		chassis.BuildPaginationResponse(w, r, params.Pagination.Page, params.Pagination.PerPage, *total)
		return &posts, nil
	}

	fullPosts := []model.PostFull{}

	owners := map[string]bool{}
	for _, p := range *posts {
		owners[p.Owner] = true
	}

	uniqueOwners := []string{}
	for owner := range owners {
		uniqueOwners = append(uniqueOwners, owner)
	}
	//getting owner complete information
	ownerMap, err := s.userSvc.Info(uniqueOwners)
	if err != nil {
		return nil, err
	}

	// For full post views, we also need to process replies and their nested replies.
	for _, post := range *posts {
		replies, err := s.getNestedReplies(post.Id)
		if err != nil {
			return chassis.BadRequest(w, "error getting nested replies :" + err.Error())
		}
		postFull := model.GetPostFull(post, ownerMap[post.Owner], replies)
		//Masking deleted message.
		if postFull.IsDeleted == true {
			if len(postFull.Replies) > 0 {
				postFull.Attrs["content"] = "This post was deleted."
			}else {
				//if there are no nested replies, omit post
				continue
			}

		}

		fullPosts = append(fullPosts, *postFull)
	}

	chassis.BuildPaginationResponse(w, r, params.Pagination.Page, params.Pagination.PerPage, *total)
	return &fullPosts, nil
}

func (s *Server) getNestedReplies(subject string) (*[]model.ReplyFull, error) {
	fullView := []model.ReplyFull{}
	replies, err := s.db.RepliesByParentId(subject)
	if err != nil {
		return nil, errors.New("FullPosts: " + err.Error())
	}

	owners := map[string]bool{}
	for _, r := range *replies {
		owners[r.Owner] = true
	}

	uniqueOwners := []string{}
	for owner := range owners {
		uniqueOwners = append(uniqueOwners, owner)
	}

	ownerMap, err := s.userSvc.Info(uniqueOwners)
	if err != nil {
		return nil, err
	}

	for _, rpl := range *replies {
		nested, err := s.getNestedReplies(rpl.Id)
		if err != nil {
			return nil, err
		}
		fullRpl := model.GetReplyFull(rpl, ownerMap[rpl.Owner], nested)
		if fullRpl.IsDeleted == true {
			if len(fullRpl.Replies) > 0 {
				fullRpl.Attrs["content"] = "This post was deleted."
			}else {
				//if there are no nested replies, omit reply
				continue
			}

		}
		fullView = append(fullView, *fullRpl)

	}
	return &fullView, nil

}

func (s *Server) getOverallRank(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	itemId := chi.URLParam(r, "item_id")
	if itemId == "" {
		return chassis.BadRequest(w, "missing item ID")
	}
	rank, err := s.db.AvgReviewRank(itemId)
	if err != nil {
		return nil, err
	}
	return rank, nil
}