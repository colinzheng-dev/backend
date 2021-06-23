package server

import (
	"github.com/go-chi/chi"
	"github.com/veganbase/backend/services/social-service/events"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/social-service/db"
	"github.com/veganbase/backend/services/social-service/model"
	"net/http"
)

func (s *Server) createReply(w http.ResponseWriter, r *http.Request) (interface{}, error) {
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
	reply := model.Reply{}
	if err = reply.UnmarshalJSON(body); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Set up post ownership. Not allowed to be inserted via request,
	// a.k.a. create posts for others
	reply.Owner = authInfo.UserID

	// Create the post.
	if err = s.db.CreateReply(&reply); err != nil {
		return nil, err
	}
	chassis.Emit(s, events.ReplyCreated, reply)

	// Add post/blob associations for new reply.
	//TODO: CHECK IF BLOB-SERVICE WILL ALLOW THIS WITHOUT ANY CHANGE
	if err = s.addBlobs(reply.Id, reply.Pictures); err != nil {
		return nil, err
	}
	return &reply, nil
}

func (s *Server) patchReply(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// authenticated users to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	// Get the post ID to update.
	replyId := chi.URLParam(r, "reply_id")
	if replyId == "" {
		return chassis.BadRequest(w, "missing reply ID")
	}

	// Read patch request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Look up post value and patch it.
	reply, err := s.db.ReplyById(replyId)
	if err != nil {
		if err == db.ErrReplyNotFound {
			return chassis.NotFoundWithMessage(w, err.Error())
		}
		return nil, err
	}

	// Determine permission information for update.
	if reply.Owner != authInfo.UserID {
		return chassis.NotFound(w)
	}

	picsBefore := map[string]bool{}
	for _, pic := range reply.Pictures {
		picsBefore[pic] = true
	}

	if err = reply.Patch(body); err != nil {
		return chassis.BadRequest(w, err.Error())
	}
	picsAfter := map[string]bool{}
	for _, pic := range reply.Pictures {
		picsAfter[pic] = true
	}
	reply.IsEdited = true
	// Do the update.
	if err = s.db.UpdateReply(reply); err != nil {
		return nil, err
	}

	chassis.Emit(s, events.ReplyUpdated, reply)

	// Update the item/blob associations.
	if err = s.updateBlobs(reply.Id, picsBefore, picsAfter); err != nil {
		return nil, err
	}

	return &reply, nil
}


func (s *Server) deleteReply(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// authenticated users to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	// Get the item ID to delete.
	replyId := chi.URLParam(r, "reply_id")
	if replyId == "" {
		return chassis.BadRequest(w, "missing reply ID")
	}


	rpl, err := s.db.ReplyById(replyId)
	if err != nil {
		return nil, err
	}

	if rpl.Owner != authInfo.UserID {
		return chassis.NotFound(w)
	}
	// Delete the item.
	rpl.IsDeleted = true
	if err := s.db.UpdateReply(rpl) ; err != nil {
		return nil, err
	}

	chassis.Emit(s, events.ReplyDeleted, rpl)

	if err = s.removeBlobs(replyId, rpl.Pictures); err != nil {
		return nil, err
	}
	return chassis.NoContent(w)
}
