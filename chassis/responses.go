package chassis

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

// JSON error response.
type ErrResp struct {
	Message string `json:"message"`
}

// BadRequest sets up an HTTP 400 Bad Request with a given error
// message and returns the (nil, nil) pair used by SimpleHandler to
// signal that the response has been dealt with.
func BadRequest(w http.ResponseWriter, msg string) (interface{}, error) {
	rsp := ErrResp{msg}
	body, _ := json.Marshal(rsp)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	w.Write(body)
	return nil, nil
}

// NotFound sets up an HTTP 404 Not Found and returns the (nil, nil)
// pair used by SimpleHandler to signal that the response has been
// dealt with.
func NotFound(w http.ResponseWriter) (interface{}, error) {
	http.NotFound(w, nil)
	return nil, nil
}

// NotFoundWithMessage sets up an HTTP 404 Not Found with a given error
// message and returns the (nil, nil) pair used by SimpleHandler to
// signal that the response has been dealt with.
func NotFoundWithMessage(w http.ResponseWriter, msg string) (interface{}, error) {
	rsp := ErrResp{msg}
	body, _ := json.Marshal(rsp)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write(body)
	return nil, nil
}

// Forbidden sets up an HTTP 403 Forbidden and returns the (nil, nil)
// pair used by SimpleHandler to signal that the response has been
// dealt with.
func Forbidden(w http.ResponseWriter) (interface{}, error) {
	w.WriteHeader(http.StatusForbidden)
	return nil, nil
}

// Unauthorized sets up an HTTP 401 StatusUnauthorized and returns the (nil, nil)
// pair used by SimpleHandler to signal that the response has been dealt with.
func Unauthorized(w http.ResponseWriter, msg string) (interface{}, error) {
	rsp := ErrResp{msg}
	body, _ := json.Marshal(rsp)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write(body)
	return nil, nil
}

// NoContent sets up an HTTP 204 No Content and returns the (nil, nil)
// pair used by SimpleHandler to signal that the response has been
// dealt with.
func NoContent(w http.ResponseWriter) (interface{}, error) {
	w.WriteHeader(http.StatusNoContent)
	return nil, nil
}

// BadRequest sets up an HTTP 400 Bad Request with a given error
// message and returns the (nil, nil) pair used by SimpleHandler to
// signal that the response has been dealt with.
func InternalServerError(w http.ResponseWriter, err error) (interface{}, error) {
	log.Warn().Msgf("internal error: %s", err)
	w.WriteHeader(http.StatusInternalServerError)
	return nil, nil
}
