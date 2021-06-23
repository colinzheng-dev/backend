package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/veganbase/backend/chassis"
)

// Forward proxies a local URL to the same URL on another server,
// passing authentication information through from the request context
// using X-Auth-... headers.
func Forward(svcURL *url.URL) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(svcURL)
	baseDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		baseDirector(r)
		fixForwardedRequest(r)
	}
	return proxy
}

func fixForwardedRequest(r *http.Request) {
	// Set X-Auth-... headers on forwarded request.
	r.Header.Del("X-Auth-Method")
	r.Header.Del("X-Auth-User-Id")
	r.Header.Del("X-Auth-Is-Admin")
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return
	}
	if authInfo.AuthMethod == chassis.SessionAuth {
		r.Header.Add("X-Auth-Method", "session")
	}
	if authInfo.AuthMethod == chassis.APIKeyAuth {
		r.Header.Add("X-Auth-Method", "api-key")
	}
	r.Header.Add("X-Auth-User-Id", authInfo.UserID)
	r.Header.Add("X-Auth-Is-Admin", fmt.Sprintf("%t", authInfo.UserIsAdmin))
}
