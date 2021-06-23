package server

import (
	"net/http"
	"time"

	"github.com/veganbase/backend/chassis"
)

const SessionMaxAge int = 3600 * 12 * 5

func (s *Server) requestLoginEmail(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Decode request body.
	var body struct {
		Language string `json:"language"`
		Email    string `json:"email"`
	}
	err := chassis.Unmarshal(r.Body, &body)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Look up request site based on Origin header.
	site := "veganlogin"
	origins, ok := r.Header["Origin"]
	if ok && len(origins) > 0 {
		site, ok = s.SiteURLs()[origins[0]]
		if !ok {
			site = "veganlogin"
		}
	}

	// Default email language to English.
	if body.Language == "" {
		body.Language = "en"
	}

	// Create login token and publish message to trigger email sending.
	token, err := s.db.CreateLoginToken(body.Email, site, body.Language)
	if err != nil {
		return nil, err
	}

	msg := chassis.LoginEmailRequestMsg{}
	msg.FixedFields = chassis.FixedFields{
		Site:       site,
		Language:   body.Language,
		Email:      body.Email,
	}
	msg.LoginToken = token

	if err = chassis.Emit(s, "login-email-request", msg );err != nil {
		return nil, err
	}

	return chassis.NoContent(w)
}

func (s *Server) siteFromURL(url string) (string, bool) {
	return "veganbase", true
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Decode request body to get login token.
	var body struct {
		LoginToken string `json:"login_token"`
	}
	err := chassis.Unmarshal(r.Body, &body)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	// Look up login token and if not found or expired, return error.
	email, site, language, err := s.db.CheckLoginToken(body.LoginToken)
	if err != nil {
		return chassis.BadRequest(w, "Unknown login token")
	}

	// Look up user by email on the user service.
	user, err := s.userSvc.Login(email, site, language)
	if err != nil {
		return nil, err
	}

	// Create a session for the user.
	token, err := s.db.CreateSession(user.ID, user.Email, user.IsAdmin)
	if err != nil {
		return nil, err
	}

	// Set the session cookie and return the user information as a JSON
	// response.
	auth := http.Cookie{
		Name:     "session",
		Value:    token,
		MaxAge:   SessionMaxAge,
		Secure:   s.secureSession,
		HttpOnly: true,
		Path:     "/",
		//SameSite: http.SameSiteNoneMode,
	}
	http.SetCookie(w, &auth)
	return user, nil
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get session cookie. If there is no session, this is a no-op.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth {
		return chassis.NoContent(w)
	}

	// Delete single session from database.
	cookie, _ := r.Cookie("session")
	s.db.DeleteSession(cookie.Value)

	s.clearSessionCookie(w)
	return chassis.NoContent(w)
}

func (s *Server) logoutAll(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// Get session cookie. If there is no session, this is a no-op.
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod != chassis.SessionAuth {
		return chassis.NotFound(w)
	}

	// Delete all sessions from database for matching email.
	s.db.DeleteUserSessions(authInfo.UserID)

	s.clearSessionCookie(w)
	return chassis.NoContent(w)
}

// Delete session cookie by setting expiry in past.
func (s *Server) clearSessionCookie(w http.ResponseWriter) {
	delAuth := http.Cookie{
		Name:     "session",
		Value:    "",
		Secure:   s.secureSession,
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
	}
	http.SetCookie(w, &delAuth)
}
