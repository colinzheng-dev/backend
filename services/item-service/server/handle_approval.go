package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/item-service/db"
	"github.com/veganbase/backend/services/item-service/model/types"
)

type approvalChangeRequest struct {
	Approval types.ApprovalState `json:"approval"`
}

func (s *Server) changeItemApproval(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get authentication information from context and only allow
	// administrators to proceed.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth || !authInfo.UserIsAdmin {
		return chassis.NotFound(w)
	}

	// Look up item.
	id := chi.URLParam(r, "id")
	item, err := s.db.ItemByID(id)
	if err == db.ErrItemNotFound {
		return chassis.NotFound(w)
	}
	if err != nil {
		return nil, err
	}

	// Read and unmarshal request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}
	update := approvalChangeRequest{}
	err = json.Unmarshal(body, &update)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	err = s.db.UpdateItemApproval(item.ID, update.Approval)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	return chassis.NoContent(w)
}
