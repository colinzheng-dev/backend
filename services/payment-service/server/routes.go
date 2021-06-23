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

	r.Post("/webhook/stripe", chassis.SimpleHandler(s.webhookHandler))
	r.Post("/payment-intent", chassis.SimpleHandler(s.createPaymentIntentWithAuth))
	r.Post("/payment-intent/{pi_id:pi_[a-zA-Z0-9]+}/confirm", chassis.SimpleHandler(s.confirmPaymentIntent))

	// PATHS FOR INTERNAL USE

	r.Post("/internal/payment-intent", chassis.SimpleHandler(s.createPaymentIntent))

	return r
}
