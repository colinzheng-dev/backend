package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/chassis/pubsub"
	"github.com/veganbase/backend/services/api-gateway/db"
	"github.com/veganbase/backend/services/api-gateway/mocks"
	site_mocks "github.com/veganbase/backend/services/site-service/mocks"
	user_client "github.com/veganbase/backend/services/user-service/client"
	user_mocks "github.com/veganbase/backend/services/user-service/mocks"
	user_model "github.com/veganbase/backend/services/user-service/model"
)

var (
	messages = map[string][][]byte{}
	dbMock   = mocks.DB{}
	siteURLs = map[string]string{
		"https://ethicalbuzz.com": "ethicalbuzz",
	}
	userMock = user_mocks.Client{}
	siteMock = site_mocks.Client{}
)

func RunWithServer(t *testing.T, test func(e *httpexpect.Expect, csrf string)) {
	userSvcURL, _ := url.Parse("http://user-service")
	blobSvcURL, _ := url.Parse("http://blob-service")
	itemSvcURL, _ := url.Parse("http://item-service")
	categorySvcURL, _ := url.Parse("http://category-service")
	s := &Server{
		userSvcURL:     userSvcURL,
		blobSvcURL:     blobSvcURL,
		itemSvcURL:     itemSvcURL,
		categorySvcURL: categorySvcURL,
		siteURLs:       siteURLs,
	}
	s.Init("api-gateway", "dev", 8090, "dev", s.routes(true, "csrf-secret"))
	dbMock = mocks.DB{}
	s.db = &dbMock
	s.PubSub = pubsub.NewMockPubSub(messages)
	userMock = user_mocks.Client{}
	s.userSvc = &userMock
	siteMock = site_mocks.Client{}
	s.siteSvc = &siteMock

	srv := httptest.NewServer(s.Srv.Handler)
	defer srv.Close()

	e := httpexpect.New(t, srv.URL)

	csrf := e.GET("/").Expect().Status(http.StatusOK).
		Header("X-CSRF-Token").Raw()

	test(e, csrf)
}

func TestTokenRequest(t *testing.T) {
	RunWithServer(t, func(e *httpexpect.Expect, csrf string) {
		dbMock.
			On("CreateLoginToken", "test@testing.com", "veganlogin", "en").
			Return("123456", nil)
		dbMock.
			On("CreateLoginToken", "test@testing.com", "ethicalbuzz", "en").
			Return("654321", nil)
		dbMock.
			On("SaveEvent", mock.Anything, mock.Anything, mock.Anything).
			Return(nil).
			Run(func(args mock.Arguments) {
				inTx := args.Get(2).(func() error)
				inTx()
			})

		// No request body.
		e.POST("/auth/request-login-email").WithHeader("X-CSRF-Token", csrf).
			Expect().
			Status(http.StatusBadRequest)

		// Valid request.
		body := map[string]string{
			"language": "en",
			"email":    "test@testing.com",
		}
		e.POST("/auth/request-login-email").
			WithHeader("X-CSRF-Token", csrf).
			WithHeader("Origin", "https://veganbase.com").
			WithJSON(body).
			Expect().
			Status(http.StatusNoContent)
		token1 := checkLoginEmailPub(t, "veganlogin", "en", "test@testing.com")

		// Valid request, same email, different site.
		delete(messages, "login-email-request")
		body = map[string]string{
			"language": "en",
			"email":    "test@testing.com",
		}
		e.POST("/auth/request-login-email").WithJSON(body).
			WithHeader("X-CSRF-Token", csrf).
			WithHeader("Origin", "https://ethicalbuzz.com").
			Expect().
			Status(http.StatusNoContent)
		token2 := checkLoginEmailPub(t, "ethicalbuzz", "en", "test@testing.com")
		assert.NotEqual(t, token1, token2)
	})
}

// Check message publication for login email trigger.
func checkLoginEmailPub(t *testing.T, site, lang, email string) string {
	assert.NotNil(t, messages["login-email-request"])
	assert.Len(t, messages["login-email-request"], 1)
	msg := chassis.LoginEmailRequestMsg{}
	err := json.Unmarshal(messages["login-email-request"][0], &msg)
	assert.Nil(t, err)
	assert.Equal(t, msg.Site, site)
	assert.Equal(t, msg.Language, lang)
	assert.Equal(t, msg.Email, email)
	assert.Len(t, msg.LoginToken, 6)
	return msg.LoginToken
}

func TestLogout(t *testing.T) {
	RunWithServer(t, func(e *httpexpect.Expect, csrf string) {
		dbMock.
			On("LookupSession", "UNKNOWN").
			Return("", "", false, db.ErrSessionNotFound)
		dbMock.
			On("LookupSession", "SESSION1").
			Return("usr_TEST1", "test1@testing.com", false, nil)
		dbMock.
			On("DeleteSession", "SESSION1").
			Return(nil)
		dbMock.
			On("LookupSession", "SESSION2").
			Return("usr_TEST2", "tes2@testing.com", false, nil)
		dbMock.
			On("DeleteUserSessions", "usr_TEST2").
			Return(nil)

		// Invalid session => does nothing.
		e.POST("/auth/logout").WithCookie("session", "UNKNOWN").
			WithHeader("X-CSRF-Token", csrf).
			Expect().
			Status(http.StatusNoContent)

		// Valid session => logout.
		e.POST("/auth/logout").WithCookie("session", "SESSION1").
			WithHeader("X-CSRF-Token", csrf).
			Expect().
			Status(http.StatusNoContent)

		// Logout all => logout.
		e.POST("/auth/logout-all").WithCookie("session", "SESSION2").
			WithHeader("X-CSRF-Token", csrf).
			Expect().
			Status(http.StatusNoContent)
	})
}

func TestLogin(t *testing.T) {
	RunWithServer(t, func(e *httpexpect.Expect, csrf string) {
		dbMock.
			On("CheckLoginToken", "BAD000").
			Return("", "", "", db.ErrLoginTokenNotFound)
		dbMock.
			On("CheckLoginToken", "EXP000").
			Return("", "", "", db.ErrLoginTokenNotFound)
		dbMock.
			On("CheckLoginToken", "VAL000").
			Return("test1@example.com", "veganlogin", "en", nil)
		userMock.
			On("Login", "test1@example.com", "veganlogin", "en").
			Return(&user_client.LoginResponse{
				&user_model.User{
					ID:      "usr_USER1",
					Email:   "test1@example.com",
					IsAdmin: false,
				},
				false,
			}, nil)
		dbMock.
			On("CreateSession", "usr_USER1", "test1@example.com", false).
			Return("123456", nil)

		// No request body => bad request.
		e.POST("/auth/login").
			WithHeader("X-CSRF-Token", csrf).
			Expect().
			Status(http.StatusBadRequest)

		// Unknown token => bad request.
		body := map[string]string{
			"login_token": "BAD000",
		}
		e.POST("/auth/login").WithJSON(body).
			WithHeader("X-CSRF-Token", csrf).
			Expect().
			Status(http.StatusBadRequest)

		// Expired token => bad request.
		body = map[string]string{
			"login_token": "EXP000",
		}
		e.POST("/auth/login").WithJSON(body).
			WithHeader("X-CSRF-Token", csrf).
			Expect().
			Status(http.StatusBadRequest)

		// Valid token.
		body = map[string]string{
			"login_token": "VAL000",
		}
		e.POST("/auth/login").WithJSON(body).
			WithHeader("X-CSRF-Token", csrf).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("id").ValueEqual("id", "usr_USER1").
			ContainsKey("email").ValueEqual("email", "test1@example.com").
			ContainsKey("is_admin").ValueEqual("is_admin", false)
	})
}
