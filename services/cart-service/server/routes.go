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

	r.Get("/carts", chassis.SimpleHandler(s.cartsSearch))
	r.Get("/cart/{cart_id}", chassis.SimpleHandler(s.cartSearch))
	r.Get("/cart/active", chassis.SimpleHandler(s.getActiveCart))

	r.Post("/cart", chassis.SimpleHandler(s.createCart))
	r.Patch("/cart/{cart_id}", chassis.SimpleHandler(s.patchCart))
	r.Put("/cart/forget", chassis.SimpleHandler(s.forgetCart))
	r.Patch("/cart/{cart_id}/merge", chassis.SimpleHandler(s.mergeCarts))
	//r.Delete("/cart/{cart_id}", chassis.SimpleHandler(s.deleteCart))

	r.Get("/cart/{cart_id}/items", chassis.SimpleHandler(s.cartItemsSearch))
	r.Get("/cart/{cart_id}/item/{citem_id}", chassis.SimpleHandler(s.cartItemSearch))
	r.Post("/cart/{cart_id}/item", chassis.SimpleHandler(s.addCartItem))
	r.Patch("/cart/{cart_id}/item/{citem_id}", chassis.SimpleHandler(s.patchCartItem))
	r.Delete("/cart/{cart_id}/item/{citem_id}", chassis.SimpleHandler(s.deleteCartItem))

	// INTERNAL-ONLY ROUTES (I.E. ACCESSED ONLY BY OTHER SERVICES,
	// EXPOSED VIA SERVICE CLIENT API).
	r.Get("/internal/cart/{user_id}/active", chassis.SimpleHandler(s.internalActive))
	r.Patch("/internal/cart/{cart_id}", chassis.SimpleHandler(s.internalPatchCart))


	return r
}
