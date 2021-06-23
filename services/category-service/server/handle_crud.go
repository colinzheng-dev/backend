package server

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/category-service/db"
	"github.com/veganbase/backend/services/category-service/events"
	"github.com/veganbase/backend/services/category-service/model"
)

func (s *Server) categories(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	cats, err := s.db.Categories()
	if err != nil {
		return nil, err
	}

	res := map[string]*model.CategorySummary{}
	for n, cat := range cats {
		res[n] = model.Summary(cat)
	}

	return res, nil
}

func (s *Server) entries(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var fixed *bool
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.UserIsAdmin {
		qs, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			return chassis.BadRequest(w, "invalid query parameters")
		}
		fval := false
		switch qs.Get("fixed") {
		case "":
			fixed = nil
		case "false":
			fval = false
			fixed = &fval
		case "true":
			fval = true
			fixed = &fval
		default:
			return nil, errors.New("invalid 'fixed' query parameter")
		}
	}

	catName := chi.URLParam(r, "category")
	entries, err := s.db.CategoryEntries(catName, fixed)
	if err != nil {
		if err == db.ErrCategoryNotFound {
			return chassis.NotFound(w)
		}
		return nil, err
	}

	return entries, nil
}

func (s *Server) addOrEditEntry(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Route needs authentication.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	// Check request parameters.
	catName := chi.URLParam(r, "category")
	label := chi.URLParam(r, "label")
	if catName == "" {
		return chassis.BadRequest(w, "missing category name")
	}
	if label == "" {
		return chassis.BadRequest(w, "missing category entry label")
	}

	// Look up the category and check that it can be extended by the
	// user.
	cat, err := s.db.CategoryByName(catName)
	if err != nil {
		return chassis.NotFound(w)
	}
	if !cat.Extensible && !authInfo.UserIsAdmin {
		return chassis.Forbidden(w)
	}

	// Determine whether the label already exists in the category. If it
	// does, this is an edit, and can only proceed if: 1. the user is an
	// administrator, or 2. the label is not yet fixed and was created
	// by the user.
	newEntry := true
	info, err := s.db.EntryInfo(catName, label)
	if err != nil {
		return nil, err
	}
	if info != nil {
		newEntry = false
		if info.Fixed {
			if !authInfo.UserIsAdmin {
				return chassis.Forbidden(w)
			}
		} else {
			if info.Creator != authInfo.UserID {
				return chassis.Forbidden(w)
			}
		}
	}

	// Read request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Try to add or update entry in category. Failure at this point is
	// because of data that doesn't match the category schema, or a
	// failed uniqueness constraint.
	if newEntry {
		err = s.db.AddCategoryEntry(catName, label, body, authInfo.UserID)
	} else {
		err = s.db.UpdateCategoryEntry(catName, label, body)
	}
	if err == db.ErrCategoryNotFound {
		return chassis.NotFound(w)
	}
	if err == db.ErrSchemaMismatch || err == db.ErrCategoryLabelNotUnique {
		return chassis.BadRequest(w, err.Error())
	}
	if err != nil {
		return nil, err
	}

	update := events.CategoryUpdateInfo{
		Name: catName,
	}
	update.Entries, err = s.db.CategoryEntries(catName, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed getting information for category update event")
	}
	err = chassis.Emit(s, events.CategoryUpdate, update)
	if err != nil {
		log.Error().Err(err).Msg("failed emitting category update event")
	}
	return chassis.NoContent(w)
}

func (s *Server) fix(state bool) chassis.SimpleHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (interface{}, error) {
		// Routes are only usable by administrators.
		authInfo := chassis.AuthInfoFromContext(r.Context())
		if authInfo.AuthMethod != chassis.SessionAuth || !authInfo.UserIsAdmin {
			return chassis.NotFound(w)
		}

		// Check request parameters.
		catName := chi.URLParam(r, "category")
		label := chi.URLParam(r, "label")
		if catName == "" {
			return chassis.BadRequest(w, "missing category name")
		}
		if label == "" {
			return chassis.BadRequest(w, "missing category entry label")
		}

		err := s.db.FixCategoryEntry(catName, label, state)
		if err == db.ErrCategoryEntryNotFound {
			return chassis.NotFound(w)
		}
		if err != nil {
			return nil, err
		}
		return chassis.NoContent(w)
	}
}
