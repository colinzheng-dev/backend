package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gavv/httpexpect"
	"github.com/stretchr/testify/mock"
	"github.com/veganbase/backend/services/user-service/db"
	"github.com/veganbase/backend/services/user-service/mocks"
	"github.com/veganbase/backend/services/user-service/model"
)

var (
	name1     = "Test User 1"
	name2     = "Test User 2"
	name3     = "Test User 3"
	avatar1   = "http://img.veganapi.com/avatar-32.png"
	avatar2   = "http://img.veganapi.com/avatar-22.png"
	avatar3   = "http://img.veganapi.com/avatar-15.png"
	loginTime = time.Date(2019, 6, 10, 12, 0, 0, 0, time.UTC)
	u1        = model.User{
		ID:        "usr_TESTUSER1",
		Email:     "user1@test.com",
		Name:      &name1,
		IsAdmin:   false,
		Avatar:    &avatar1,
		LastLogin: time.Date(2019, 6, 10, 12, 0, 0, 0, time.UTC),
	}
	p1        = model.PayoutAccount{
		ID:        "acc_123456789",
		AccountNumber:     "pac_14saf1TEST",
		Owner:      "usr_TESTUSER1",
		CreatedAt: time.Date(2019, 6, 10, 12, 0, 0, 0, time.UTC),
	}
	p2        = model.PayoutAccount{
		ID:        "acc_1234567890",
		AccountNumber:     "pac_14saf1TEST2",
		Owner:      "usr_TESTUSER2",
		CreatedAt: time.Date(2019, 6, 10, 12, 0, 0, 0, time.UTC),
	}
	u2 = model.User{
		ID:        "usr_TESTUSER2",
		Email:     "user2@test.com",
		Name:      &name2,
		IsAdmin:   true,
		Avatar:    &avatar2,
		LastLogin: loginTime,
	}
	u3 = model.User{
		ID:        "usr_TESTUSER3",
		Email:     "user3@test.com",
		Name:      &name3,
		IsAdmin:   false,
		Avatar:    &avatar3,
		LastLogin: loginTime,
	}
	sess = map[string]string{
		"X-Auth-Method":   "session",
		"X-Auth-User-Id":  "usr_TESTUSER1",
		"X-Auth-Is-Admin": "false",
	}
	sessAdmin = map[string]string{
		"X-Auth-Method":   "session",
		"X-Auth-User-Id":  "usr_TESTUSER2",
		"X-Auth-Is-Admin": "true",
	}
	sess3 = map[string]string{
		"X-Auth-Method":   "session",
		"X-Auth-User-Id":  "usr_TESTUSER3",
		"X-Auth-Is-Admin": "true",
	}
	dbMock = mocks.DB{}
)

func RunWithServer(t *testing.T, test func(e *httpexpect.Expect)) {
	s := &Server{
		avatarGen: generateAvatar("http://img.veganapi.com/avatar-%02d.png", 48),
	}
	s.Init("user-service", "dev", 8090, "dev", s.routes())
	dbMock = mocks.DB{}
	s.db = &dbMock

	srv := httptest.NewServer(s.Srv.Handler)
	defer srv.Close()

	e := httpexpect.New(t, srv.URL)

	test(e)
}

// Test login route.
func TestLogin(t *testing.T) {
	RunWithServer(t, func(e *httpexpect.Expect) {
		newName := "test@example.com"
		newAvatar := "http://img.veganapi.com/avatar-22.png"
		dbMock.
			On("LoginUser", "test@example.com", mock.Anything).
			Return(&model.User{
				ID:        "usr_NEWUSER",
				Email:     "test@example.com",
				Name:      &newName,
				IsAdmin:   false,
				Avatar:    &newAvatar,
				LastLogin: time.Date(2019, 6, 10, 12, 0, 0, 0, time.UTC),
			}, true, nil)
		dbMock.
			On("LoginUser", "user1@test.com", mock.Anything).
			Return(&u1, false, nil)
		dbMock.
			On("SaveEvent", mock.Anything, mock.Anything, mock.Anything).
			Return(nil)

		// Invalid method.
		e.GET("/login").
			Expect().
			Status(http.StatusMethodNotAllowed)

		// Empty body.
		e.POST("/login").
			Expect().
			Status(http.StatusBadRequest)

		// New user => success.
		o := e.POST("/login").
			WithJSON(map[string]string{"email": "test@example.com"}).
			Expect().
			Status(http.StatusOK).
			JSON().Object()
		o.ContainsKey("id").
			ContainsKey("email").ValueEqual("email", "test@example.com").
			ContainsKey("is_admin").ValueEqual("is_admin", false).
			ContainsKey("avatar")
		o.Value("avatar").String().
			Match("http://img.veganapi.com/avatar-[0-9]{2}.png").
			Length().Equal(1)

		// Existing user => success.
		e.POST("/login").
			WithJSON(map[string]string{"email": "user1@test.com"}).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("id").ValueEqual("id", "usr_TESTUSER1").
			ContainsKey("email").ValueEqual("email", "user1@test.com").
			ContainsKey("is_admin").ValueEqual("is_admin", false).
			ContainsKey("last_login")
	})
}

// Test permissions and returns for "GET /me" user profile route.
func TestDetail(t *testing.T) {
	RunWithServer(t, func(e *httpexpect.Expect) {
		emptyOrgs := []*model.OrgWithUserInfo{}
		dbMock.On("UserByID", "usr_TESTUSER1").Return(&u1, nil)
		dbMock.On("UserOrgs", mock.Anything).Return(emptyOrgs, nil)
		dbMock.On("UserByID", "usr_TESTUSER2").Return(&u2, nil)
		dbMock.On("UserByID", mock.Anything).Return(nil, db.ErrUserNotFound)

		var tests = []struct {
			path    string
			headers *map[string]string
			status  int
			id      string
			full    bool
		}{
			// Unauthenticated => not found.
			{"/me", nil, http.StatusNotFound, "", false},

			// Unauthenticated => public profile.
			{"/user/usr_TESTUSER1", nil, http.StatusOK, "usr_TESTUSER1", false},

			// Authenticated => normal access.
			{"/me", &sess, http.StatusOK, "usr_TESTUSER1", true},

			// Same user => access OK.
			{"/user/usr_TESTUSER1", &sess, http.StatusOK, "usr_TESTUSER1", true},

			// Different user => public profile.
			{"/user/usr_TESTUSER2", &sess, http.StatusOK, "usr_TESTUSER2", false},

			// Admin => normal access.
			{"/me", &sessAdmin, http.StatusOK, "usr_TESTUSER2", true},

			// Admin + different user => access OK.
			{"/user/usr_TESTUSER1", &sessAdmin, http.StatusOK, "usr_TESTUSER1", true},

			// Admin + same user => access OK.
			{"/user/usr_TESTUSER2", &sessAdmin, http.StatusOK, "usr_TESTUSER2", true},
		}

		for _, test := range tests {
			req := e.GET(test.path)
			if test.headers != nil {
				req = req.WithHeaders(*test.headers)
			}
			rsp := req.Expect().
				Status(test.status)
			if test.id != "" {
				flds := rsp.JSON().Object()
				flds.ContainsKey("id").ValueEqual("id", test.id)
				if !test.full {
					flds.NotContainsKey("email")
					flds.NotContainsKey("api_key")
					flds.NotContainsKey("name")
					flds.NotContainsKey("last_login")
				}
			}
		}
	})
}

func TestList(t *testing.T) {
	RunWithServer(t, func(e *httpexpect.Expect) {
		dbMock.
			On("Users", "", uint(0), uint(0)).
			Return([]model.User{u1, u2, u3}, nil)
		dbMock.
			On("Users", "user1", uint(0), uint(0)).
			Return([]model.User{u1}, nil)

		e.GET("/users").WithHeaders(sessAdmin).
			Expect().
			Status(http.StatusOK).
			JSON().Array().Length().Equal(3)

		e.GET("/users").WithQuery("q", "user1").WithHeaders(sessAdmin).
			Expect().
			Status(http.StatusOK).
			JSON().Array().Length().Equal(1)
	})
}

// Test permissions and returns for "PATCH /me" user profile update
// route.
func TestUpdate(t *testing.T) {
	RunWithServer(t, func(e *httpexpect.Expect) {
		dbMock.On("UserByID", "usr_TESTUSER1").Return(&u1, nil)
		dbMock.On("UserByID", "usr_TESTUSER2").Return(&u2, nil)
		dbMock.On("UserByID", mock.Anything).Return(nil, db.ErrUserNotFound)
		dbMock.
			On("UpdateUser", mock.Anything).Return(nil)
		dbMock.
			On("SaveEvent", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		var tests1 = []struct {
			path    string
			headers *map[string]string
			status  int
		}{
			// Unauthenticated => not found.
			{"/me", nil, http.StatusNotFound},

			// Authenticated => normal access.
			{"/me", &sess, http.StatusOK},

			// Same user => access OK.
			{"/user/usr_TESTUSER1", &sess, http.StatusOK},

			// Different user => access denied.
			{"/user/usr_TESTUSER2", &sess, http.StatusNotFound},

			// Admin => normal access.
			{"/me", &sessAdmin, http.StatusOK},

			// Admin + different user => access OK.
			{"/user/usr_TESTUSER1", &sessAdmin, http.StatusOK},

			// Admin + same user => access OK.
			{"/user/usr_TESTUSER2", &sessAdmin, http.StatusOK},
		}

		for i, test := range tests1 {
			newName := fmt.Sprintf("New test name %d", i+1)
			req := e.PATCH(test.path).
				WithJSON(map[string]string{"name": newName})
			if test.headers != nil {
				req = req.WithHeaders(*test.headers)
			}
			rsp := req.Expect().
				Status(test.status)
			if test.status == http.StatusOK {
				rsp.JSON().Object().ContainsKey("name").ValueEqual("name", newName)
			}
		}

		// Try updating fields that we're not allowed to update.
		var tests2 = []struct {
			field string
			value string
		}{
			{"email", "not@allowed.net"},
			{"id", "usr_RANDOM"},
			{"last_login", "2019-01-01 09:00:00"},
		}
		for _, test := range tests2 {
			e.PATCH("/me").WithHeaders(sess).
				WithJSON(map[string]string{test.field: test.value}).
				Expect().
				Status(http.StatusForbidden)
			e.PATCH("/me").WithHeaders(sessAdmin).
				WithJSON(map[string]string{test.field: test.value}).
				Expect().
				Status(http.StatusForbidden)
		}

		// Test the rules for updating the admin status of users: ordinary
		// users cannot modify their is_admin flag; administrators can
		// modify the is_admin flag of other users, but cannot change
		// their own administrator status.
		e.PATCH("/me").WithHeaders(sess).
			WithJSON(map[string]string{"is_admin": "true"}).
			Expect().
			Status(http.StatusForbidden)
		e.PATCH("/me").WithHeaders(sessAdmin).
			WithJSON(map[string]string{"is_admin": "false"}).
			Expect().
			Status(http.StatusForbidden)
		e.PATCH("/user/usr_TESTUSER1").WithHeaders(sessAdmin).
			WithJSON(map[string]interface{}{"is_admin": true}).
			Expect().
			Status(http.StatusOK)
		// assert.Equal(t, users["usr_TESTUSER1"].IsAdmin, true)
	})
}

func TestDelete(t *testing.T) {
	RunWithServer(t, func(e *httpexpect.Expect) {
		dbMock.On("UserByID", "usr_TESTUSER1").Return(&u1, nil)
		dbMock.On("UserByID", "usr_TESTUSER2").Return(&u2, nil)
		dbMock.On("UserByID", "usr_TESTUSER3").Return(&u3, nil)
		dbMock.On("UserByID", mock.Anything).Return(nil, db.ErrUserNotFound)
		dbMock.On("DeleteUser", "usr_TESTUSER1").Return(nil)
		dbMock.On("DeleteUser", "usr_TESTUSER2").Return(nil)
		dbMock.On("DeleteUser", "usr_TESTUSER3").Return(nil)
		dbMock.
			On("SaveEvent", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		// Other user => not allowed.
		e.DELETE("/user/usr_TESTUSER2").WithHeaders(sess).
			Expect().
			Status(http.StatusNotFound)

		// Delete self => success.
		e.DELETE("/me").WithHeaders(sess3).
			Expect().
			Status(http.StatusNoContent)
		// assert.NotContains(t, users, "usr_TESTUSER3")
	})
}

func setupAPIKeys(k string) {
	dbMock.ExpectedCalls = []*mock.Call{}
	dbMock.On("UserByID", "usr_TESTUSER1").Return(&u1, nil)
	dbMock.On("UserByID", "usr_TESTUSER2").Return(&u2, nil)
	dbMock.On("UserByID", "usr_TESTUSER3").Return(&u3, nil)
	dbMock.On("UserByID", mock.Anything).Return(nil, db.ErrUserNotFound)
	dbMock.On("SaveHashedAPIKey", mock.Anything, mock.Anything, "usr_TESTUSER1").Return(nil)
	dbMock.On("SaveHashedAPIKey", mock.Anything, mock.Anything, "usr_TESTUSER2").Return(nil)
	dbMock.On("SaveHashedAPIKey", mock.Anything, mock.Anything, "usr_TESTUSER3").Return(nil)
	dbMock.On("SaveHashedAPIKey", mock.Anything, mock.Anything, "usr_TESTUSER4").Return( db.ErrUserNotFound)
	dbMock.On("DeleteAPIKey", mock.Anything).Return(nil)
	dbMock.
		On("SaveEvent", mock.Anything, mock.Anything, mock.Anything).Return(nil)
}

func TestAPIKeys(t *testing.T) {
	RunWithServer(t, func(e *httpexpect.Expect) {
		setupAPIKeys("abcdef")

		// Create an API key.
		e.POST("/me/api-key").WithHeaders(sess).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("api_key").Value("api_key").String()

		setupAPIKeys("123456")

		// Rotate API key.
		e.POST("/me/api-key").WithHeaders(sess).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("api_key").Value("api_key").String()

		// Delete API key.
		e.DELETE("/me/api-key").WithHeaders(sess).
			Expect().
			Status(http.StatusNoContent)

		// Admin + unknown user => not found.
		e.POST("/user/usr_TESTUSER4/api-key").WithHeaders(sessAdmin).
			Expect().
			Status(http.StatusNotFound)
	})
}

//
//
//func setupPayoutAccounts(k string) {
//	dbMock.ExpectedCalls = []*mock.Call{}
//	dbMock.On("CreatePayoutAccount", &model.PayoutAccount{AccountNumber:k, Owner:"usr_TESTUSER1"}).Return(&p1, nil)
//	dbMock.On("PayoutAccountByOwner", "usr_TESTUSER1").Return(&p1, nil)
//	dbMock.On("PayoutAccountByOwner", "usr_TESTUSER2").Return(nil, db.ErrPayoutAccountNotFound)
//	dbMock.On("CreatePayoutAccount", &model.PayoutAccount{AccountNumber:p2.AccountNumber, Owner:"usr_TESTUSER2"}).Return(&p2, nil)
//
//
//}
//
//func TestPayoutAccounts(t *testing.T) {
//	RunWithServer(t, func(e *httpexpect.Expect) {
//		setupPayoutAccounts("pac_14saf1TEST")
//
//		//Create a payout account.
//		//e.POST("/me/payout-account").WithHeaders(sess).
//		//	WithJSON(map[string]string{"account": "pac_14saf1TEST"}).
//		//	Expect().
//		//	Status(http.StatusOK).
//		//	JSON().Object().
//		//	ContainsKey("id").Value("id").String().Equal("acc_123456789")
//
//		e.GET("/me/payout-account").WithHeaders(sessAdmin).
//			Expect().
//			Status(http.StatusOK).
//			JSON().Object().ContainsKey("id").Value("id").Equal(p2.ID)
///*
//		setupAPIKeys("123456")
//
//		// Rotate API key.
//		e.POST("/me/api-key").WithHeaders(sess).
//			Expect().
//			Status(http.StatusOK).
//			JSON().Object().
//			ContainsKey("api_key").Value("api_key").String().Equal("123456")
//
//		// Delete API key.
//		e.DELETE("/me/api-key").WithHeaders(sess).
//			Expect().
//			Status(http.StatusNoContent)
//
//		// Admin + unknown user => not found.
//		e.POST("/user/usr_TESTUSER4/api-key").WithHeaders(sessAdmin).
//			Expect().
//			Status(http.StatusNotFound)*/
//	})
//}
