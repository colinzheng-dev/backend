package server

import (
	"github.com/go-chi/chi"

	"github.com/veganbase/backend/chassis"
)

func (s *Server) routes() chi.Router {
	r := chi.NewRouter()

	// Add common middleware.
	chassis.AddCommonMiddleware(r, true)

	// Inject authentication information into request context.
	r.Use(chassis.AuthCtx)

	// Service health checks.
	r.Get("/", chassis.Health)
	r.Get("/healthz", chassis.Health)

	// Webhooks endpoints.
	r.Get("/me/subscriptions", chassis.SimpleHandler(s.getWebhooks))
	r.Get("/me/subscription/{id}", chassis.SimpleHandler(s.getWebhook))
	r.Post("/me/subscription", chassis.SimpleHandler(s.createWebhook))
	r.Patch("/me/subscription/{id}", chassis.SimpleHandler(s.patchWebhook))
	r.Delete("/me/subscription/{id}", chassis.SimpleHandler(s.deleteWebhook))

	r.Post("/webhooks/send-test-event", chassis.SimpleHandler(s.sendTestEvent))

	return r
}
