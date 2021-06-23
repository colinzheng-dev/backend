package server

import (
	"github.com/go-chi/chi"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/webhook-service/db"
	"github.com/veganbase/backend/services/webhook-service/model"
	"net/http"
)


func (s *Server) getWebhook(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth {
		return chassis.NotFound(w)
	}
	hookID := chi.URLParam(r, "id")
	if hookID == "" {
		return chassis.BadRequest(w, "missing webhook ID")
	}
	addr, err := s.db.WebhookByID(hookID)
	if err != nil {
		if err == db.ErrWebhookNotFound {
			return chassis.NotFoundWithMessage(w, "webhook not found")
		}
		return nil, err
	}

	return addr, nil
}

func (s *Server) getWebhooks(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth {
		return chassis.NotFound(w)
	}

	hooks, err := s.db.WebhooksByOwner(authInfo.UserID)
	if err != nil {
		return nil, err
	}

	return hooks, nil
}

func (s *Server) createWebhook(w http.ResponseWriter, r *http.Request) (interface{}, error) {

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
	hook := model.Webhook{}
	if err = hook.UnmarshalJSON(body); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Create the item.
	if err = s.db.CreateWebhook(&hook);err != nil {
		return nil, err
	}
	//s.emit(events.ItemCreated, item.ID)

	return hook, nil
}



func (s *Server) patchWebhook(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	// Get the item ID to update.
	hookID := chi.URLParam(r, "id")
	if hookID == "" {
		return chassis.BadRequest(w, "missing webhook ID")
	}

	// Read patch request body.
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Look up item value and patch it.
	hook, err := s.db.WebhookByID(hookID)
	if err != nil {
		if err == db.ErrWebhookNotFound {
			return chassis.NotFoundWithMessage(w, "webhook not found")
		}
		return nil, err
	}
	if hook.Owner != authInfo.UserID {
		return chassis.Forbidden(w)
	}

	if 	err = hook.Patch(body); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Do the update.
	if err = s.db.UpdateWebhook(hook); err != nil {
		if err == db.ErrWebhookNotFound {
			return chassis.NotFound(w)
		}
		return nil, err
	}

	//s.emit(events.ItemUpdated, item.ID)

	return hook, nil
}


func (s *Server) deleteWebhook(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	// Get the item ID to delete.
	hookID := chi.URLParam(r, "id")
	if hookID == "" {
		return chassis.BadRequest(w, "missing webhook ID")
	}

	hook, err := s.db.WebhookByID(hookID)
	if err != nil {
		if err == db.ErrWebhookNotFound {
			return chassis.NotFoundWithMessage(w, "webhook not found")
		}
		return nil, err
	}
	//TODO: WEBHOOKS MUST BE ENABLED TO ORGS AS WELL
	if hook.Owner != authInfo.UserID {
		return chassis.Forbidden(w)
	}

	if err = s.db.DeleteWebhook(hookID); err != nil {
		return nil, err
	}

	//s.emit(events.WebhookDeleted, hook)

	return chassis.NoContent(w)
}
