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

	// Routes for authenticated user.
	r.Route("/me", func(r chi.Router) {
		r.Get("/", chassis.SimpleHandler(s.detail))
		r.Patch("/", chassis.SimpleHandler(s.update))
		r.Delete("/", chassis.SimpleHandler(s.delete))
		r.Post("/api-key", chassis.SimpleHandler(s.createAPIKey))
		r.Delete("/api-key", chassis.SimpleHandler(s.deleteAPIKey))
		r.Get("/orgs", chassis.SimpleHandler(s.userOrgs))

		r.Get("/payout-account", chassis.SimpleHandler(s.getUserPayoutAccount))
		r.Post("/payout-account", chassis.SimpleHandler(s.createPayoutAccount))
		r.Delete("/payout-account", chassis.SimpleHandler(s.deleteUserPayoutAccount))
		r.Patch("/payout-account", chassis.SimpleHandler(s.updateUserPayoutAccount))

		r.Get("/payment-methods", chassis.SimpleHandler(s.getPaymentMethods))
		r.Get("/payment-method/default", chassis.SimpleHandler(s.getDefaultPaymentMethod))
		r.Get("/payment-method/{pmt_id:pmt_[a-zA-Z0-9]+}", chassis.SimpleHandler(s.getPaymentMethod))
		r.Post("/payment-method", chassis.SimpleHandler(s.createPaymentMethod))
		r.Delete("/payment-method/{pmt_id:pmt_[a-zA-Z0-9]+}", chassis.SimpleHandler(s.deletePaymentMethod))
		r.Patch("/payment-method/{pmt_id:pmt_[a-zA-Z0-9]+}", chassis.SimpleHandler(s.updatePaymentMethod))

		r.Get("/customer", chassis.SimpleHandler(s.getCustomer))
		r.Post("/customer", chassis.SimpleHandler(s.createCustomer))
		r.Delete("/customer", chassis.SimpleHandler(s.deleteCustomer))

		r.Get("/addresses", chassis.SimpleHandler(s.getAddresses))
		r.Get("/address/default", chassis.SimpleHandler(s.getDefaultAddress))
		r.Get("/address/{addr_id:adr_[a-zA-Z0-9]+}", chassis.SimpleHandler(s.getAddress))
		r.Post("/address", chassis.SimpleHandler(s.createAddress))
		r.Delete("/address/{addr_id:adr_[a-zA-Z0-9]+}", chassis.SimpleHandler(s.deleteAddress))
		r.Patch("/address/{addr_id:adr_[a-zA-Z0-9]+}", chassis.SimpleHandler(s.updateAddress))

		r.Get("/delivery-fees", chassis.SimpleHandler(s.getUserDeliveryFees))
		r.Post("/delivery-fees", chassis.SimpleHandler(s.createDeliveryFees))
		r.Delete("/delivery-fees", chassis.SimpleHandler(s.deleteUserDeliveryFees))
		r.Patch("/delivery-fees", chassis.SimpleHandler(s.updateUserDeliveryFees))
	})

	// Routes for other user (administrator only apart from GET
	// /user/{id} which returns a user's public profile).
	r.Route("/user/{id:usr_[a-zA-Z0-9]+}", func(r chi.Router) {
		r.Get("/", chassis.SimpleHandler(s.detail))
		r.Patch("/", chassis.SimpleHandler(s.update))
		r.Delete("/", chassis.SimpleHandler(s.delete))
		r.Post("/api-key", chassis.SimpleHandler(s.createAPIKey))
		r.Delete("/api-key", chassis.SimpleHandler(s.deleteAPIKey))
		r.Get("/orgs", chassis.SimpleHandler(s.userOrgs))
	})

	// Organisations.
	r.Route("/orgs", func(r chi.Router) {
		r.Post("/", chassis.SimpleHandler(s.createOrg))
		r.Get("/", chassis.SimpleHandler(s.orgs))
	})
	r.Route("/org/{id_or_slug:[a-zA-Z0-9-_]+}", func(r chi.Router) {
		r.Get("/", chassis.SimpleHandler(s.orgDetail))
		r.Patch("/", chassis.SimpleHandler(s.updateOrg))
		r.Delete("/", chassis.SimpleHandler(s.deleteOrg))
		r.Get("/users", chassis.SimpleHandler(s.orgUsers))
		r.Post("/users", chassis.SimpleHandler(s.orgAddUser))
		r.Patch("/user/{user_id:usr_[a-zA-Z0-9]+}", chassis.SimpleHandler(s.orgPatchUser))
		r.Delete("/user/{user_id:usr_[a-zA-Z0-9]+}", chassis.SimpleHandler(s.orgDeleteUser))

		r.Get("/payout-account", chassis.SimpleHandler(s.getOrgPayoutAccount))
		r.Post("/payout-account", chassis.SimpleHandler(s.createPayoutAccount))
		r.Delete("/payout-account", chassis.SimpleHandler(s.deleteOrgPayoutAccount))
		r.Patch("/payout-account", chassis.SimpleHandler(s.updateOrgPayoutAccount))

		r.Get("/delivery-fees", chassis.SimpleHandler(s.getOrgDeliveryFees))
		r.Post("/delivery-fees", chassis.SimpleHandler(s.createDeliveryFees))
		r.Delete("/delivery-fees", chassis.SimpleHandler(s.deleteOrgDeliveryFees))
		r.Patch("/delivery-fees", chassis.SimpleHandler(s.updateOrgDeliveryFees))

		// SSO work has been disabled as the Kube manifest on prod doesn't contain an ENCRYPTION_KEY value
		// r.Get("/sso-secret", chassis.SimpleHandler(s.getSSOSecret))
		// r.Post("/sso-secret", chassis.SimpleHandler(s.createSSOSecret))
		// r.Delete("/sso-secret", chassis.SimpleHandler(s.dropSSOSecret))
	})

	// Admin-only user list.
	r.Get("/users", chassis.SimpleHandler(s.list))

	// INTERNAL-ONLY ROUTES (I.E. ACCESSED ONLY BY OTHER SERVICES,
	// EXPOSED VIA SERVICE CLIENT API).

	// Successful login.
	r.Post("/login", chassis.SimpleHandler(s.login))

	// Minimal user information for a list of user IDs (name and email
	// for each).
	r.Get("/info", chassis.SimpleHandler(s.info))

	//routes used by purchase and payment services
	r.Get("/internal/notification-info/{id:(usr|org)_[a-zA-Z0-9]+}", chassis.SimpleHandler(s.getNotificationInfoInternal))
	r.Get("/internal/user/{user_id:usr_[a-zA-Z0-9]+}/customer", chassis.SimpleHandler(s.getCustomerInternal))
	r.Get("/internal/user/{user_id:usr_[a-zA-Z0-9]+}/address/{addr_id:adr_[a-zA-Z0-9]+}", chassis.SimpleHandler(s.getAddressInternal))
	r.Get("/internal/user/{user_id:usr_[a-zA-Z0-9]+}/address/default", chassis.SimpleHandler(s.getAddressInternal))
	r.Get("/internal/user/{user_id:usr_[a-zA-Z0-9]+}/payment-method/default", chassis.SimpleHandler(s.getDefaultPaymentMethodInternal))
	r.Get("/internal/payout-account/{id:(usr|org)_[a-zA-Z0-9]+}", chassis.SimpleHandler(s.getPayoutInternal))
	r.Get("/internal/api-key/{key}", chassis.SimpleHandler(s.getUserByApiKeyInternal))
	r.Get("/internal/delivery-fees", chassis.SimpleHandler(s.getDeliveryFeesInternal))
	// r.Get("/internal/org/{id_or_slug:[a-zA-Z0-9-_]+}/sso-secret", chassis.SimpleHandler(s.getSSOSecretInternal))

	return r
}
