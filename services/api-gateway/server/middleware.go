package server

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/csrf"
	"github.com/rs/cors"

	"github.com/veganbase/backend/chassis"
)

// Add CSRF protection to API gateway.
func (s *Server) addCSRFMiddleware(r chi.Router, devMode bool, csrfSecret string) {

	// Add middleware to implement double token submission CSRF
	// protection. One token is stored in a secure HTTP-only cookie
	// (security switched off for local development...), and the other is
	// sent in each response in an X-CSRF-Token header. This second token
	// must be submitted in each request that makes state changes by
	// adding an X-CSRF-Token header to the request.
	options := []csrf.Option{}
	options = append(options, csrf.CookieName("csrftoken"))
	options = append(options, csrf.Path("/"))
	if devMode {
		options = append(options, csrf.Secure(false))
	}

	options = append(options, csrf.MaxAge(3600*24*7)) //TEST MAXAGE
	//options = append(options, csrf.SameSite(csrf.SameSiteNoneMode))

	csrfMiddleware := csrf.Protect([]byte(csrfSecret), options...)
	r.Use(csrfMiddleware)

	// Extra middleware to add CSRF token to all responses.
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-CSRF-Token", csrf.Token(r))
			next.ServeHTTP(w, r)
		})
	})
}

// Add middleware specific to API gateway.
// Composed by CORS protection, RealIP and logging features.
func (s *Server) addCORSMiddleware(r chi.Router) {
	r.Use(middleware.RealIP)

	// Add common middleware.
	chassis.AddCommonMiddleware(r, false)

	// Very basic concurrent request service throttling.
	// TODO: PER-ROUTE, PER-IP RATE LIMITING.
	r.Use(middleware.Throttle(1000))

	// Add CORS configuration. The origin checking function uses a list
	// of permitted origins constructed from sites registered with the
	// site service (plus a list from the API gateway's configuration).
	opts := cors.Options{
		AllowOriginRequestFunc: s.checkCORS(),
		AllowCredentials:       true,
		AllowedMethods:         []string{"POST", "PUT", "PATCH", "GET", "DELETE"},
		// TODO: SORT THIS NEXT SETTING OUT...
		AllowedHeaders: []string{"*"},
		ExposedHeaders: []string{"X-CSRF-Token"},
	}
	// if s.relaxCORSKey != "no" {
	// 	opts.AllowedHeaders = []string{"*"}
	// }
	co := cors.New(opts)
	r.Use(co.Handler)
}

// CORS origin checking is based on a list of known sites. Origin
// checking can be bypassed for development purposes by setting an
// X-Relax-Cors header containing a secret key value set up on server
// start.
func (s *Server) checkCORS() func(r *http.Request, origin string) bool {
	return func(r *http.Request, origin string) bool {
		// Check known site origins first for speed.
		origins := s.CORSOrigins()
		_, ok := origins[origin]
		if ok {
			return true
		}

		shimPrefix := "https://www."

		if strings.HasPrefix(origin, shimPrefix) {
			if _, ok := origins[strings.Replace(origin, shimPrefix, "https://", 1)]; ok {
				return true
			}
		}

		return false
	}
}

// CredentialCtx extracts credential information from the request
// (either via a session cookie or an API key), looks up the
// corresponding user information and injects the resulting
// authorisation information into the request context as a
// chassis.AuthInfo value.
func CredentialCtx(s *Server) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get session cookie.
			session := ""
			if cookie, err := r.Cookie("session"); err == nil {
				session = cookie.Value
			}

			// Get API key header.
			apiKey := ""
			if apiKeys, ok := r.Header["X-Api-Key"]; ok {
				apiKey = apiKeys[0]
			}
			apiSecret := ""
			if apiSecrets, ok := r.Header["X-Api-Secret"]; ok {
				apiSecret = apiSecrets[0]
			}

			if session == "" && apiKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			authInfo := chassis.AuthInfo{}
			if session != "" {
				if userID, _, isAdmin, err := s.db.LookupSession(session); err == nil {
					authInfo.AuthMethod = chassis.SessionAuth
					authInfo.UserID = userID
					authInfo.UserIsAdmin = isAdmin
				}
			} else {
				//TODO: CACHE MAY BE WORTH IT
				if user, err := s.userSvc.GetUserByApiKey(apiKey, apiSecret); err == nil {
					authInfo.AuthMethod = chassis.SessionAuth
					authInfo.UserID = user.ID
					authInfo.UserIsAdmin = user.IsAdmin
				}
			}
			next.ServeHTTP(w, r.WithContext(chassis.NewAuthContext(r.Context(), &authInfo)))
		})
	}
}
