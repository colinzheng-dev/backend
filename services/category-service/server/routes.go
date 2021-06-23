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

	// PASS-THROUGH ROUTES (I.E. FORWARDED DIRECTLY FROM API GATEWAY).

	// Service health checks.
	r.Get("/", chassis.Health)
	r.Get("/healthz", chassis.Health)

	r.Get("/categories", chassis.SimpleHandler(s.categories))
	r.Route("/category/{category}", func(r chi.Router) {
		r.Get("/", chassis.SimpleHandler(s.entries))
		r.Put("/{label}", chassis.SimpleHandler(s.addOrEditEntry))
		r.Put("/{label}/fix", chassis.SimpleHandler(s.fix(true)))
		r.Delete("/{label}/fix", chassis.SimpleHandler(s.fix(false)))
	})

	return r
}
