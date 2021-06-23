package server

import (
	"github.com/go-chi/chi"

	"github.com/veganbase/backend/chassis"
)

func (s *Server) routes() chi.Router {
	r := chi.NewRouter()

	// Add common middleware.
	chassis.AddCommonMiddleware(r, true)

	// Service health checks.
	r.Get("/", chassis.Health)
	r.Get("/healthz", chassis.Health)

	// Get site list.
	r.Get("/sites", chassis.SimpleHandler(s.list))

	return r
}
