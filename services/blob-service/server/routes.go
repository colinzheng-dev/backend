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

	// Create blob.
	r.Post("/blobs", chassis.SimpleHandler(s.create))

	// Blob tags in use for user.
	r.Get("/blobs/tags", chassis.SimpleHandler(s.tagList))

	// Single blob (c)RUD.
	r.Route("/blob/{id:[a-zA-Z0-9]+}", func(r chi.Router) {
		r.Get("/", chassis.SimpleHandler(s.detail))
		r.Patch("/", chassis.SimpleHandler(s.update))
		r.Delete("/", chassis.SimpleHandler(s.delete))
	})

	// Blob list for user.
	r.Get("/me/blobs", chassis.SimpleHandler(s.list))
	r.Get("/user/{userid:usr_[a-zA-Z0-9]+}/blobs", chassis.SimpleHandler(s.list))

	// UTILITY ROUTES ACCESSED ONLY BY OTHER SERVICES.

	// Blob-item associations.
	r.Route("/blob-item-assoc", func(r chi.Router) {
		r.Post("/{item:[a-z]+_[a-zA-Z0-9]+}", chassis.SimpleHandler(s.addItemBlobs))
		r.Delete("/{item:[a-z]+_[a-zA-Z0-9]+}", chassis.SimpleHandler(s.removeItemBlobs))
	})

	return r
}
