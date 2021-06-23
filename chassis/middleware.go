package chassis

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

// AddCommonMiddleware adds common middleware for all routes in all
// services. (This is intended to include *only* middleware that
// should be used both for internal services and those exposed
// externally. It's mostly about getting a consistent logging story to
// help with log aggregation.)
func AddCommonMiddleware(r chi.Router, authLogs bool) {
	// Set up zerolog request logging.
	r.Use(hlog.NewHandler(log.Logger))
	logs := func(r *http.Request, status, size int, duration time.Duration) {
		basicRequestLog(r, status, size, duration).Msg("")
	}
	if authLogs {
		logs = func(r *http.Request, status, size int, duration time.Duration) {
			basic := basicRequestLog(r, status, size, duration)
			authLog(r, basic).Msg("")
		}
	}
	r.Use(hlog.AccessHandler(logs))
	r.Use(hlog.RemoteAddrHandler("ip"))
	r.Use(hlog.UserAgentHandler("user_agent"))
	r.Use(hlog.RefererHandler("referer"))

	r.Use(logHandler)

	// Add correlation ID to requests.
	r.Use(hlog.RequestIDHandler("request_id", "Request-Id"))

	// Panic recovery.
	r.Use(middleware.Recoverer)

	// QUESTION: I DON'T THINK THIS WILL WORK WITHOUT MAKING ALL THE
	// HANDLERS CONTEXT-AWARE. WHAT'S THE RIGHT WAY TO DEAL WITH THIS?
	// r.Use(middleware.Timeout(60 * time.Second))
}

func logHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		x, err := httputil.DumpRequest(r, false)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
		save := r.Body
		if r.Body == nil {
			r.Body = nil
		} else {
			save, r.Body, _ = drainBody(r.Body)
		}
		reqBody, err := ioutil.ReadAll(save)
		if !utf8.Valid(reqBody) {
			reqBody = []byte("** BINARY DATA IN BODY **")
		}
		x = append(x, reqBody...)
		log.Info().Str("dir", "request").Msg(string(x))
		rec := httptest.NewRecorder()
		next.ServeHTTP(rec, r)
		resp := fmt.Sprintf("%d\n", rec.Code)
		for k, v := range rec.HeaderMap {
			resp += k + ": " + strings.Join(v, ",") + "\n"
		}
		body := rec.Body.String()
		if !utf8.Valid([]byte(body)) {
			body = "** BINARY DATA IN BODY **"
		}
		log.Info().Str("dir", "response").Msg(resp + body)

		// this copies the recorded response to the response writer
		for k, v := range rec.HeaderMap {
			w.Header()[k] = v
		}
		w.WriteHeader(rec.Code)
		rec.Body.WriteTo(w)
	})
}

func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	if b == http.NoBody {
		// No copying needed. Preserve the magic sentinel meaning of NoBody.
		return http.NoBody, http.NoBody, nil
	}
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}
	if err = b.Close(); err != nil {
		return nil, b, err
	}
	return ioutil.NopCloser(&buf), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

// Basic HTTP request logging.
func basicRequestLog(r *http.Request, status, size int, duration time.Duration) *zerolog.Event {
	if r.URL.Path == "/healthz" {
		return nil
	}
	return hlog.FromRequest(r).Info().
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Int("status", status).
		Int("size", size).
		Dur("duration", duration)
}

// Additional logging to record authentication information in backend
// services.
func authLog(r *http.Request, e *zerolog.Event) *zerolog.Event {
	method, ok := r.Header["X-Auth-Method"]
	if ok {
		e = e.Str("auth-method", method[0])
	}
	userID, ok := r.Header["X-Auth-User-Id"]
	if ok {
		e = e.Str("auth-user-id", userID[0])
	}
	isAdmin, ok := r.Header["X-Auth-Is-Admin"]
	if ok {
		e = e.Bool("auth-is-admin", isAdmin[0] == "true")
	}
	return e
}

// AuthMethod is an enumerated type that distinguishes between
// session-based authentication and authentication by API key.
type AuthMethod int

// AuthMethod enumeration values for no authentication, session
// authentication and API key authentication.
const (
	NoAuth AuthMethod = iota
	SessionAuth
	APIKeyAuth
)

// AuthInfo carries all the authentication information passed from the
// API gateway to internal services. This provides enough information
// to do authentication and RBAC and ACL authorisation checking.
type AuthInfo struct {
	// Authentication method used for request (one of none, session or
	// API key).
	AuthMethod AuthMethod

	// User ID for the authenticated user making the request (will be
	// empty for an unauthenticated request).
	UserID string

	// Is the authenticated user an administrator?
	UserIsAdmin bool
}

// ctxKey is a key type for request context information.
type ctxKey int

// Request context keys that we use.
const (
	authInfoCtxKey ctxKey = iota
)

// AuthCtx injects authentication information derived from
// inter-service HTTP headers into the request context. Information is
// extracted from the following headers:
//
// X-Auth-Method:
//
// If present, this will be one of "session", "api-key" or
// "service-client"; any other values are treated as "none".
//
// X-Auth-User-Id
//
// If present, this is the user ID of the authenticated user making
// the request.
//
// X-Auth-Is-Admin
//
// If present, this is a boolean flag marking whether the
// authenticated user is an administrator.
func AuthCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authMethod := NoAuth
		userID := ""
		isAdmin := false

		// Authentication method header.
		methodHeader, ok := r.Header["X-Auth-Method"]
		isServiceClient := false
		if ok && len(methodHeader) > 0 {
			switch methodHeader[0] {
			default:
				authMethod = NoAuth
			case "session":
				authMethod = SessionAuth
			case "api-key":
				authMethod = APIKeyAuth
			case "service-client":
				isServiceClient = true
				authMethod = SessionAuth
				userID = "service-client"
				isAdmin = true
			}
		}

		// User ID header.
		if !isServiceClient {
			userIDHeader, ok := r.Header["X-Auth-User-Id"]
			if ok && len(userIDHeader) > 0 {
				userID = userIDHeader[0]
			}
		}

		// Administrator status header.
		if !isServiceClient {
			isAdminHeader, ok := r.Header["X-Auth-Is-Admin"]
			if ok && len(isAdminHeader) > 0 {
				if strings.ToLower(isAdminHeader[0]) == "true" {
					isAdmin = true
				}
			}
		}

		// Create and inject authentication information.
		authInfo := &AuthInfo{authMethod, userID, isAdmin}
		next.ServeHTTP(w, r.WithContext(NewAuthContext(r.Context(), authInfo)))
	})
}

// NewAuthContext returns a new Context that carries authentication
// information.
func NewAuthContext(ctx context.Context, info *AuthInfo) context.Context {
	return context.WithValue(ctx, authInfoCtxKey, info)
}

// AuthInfoFromContext returns the authentication information in a
// context, if any.
func AuthInfoFromContext(ctx context.Context) *AuthInfo {
	info, ok := ctx.Value(authInfoCtxKey).(*AuthInfo)
	if !ok {
		info = &AuthInfo{}
	}
	return info
}
