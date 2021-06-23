package server

import (
	"fmt"
	"github.com/gavv/httpexpect"
	"github.com/veganbase/backend/services/social-service/mocks"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	dbMock = mocks.DB{}
	auth = map[string]string{
	"X-Auth-Method":   "session",
	"X-Auth-User-Id":  "usr_follower",
	"X-Auth-Is-Admin": "false",
	}
)

func RunWithServer(t *testing.T, test func(e *httpexpect.Expect)) {
	s := &Server{}
	s.Init("user-service", "dev", 8090, "dev", s.routes())
	s.db = &dbMock

	srv := httptest.NewServer(s.Srv.Handler)
	defer srv.Close()

	e := httpexpect.New(t, srv.URL)

	test(e)
}

func TestCreateSubscription(t *testing.T) {
	userID := "usr_follower"
	subscriptionID := "usr_subscription"
	RunWithServer(t, func(e *httpexpect.Expect) {
		dbMock.On("CreateUserSubscription", userID, subscriptionID).
			Return(nil)

		e.POST(fmt.Sprintf("/social/follow/%s", subscriptionID)).
			WithHeaders(auth).
			Expect().
			Status(http.StatusCreated)
	})
}

func TestDeleteSubscription(t *testing.T) {
	userID := "usr_follower"
	subscriptionID := "usr_subscription"
	RunWithServer(t, func(e *httpexpect.Expect) {
		dbMock.On("DeleteUserSubscription", userID, subscriptionID).
			Return(nil)

		e.DELETE(fmt.Sprintf("/social/forget/%s", subscriptionID)).
			WithHeaders(auth).
			Expect().
			Status(http.StatusNoContent)
	})
}

func TestListSubscriptions(t *testing.T) {
	userID := "usr_follower"
	subscriptions := []string{"usr_one","usr_two","org_one"}

	RunWithServer(t, func(e *httpexpect.Expect) {
		dbMock.On("ListUserSubscriptions", userID).
			Return(subscriptions, nil)

		e.GET(fmt.Sprintf("/social/subscriptions/")).
			WithHeaders(auth).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("subscriptions").
			Value("subscriptions").Equal(subscriptions)
	})
}

func TestListFollowers(t *testing.T) {
	userID := "usr_toFollow"
	followers := []string{"usr_one","usr_two","usr_three"}

	RunWithServer(t, func(e *httpexpect.Expect) {
		dbMock.On("ListFollowers", userID).
			Return(followers, nil)

		e.GET(fmt.Sprintf("/social/followers/%s", userID)).
			WithHeaders(auth).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("followers").
			Value("followers").Equal(followers)
	})
}
