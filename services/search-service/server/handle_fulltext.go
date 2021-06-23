package server

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/item-service/model"
	"github.com/veganbase/backend/services/item-service/model/types"
	"github.com/veganbase/backend/services/item-service/utils"
)

func (s *Server) fullText(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return nil, errors.New("invalid query parameters")
	}

	var q *string
	chassis.StringParam(qs, "q", &q)
	if q == nil {
		return chassis.BadRequest(w, "empty full-text query string")
	}

	var itemType *model.ItemType
	if err := utils.ItemTypeParam(qs, &itemType); err != nil {
		return chassis.BadRequest(w, "invalid type parameter")
	}

	var approval *[]types.ApprovalState
	if err := utils.ApprovalParam(qs, &approval); err != nil {
		return chassis.BadRequest(w, "invalid approval parameter")
	}

	res, err := s.db.FullText(*q, itemType, approval)
	if err != nil {
		return nil, err
	}
	return res, nil
}
