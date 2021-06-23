package server

import (
	"github.com/go-chi/chi"

	"github.com/veganbase/backend/chassis"
)

//returning two routers, one for the API with CSRF protection and the other for stripe's webhook
func (s *Server) routes(devMode bool, csrfSecret string) chi.Router {
	r := chi.NewRouter()

	//all routes must validate the origin
	s.addCORSMiddleware(r)

	//unprotected path to be called by Stripe
	r.Method("POST", "/webhook/stripe", Forward(s.paymentSvcURL))

	r.Group(func (r chi.Router){
		//all veganbase paths must perform CSRF validation
		if !devMode && !s.disabledCSRF {
			s.addCSRFMiddleware(r, devMode, csrfSecret)
		}
		r.Get("/", chassis.Health)
		r.Get("/healthz", chassis.Health)

		// These routes need to be outside of the following block to exempt
		// them from CSRF protection since they're used before a session is
		// established.
		r.Post("/auth/request-login-email", chassis.SimpleHandler(s.requestLoginEmail))
		r.Post("/auth/login", chassis.SimpleHandler(s.login))

		r.Group(func(r chi.Router) {
			r.Use(CredentialCtx(s))

			// Authentication.
			r.Post("/auth/logout", chassis.SimpleHandler(s.logout))
			r.Post("/auth/logout-all", chassis.SimpleHandler(s.logoutAll))

			// Routes for authenticated user.
			r.Route("/me", s.userRoutes)

			// Routes for customer (separate from userRoutes because /user/id/customer is internal)
			r.Route("/me/customer", s.customerRoutes)
			r.Route("/me/delivery-fees", s.deliveryFeesRoutes)

			//Routes for other user (administrator only apart from GET
			// /user/{id} which returns a user's public profile).
			r.Route("/user/{id:usr_[a-zA-Z0-9]+}", s.userRoutes)

			// Admin-only user list.
			r.Method("GET", "/users", Forward(s.userSvcURL))

			// Organisations.
			r.Method("GET", "/orgs", Forward(s.userSvcURL))
			r.Method("POST", "/orgs", Forward(s.userSvcURL))
			r.Route("/org/{id_or_slug:[a-zA-Z0-9-_]+}", s.orgRoutes)

			// Blobs.
			r.Route("/blobs", s.blobsRoutes)
			r.Route("/blob/{id}", s.blobRoutes)

			// Items.
			r.Route("/items", s.itemsRoutes)
			r.Route("/item/{id:[a-z]+_[a-zA-Z0-9]+}", s.itemRoutes)
			r.Method("GET", "/item/{slug:[a-z0-9-]+}", Forward(s.itemSvcURL))
			r.Method("GET", "/tag/{tag:[a-z-]+}", Forward(s.itemSvcURL))
			r.Route("/item-link/{id:lnk_[a-zA-Z0-9]+}", s.linkRoutes)

			// Item ownership claims.
			r.Method("GET", "/ownership-claims", Forward(s.itemSvcURL))
			r.Route("/ownership-claim/claim_{id:[a-zA-Z0-9]+}", s.claimRoutes)

			// Item collections.
			r.Method("POST", "/item-collections", Forward(s.itemSvcURL))
			r.Method("GET", "/item-collections", Forward(s.itemSvcURL))
			r.Method("GET", "/item-collections/list", Forward(s.itemSvcURL))
			r.Route(`/item-collection/{coll:[a-zA-Z0-9-:_\s&,]+}`, s.itemCollectionRoutes)

			// Categories.
			r.Method("GET", "/categories", Forward(s.categorySvcURL))
			r.Route("/category", s.categoryRoutes)

			// Carts
			r.Method("GET", "/carts", Forward(s.cartSvcURL))
			r.Method("POST", "/cart", Forward(s.cartSvcURL))
			r.Method("GET", "/cart/active", Forward(s.cartSvcURL))
			r.Method("PUT", "/cart/forget", Forward(s.cartSvcURL))
			r.Route(`/cart/{cart_id:car_[a-zA-Z0-9]+}`, s.cartRoutes)

			// Purchase
			r.Method("GET", "/purchases", Forward(s.purchaseSvcURL))
			r.Method("GET", "/orders", Forward(s.purchaseSvcURL))
			r.Method("GET", "/bookings", Forward(s.purchaseSvcURL))
			r.Method("POST", "/simple-purchase", Forward(s.purchaseSvcURL))
			r.Method("GET", "/item-subscriptions", Forward(s.purchaseSvcURL))
			r.Route(`/purchase`, s.purchaseRoutes)
			r.Route(`/booking`, s.purchaseRoutes)
			r.Route(`/order`, s.purchaseRoutes)
			r.Route(`/item-subscription`, s.purchaseRoutes)


			// Social
			r.Route("/social", s.socialRoutes)
			r.Route("/post", s.socialRoutes)
			r.Route("/reply", s.socialRoutes)

			// Payment
			r.Method("POST", "/payment-intent", Forward(s.paymentSvcURL))
			r.Method("POST", "/payment-intent/{pi_id:pi_[a-zA-Z0-9]+}/confirm", Forward(s.paymentSvcURL))

			// Search
			r.Route( "/search", s.searchRoutes)

			// Webhooks
			r.Method("POST", "/webhooks/send-test-event", Forward(s.webhookSvcURL))
		})
	})

	return r
}

func (s *Server) userRoutes(r chi.Router) {
	r.Method("GET", "/", Forward(s.userSvcURL))
	r.Method("PATCH", "/", Forward(s.userSvcURL))
	r.Method("DELETE", "/", Forward(s.userSvcURL))
	r.Method("POST", "/api-key", Forward(s.userSvcURL))
	r.Method("DELETE", "/api-key", Forward(s.userSvcURL))
	r.Method("GET", "/blobs", Forward(s.blobSvcURL))
	r.Method("GET", "/items", Forward(s.itemSvcURL))
	r.Method("GET", "/tags", Forward(s.userSvcURL))
	r.Method("GET", "/orgs", Forward(s.userSvcURL))
	r.Method("GET", "/item-collections", Forward(s.itemSvcURL))
	r.Method("GET", "/item-collections/list", Forward(s.itemSvcURL))
	r.Method("GET", "/payout-account", Forward(s.userSvcURL))
	r.Method("POST","/payout-account", Forward(s.userSvcURL))
	r.Method("DELETE","/payout-account", Forward(s.userSvcURL))
	r.Method("PATCH","/payout-account", Forward(s.userSvcURL))
	r.Method("GET", "/payment-methods", Forward(s.userSvcURL))
	r.Method("GET", "/payment-method/{pmt_id:pmt_[a-zA-Z0-9]+}", Forward(s.userSvcURL))
	r.Method("GET","/payment-method/default", Forward(s.userSvcURL))
	r.Method("POST","/payment-method", Forward(s.userSvcURL))
	r.Method("DELETE","/payment-method/{pmt_id:pmt_[a-zA-Z0-9]+}", Forward(s.userSvcURL))
	r.Method("PATCH","/payment-method/{pmt_id:pmt_[a-zA-Z0-9]+}", Forward(s.userSvcURL))
	r.Method("GET", "/addresses", Forward(s.userSvcURL))
	r.Method("GET", "/address/{addr_id:adr_[a-zA-Z0-9]+}", Forward(s.userSvcURL))
	r.Method("GET","/address/default", Forward(s.userSvcURL))
	r.Method("POST","/address", Forward(s.userSvcURL))
	r.Method("DELETE","/address/{adr_id:adr_[a-zA-Z0-9]+}", Forward(s.userSvcURL))
	r.Method("PATCH","/address/{adr_id:adr_[a-zA-Z0-9]+}", Forward(s.userSvcURL))
	r.Method("GET","/webhook-subscriptions", Forward(s.webhookSvcURL))
	r.Method("GET","/webhook-subscription/{id:whk_[a-zA-Z0-9]+}", Forward(s.webhookSvcURL))
	r.Method("POST","/webhook-subscription", Forward(s.webhookSvcURL))
	r.Method("PATCH","/webhook-subscription/{id:whk_[a-zA-Z0-9]+}", Forward(s.webhookSvcURL))
	r.Method("DELETE","/webhook-subscription/{id:whk_[a-zA-Z0-9]+}", Forward(s.webhookSvcURL))
}

func (s *Server) orgRoutes(r chi.Router) {
	r.Method("GET", "/", Forward(s.userSvcURL))
	r.Method("PATCH", "/", Forward(s.userSvcURL))
	r.Method("DELETE", "/", Forward(s.userSvcURL))
	r.Method("GET", "/users", Forward(s.userSvcURL))
	r.Method("POST", "/users", Forward(s.userSvcURL))
	r.Method("GET", "/items", Forward(s.itemSvcURL))
	r.Method("GET", "/item-collections", Forward(s.itemSvcURL))
	r.Method("GET", "/item-collections/list", Forward(s.itemSvcURL))
	r.Method("PATCH", "/user/{user_id:usr_[a-zA-Z0-9]+}", Forward(s.userSvcURL))
	r.Method("DELETE", "/user/{user_id:usr_[a-zA-Z0-9]+}", Forward(s.userSvcURL))
	r.Method("GET", "/payout-account", Forward(s.userSvcURL))
	r.Method("POST","/payout-account",  Forward(s.userSvcURL))
	r.Method("DELETE","/payout-account",  Forward(s.userSvcURL))
	r.Method("PATCH","/payout-account",  Forward(s.userSvcURL))
	r.Method("GET","/delivery-fees", Forward(s.userSvcURL))
	r.Method("POST","/delivery-fees", Forward(s.userSvcURL))
	r.Method("DELETE","/delivery-fees", Forward(s.userSvcURL))
	r.Method("PATCH","/delivery-fees", Forward(s.userSvcURL))
}

func (s *Server) blobsRoutes(r chi.Router) {
	r.Method("POST", "/", Forward(s.blobSvcURL))
	r.Method("GET", "/tags", Forward(s.blobSvcURL))
}

func (s *Server) blobRoutes(r chi.Router) {
	r.Method("GET", "/", Forward(s.blobSvcURL))
	r.Method("PATCH", "/", Forward(s.blobSvcURL))
	r.Method("DELETE", "/", Forward(s.blobSvcURL))
}

func (s *Server) itemsRoutes(r chi.Router) {
	r.Method("GET", "/", Forward(s.itemSvcURL))
	r.Method("POST", "/", Forward(s.itemSvcURL))
	r.Method("GET", "/info", Forward(s.itemSvcURL))
}

func (s *Server) itemRoutes(r chi.Router) {
	r.Method("GET", "/", Forward(s.itemSvcURL))
	r.Method("PATCH", "/", Forward(s.itemSvcURL))
	r.Method("DELETE", "/", Forward(s.itemSvcURL))
	r.Method("POST", "/approval", Forward(s.itemSvcURL))
	r.Method("POST", "/claim-ownership", Forward(s.itemSvcURL))
	r.Method("POST", "/links", Forward(s.itemSvcURL))
	r.Method("GET", "/links", Forward(s.itemSvcURL))
	r.Method("POST", "/upvote", Forward(s.socialSvcURL))
	r.Method("DELETE", "/upvote", Forward(s.socialSvcURL))
	r.Method("GET", "/posts", Forward(s.socialSvcURL))
	r.Method("POST", "/post", Forward(s.socialSvcURL))
}

func (s *Server) linkRoutes(r chi.Router) {
	r.Method("GET", "/", Forward(s.itemSvcURL))
	r.Method("DELETE", "/", Forward(s.itemSvcURL))
}

func (s *Server) claimRoutes(r chi.Router) {
	r.Method("DELETE", "/", Forward(s.itemSvcURL))
	r.Method("POST", "/status", Forward(s.itemSvcURL))
}

func (s *Server) itemCollectionRoutes(r chi.Router) {
	r.Method("GET", "/", Forward(s.itemSvcURL))
	r.Method("DELETE", "/", Forward(s.itemSvcURL))
	r.Method("PUT", "/item/{item_id}", Forward(s.itemSvcURL))
	r.Method("DELETE", "/item/{item_id}", Forward(s.itemSvcURL))
	r.Method("GET", "/items", Forward(s.itemSvcURL))
}

func (s *Server) categoryRoutes(r chi.Router) {
	r.Method("GET", "/{category:[a-z-]+}", Forward(s.categorySvcURL))
	r.Method("PUT", "/{category:[a-z-]+}/{label:[a-z-]+}", Forward(s.categorySvcURL))
	r.Method("PUT", "/{category:[a-z-]+}/{label:[a-z-]+}/fix", Forward(s.categorySvcURL))
	r.Method("DELETE", "/{category:[a-z-]+}/{label:[a-z-]+}/fix", Forward(s.categorySvcURL))
}

func (s *Server) socialRoutes(r chi.Router) {
	r.Method("GET", "/subscriptions/", Forward(s.socialSvcURL))
	r.Method("GET", "/followers/{subscription_id}", Forward(s.socialSvcURL))
	r.Method("POST", "/follow/{subscription_id}", Forward(s.socialSvcURL))
	r.Method("DELETE", "/forget/{subscription_id}", Forward(s.socialSvcURL))
	r.Method("POST", "/", Forward(s.socialSvcURL))
	r.Method("DELETE", "/", Forward(s.socialSvcURL))
	//r.Method("POST", "/", Forward(s.socialSvcURL))
	r.Method("PATCH", "/{reply_id}", Forward(s.socialSvcURL))
	r.Method("DELETE", "/{reply_id}", Forward(s.socialSvcURL))
}

func (s *Server) purchaseRoutes(r chi.Router) {

	r.Method("POST", "/", Forward(s.purchaseSvcURL))
	r.Method("GET", "/{pur_id:[A-Z]{3}[-][0-9]{6}}", Forward(s.purchaseSvcURL))
	r.Method("GET", "/{pur_id:[A-Z]{3}[-][0-9]{6}}/bookings", Forward(s.purchaseSvcURL))
	r.Method("GET", "/{pur_id:[A-Z]{3}[-][0-9]{6}}/orders", Forward(s.purchaseSvcURL))

	r.Method("GET", "/{bok_id:[A-Z]{3}[-][0-9]{6}}", Forward(s.purchaseSvcURL))
	r.Method("GET", "/{ord_id:[A-Z]{3}[-][0-9]{6}}", Forward(s.purchaseSvcURL))


	r.Method("GET","/{sub_id:sub_[a-zA-Z0-9]+}", Forward(s.purchaseSvcURL))
	r.Method("PATCH","/sub_id:sub_[a-zA-Z0-9]+}", Forward(s.purchaseSvcURL))
	r.Method("PATCH","/{sub_id:sub_[a-zA-Z0-9]+}/flip-state", Forward(s.purchaseSvcURL))
	r.Method("DELETE","/{sub_id:sub_[a-zA-Z0-9]+}", Forward(s.purchaseSvcURL))

}

func (s *Server) cartRoutes(r chi.Router) {
	r.Method("GET", "/", Forward(s.cartSvcURL))
	r.Method("PATCH", "/", Forward(s.cartSvcURL))
	r.Method("PATCH", "/merge", Forward(s.cartSvcURL))
	r.Method("GET", "/items", Forward(s.cartSvcURL))
	r.Method("GET", "/item/{citem_id:[0-9]+}", Forward(s.cartSvcURL))
	r.Method("PATCH", "/item/{citem_id:[0-9]+}", Forward(s.cartSvcURL))
	r.Method("DELETE", "/item/{citem_id:[0-9]+}", Forward(s.cartSvcURL))
	r.Method("POST", "/item", Forward(s.cartSvcURL))
}

func (s *Server) customerRoutes(r chi.Router) {
	r.Method("GET", "/", Forward(s.userSvcURL))
	r.Method("POST", "/", Forward(s.userSvcURL))
	r.Method("DELETE", "/", Forward(s.userSvcURL))
}

func (s *Server) deliveryFeesRoutes(r chi.Router) {
	r.Method("GET", "/", Forward(s.userSvcURL))
	r.Method("POST", "/", Forward(s.userSvcURL))
	r.Method("DELETE", "/", Forward(s.userSvcURL))
	r.Method("PATCH", "/", Forward(s.userSvcURL))
}

func (s *Server) searchRoutes(r chi.Router) {
	r.Method("GET", "/countries", Forward(s.searchSvcURL))
	r.Method("GET", "/country/{country_id}/states", Forward(s.searchSvcURL))
}
