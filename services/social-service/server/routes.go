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

	// Service health checks.
	r.Get("/", chassis.Health)
	r.Get("/healthz", chassis.Health)

	r.Route("/social", func(r chi.Router) {
		r.Get("/subscriptions/", chassis.SimpleHandler(s.listSubscriptions))
		r.Get("/followers/{subscription_id:(?:usr|org)_[a-zA-Z0-9]+}", chassis.SimpleHandler(s.listFollowers))
		r.Post("/follow/{subscription_id:(?:usr|org)_[a-zA-Z0-9]+}", chassis.SimpleHandler(s.addSubscription))
		r.Delete("/forget/{subscription_id:(?:usr|org)_[a-zA-Z0-9]+}", chassis.SimpleHandler(s.deleteSubscription))
	})

	//UPVOTE FEATURE
	r.Get("/me/upvotes", chassis.SimpleHandler(s.getUserUpvotesProtected))
	r.Route("/item", func(r chi.Router) {
		r.Post("/{item_id}/upvote", chassis.SimpleHandler(s.createUpvote))
		r.Delete("/{item_id}/upvote", chassis.SimpleHandler(s.deleteUpvote))
		r.Post("/{item_id}/post",chassis.SimpleHandler(s.createItemPost))
		r.Get("/{item_id}/posts",chassis.SimpleHandler(s.doPostSearchForItems))
		//r.Get("/{item_id}/post/{post_id}",chassis.SimpleHandler(s.doSubjectSearch))
	})

	//r.Get("/posts", chassis.SimpleHandler(s.doSubjectSearch))

	r.Route("/post", func(r chi.Router) {
		r.Delete("/{post_id}", chassis.SimpleHandler(s.deletePost))
		r.Patch("/{post_id}", chassis.SimpleHandler(s.patchPost))

	})

	r.Route("/reply", func(r chi.Router) {
		r.Post("/", chassis.SimpleHandler(s.createReply))
		r.Delete("/{reply_id}", chassis.SimpleHandler(s.deleteReply))
		r.Patch("/{reply_id}", chassis.SimpleHandler(s.patchReply))
	})



	//PATHS FOR INTERNAL USAGE ONLY
	//r.Get("/internal/user/{user_id}/upvotes", chassis.SimpleHandler(s.getUserUpvotes))
	r.Get("/internal/user/{user_id}/upvotes", chassis.SimpleHandler(s.getItemsUpvotedIds))
	r.Get("/internal/item/{item_id}/rank", chassis.SimpleHandler(s.getOverallRank))
	r.Get("/internal/upvotes", chassis.SimpleHandler(s.getAllItemsUpvotesQuantities))

	/*
	POST	/item/{subject}/post
	GET		/item/{subject}/posts
	GET		/item/{subject}/post/{post_id}


	DELETE /post/{post_id}
	PATCH /post/{post_id}

	POST   /reply
	PATCH  /reply/{reply_id}
	DELETE /reply/{reply_id}

	TODO: GET INDIVIDUALLY ONLY MAKES SENSE IN SPECIFIC SCENARIOS. CAN BE IMPLEMENTED LATER
	    SO AS GETTING ALL POSTS WITHOUT A SUBJECT
	*/

	return r
}
