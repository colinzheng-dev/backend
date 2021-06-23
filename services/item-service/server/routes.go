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

	r.Post("/items", chassis.SimpleHandler(s.createItem))
	r.Get("/items", chassis.SimpleHandler(s.itemSearch))
	r.Get("/items/info", chassis.SimpleHandler(s.getItemTypeSummaryInfo))

	r.Get("/item/{id_or_slug}", chassis.SimpleHandler(s.getByIDOrSlug))
	r.Patch("/item/{id}", chassis.SimpleHandler(s.patchItem))
	r.Delete("/item/{id}", chassis.SimpleHandler(s.deleteItem))
	r.Post("/item/{id}/approval", chassis.SimpleHandler(s.changeItemApproval))
	r.Post("/item/{id}/claim-ownership", chassis.SimpleHandler(s.claimItemOwnership))
	r.Post("/item/{item_id}/links", chassis.SimpleHandler(s.createLink))
	r.Get("/item/{item_id}/links", chassis.SimpleHandler(s.getItemLinks))

	r.Get("/me/tags", chassis.SimpleHandler(s.tagsForUser))
	r.Get("/me/items", chassis.SimpleHandler(s.itemsForUser))
	r.Get("/me/item-collections", chassis.SimpleHandler(s.collsForUser))

	r.Get("/user/{id}/tags", chassis.SimpleHandler(s.tagsForUser))
	r.Get("/user/{id}/items", chassis.SimpleHandler(s.itemsForUser))
	r.Get("/user/{user_id}/item-collections", chassis.SimpleHandler(s.collsForUser))

	r.Get("/org/{id_or_slug}/item-collections", chassis.SimpleHandler(s.collsForOrg))
	r.Get("/org/{id_or_slug}/items", chassis.SimpleHandler(s.itemsForOrg))

	r.Get("/tag/{tag}", chassis.SimpleHandler(s.tagSearch))

	r.Get("/ownership-claims", chassis.SimpleHandler(s.listOwnershipClaims))
	r.Delete("/ownership-claim/{id}", chassis.SimpleHandler(s.deleteOwnershipClaim))
	r.Post("/ownership-claim/{id}/status", chassis.SimpleHandler(s.changeOwnershipClaimStatus))

	r.Delete("/item-link/{link_id}", chassis.SimpleHandler(s.deleteLink))
	r.Get("/item-link/{link_id}", chassis.SimpleHandler(s.getItemLink))

	r.Post("/item-collections", chassis.SimpleHandler(s.createColl))
	r.Get("/item-collections", chassis.SimpleHandler(s.listColls))
	r.Get("/item-collections/list", chassis.SimpleHandler(s.collListAll))

	r.Get("/item-collection/{coll_id}", chassis.SimpleHandler(s.collDetail))
	r.Delete("/item-collection/{coll_id}", chassis.SimpleHandler(s.deleteColl))
	r.Put("/item-collection/{coll_id}/item/{item_id}", chassis.SimpleHandler(s.collAddItem))
	r.Delete("/item-collection/{coll_id}/item/{item_id}", chassis.SimpleHandler(s.collRemoveItem))
	r.Get("/item-collection/{coll_id}/items", chassis.SimpleHandler(s.collPreview))

	// INTERNAL-ONLY ROUTES (I.E. ACCESSED ONLY BY OTHER SERVICES,
	// EXPOSED VIA SERVICE CLIENT API).

	r.Get("/ids", chassis.SimpleHandler(s.ids))
	r.Get("/search_info/{id}", chassis.SimpleHandler(s.searchInfo))
	r.Patch("/internal/item/{id}", chassis.SimpleHandler(s.updateAvailability))
	r.Get("/internal/info", chassis.SimpleHandler(s.getItemsBasicInfo))


	return r
}
