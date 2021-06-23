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

	// PATHS FOR PURCHASES
	r.Get("/purchases", chassis.SimpleHandler(s.purchasesSearch))
	r.Get("/purchase/{pur_id}", chassis.SimpleHandler(s.purchaseSearch))
	r.Post("/purchase", chassis.SimpleHandler(s.createPurchase))
	r.Post("/simple-purchase", chassis.SimpleHandler(s.createSimplePurchase))

	//PATHS FOR ORDERS OR BOOKINGS OF A CERTAIN PURCHASE
	r.Get("/purchase/{pur_id}/orders", chassis.SimpleHandler(s.ordersByPurchaseSearch))
	r.Get("/purchase/{pur_id}/bookings", chassis.SimpleHandler(s.bookingsByPurchaseSearch))

	//GENERAL PATHS FOR ORDERS AND BOOKINGS

	r.Get("/orders", chassis.SimpleHandler(s.ordersSearch))
	r.Get("/order/{ord_id}", chassis.SimpleHandler(s.orderSearch))
	r.Get("/bookings", chassis.SimpleHandler(s.bookingsSearch))
	r.Get("/booking/{bok_id}", chassis.SimpleHandler(s.bookingSearch))

	//ITEM SUBSCRIPTIONS
	// r.Get("/item-subscriptions", chassis.SimpleHandler(s.subscriptionItemSearch))
	// r.Get("/item-subscription/{sub_id}", chassis.SimpleHandler(s.subscriptionItemSearch))
	// r.Delete("/item-subscription/{sub_id}", chassis.SimpleHandler(s.deleteSubscriptionItem))
	// r.Patch("/item-subscription/{sub_id}", chassis.SimpleHandler(s.patchSubscriptionItem))
	// r.Patch("/item-subscription/{sub_id}/flip-state", chassis.SimpleHandler(s.flipStateSubscriptionItem))

	// INTERNAL-ONLY ROUTES (I.E. ACCESSED ONLY BY OTHER SERVICES,
	// EXPOSED VIA SERVICE CLIENT API).
	r.Patch("/internal/purchase/{pur_id}", chassis.SimpleHandler(s.patchPurchaseInternal))
	r.Patch("/internal/booking/{bok_id}", chassis.SimpleHandler(s.patchBookingInternal))
	r.Patch("/internal/order/{ord_id}", chassis.SimpleHandler(s.patchOrderInternal))
	r.Get("/internal/purchase/{pur_id}", chassis.SimpleHandler(s.purchaseSearchInternal))
	r.Get("/internal/purchase/item-bought", chassis.SimpleHandler(s.userPurchaseItem))
	return r
}
