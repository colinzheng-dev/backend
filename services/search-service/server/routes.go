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

	// Search endpoints.
	r.Get("/search/geo", chassis.SimpleHandler(s.geo))
	r.Get("/search/full_text", chassis.SimpleHandler(s.fullText))
	r.Get("/search/region", chassis.SimpleHandler(s.region))
	r.Get("/search/check-region", chassis.SimpleHandler(s.checkRegion))
	r.Get("/search/countries", chassis.SimpleHandler(s.Countries))
	r.Get("/search/country/{country_id}/states", chassis.SimpleHandler(s.States))

	return r
}
