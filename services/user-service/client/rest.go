package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/chassis/pubsub"
	"github.com/veganbase/backend/services/user-service/events"
	"github.com/veganbase/backend/services/user-service/model"
)

// RESTClient is a user service client that connects via REST.
type RESTClient struct {
	baseURL      string
	infoCache    *chassis.Cache
	userOrgCache *chassis.Cache
}

// New creates a new user service client.
func New(baseURL string, ps pubsub.PubSub, appName string) (*RESTClient, error) {
	infoCache, err := chassis.NewCache(256, ps, events.UserCacheInvalTopic, appName)
	if err != nil {
		return nil, err
	}
	userOrgCache, err := chassis.NewCache(256, ps, events.UserCacheInvalTopic, appName)
	if err != nil {
		return nil, err
	}
	return &RESTClient{
		baseURL:      baseURL,
		infoCache:    infoCache,
		userOrgCache: userOrgCache,
	}, nil
}

// Login invokes the login method on the user service.
func (c *RESTClient) Login(email string,
	site string, language string) (*LoginResponse, error) {
	// Encode JSON request body.
	req := struct {
		Email    string `json:"email"`
		Site     string `json:"site"`
		Language string `json:"language"`
	}{email, site, language}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// Do POST to endpoint.
	buf := bytes.NewBuffer(body)
	rsp, err := http.Post(c.baseURL+"/login", "application/json", buf)
	if err != nil {
		return nil, err
	}

	// Decode response.
	resp := LoginResponse{}
	rspBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	err = json.Unmarshal(rspBody, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// Info invokes the user info method on the user service.
func (c *RESTClient) Info(ids []string) (map[string]*model.Info, error) {
	fmt.Println("---> ids =", ids)
	// Look up IDs in cache.
	resp := map[string]*model.Info{}
	reqIDs := map[string]bool{}
	for _, id := range ids {
		if info, ok := c.infoCache.Get(id); ok {
			uinfo := info.(*model.Info)
			resp[id] = uinfo
		} else {
			reqIDs[id] = true
		}
	}

	if len(reqIDs) == 0 {
		return resp, nil
	}

	// Encode query parameters.
	is := []string{}
	for k := range reqIDs {
		is = append(is, k)
	}
	idparam := strings.Join(is, ",")

	// Do GET to endpoint.
	rsp, err := http.Get(c.baseURL + "/info?ids=" + idparam)
	if err != nil {
		return nil, err
	}

	// Decode response.
	resp2 := map[string]*model.Info{}
	rspBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	fmt.Println("---> rspBody =", string(rspBody))
	err = json.Unmarshal(rspBody, &resp2)
	if err != nil {
		return nil, err
	}

	// Add new entries to cache.
	for k, v := range resp2 {
		resp[k] = v
		c.infoCache.Set(k, v)
	}

	return resp, nil
}

// OrgsForUser invokes the user organisation membership list method on
// the user service and returns a map from organisation names to admin
// status flags.
func (c *RESTClient) OrgsForUser(id string) (map[string]bool, error) {
	// Look up user ID in cache.
	//if orgs, ok := c.userOrgCache.Get(id); ok {
	//	return orgs.(map[string]bool), nil
	//}

	url := fmt.Sprintf("%s/user/%s/orgs", c.baseURL, id)

	// Do GET to endpoint.
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Auth-Method", "service-client")
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Decode response.
	resp := []model.OrgWithUserInfo{}
	rspBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	if err = json.Unmarshal(rspBody, &resp); err != nil {
		return nil, err
	}

	result := map[string]bool{}
	for _, o := range resp {
		result[o.ID] = o.UserIsAdmin
	}

	// Add new entry to cache.
	//c.userOrgCache.Set(id, result)

	return result, nil
}

// IsUserOrgMember checks to see whether a user is a member of an
// organisation.
func (c *RESTClient) IsUserOrgMember(userID string, orgID string) (bool, error) {
	orgs, err := c.OrgsForUser(userID)
	if err != nil {
		return false, err
	}

	fmt.Println("IsUserOrgMember:", userID, orgID)
	fmt.Println("orgs:", orgs, orgs[orgID])

	_, ok := orgs[orgID]
	return ok, nil
}

// IsUserOrgAdmin checks to see whether a user is an administrator of
// an organisation.
func (c *RESTClient) IsUserOrgAdmin(userID string, orgID string) (bool, error) {
	orgs, err := c.OrgsForUser(userID)
	if err != nil {
		return false, err
	}

	admin, ok := orgs[orgID]
	return ok && admin, nil
}

// GetDefaultPaymentMethod gets the user default payment method.
func (c *RESTClient) GetDefaultPaymentMethod(userID string) (*model.PaymentMethod, error) {
	url := fmt.Sprintf("%s/internal/user/%s/payment-method/default", c.baseURL, userID)

	// Do GET to endpoint.
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode == http.StatusOK {
		resp := model.PaymentMethod{}
		rspBody, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()
		if err = json.Unmarshal(rspBody, &resp); err != nil {
			return nil, err
		}
		return &resp, nil
	}

	return nil, chassis.BuildErrorFromErrMsg(rsp)
}

func (c *RESTClient) GetSSOSecret(orgIDorSlug string) (*string, error) {
	url := fmt.Sprintf("%s/internal/org/%s/sso-secret", c.baseURL, orgIDorSlug)

	// Do GET to endpoint.
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode == http.StatusOK {
		var secret string
		rspBody, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()
		if err = json.Unmarshal(rspBody, &secret); err != nil {
			return nil, err
		}
		return &secret, nil
	}

	return nil, chassis.BuildErrorFromErrMsg(rsp)
}

// GetCustomer gets the user customer reference.
func (c *RESTClient) GetCustomer(userID string) (*model.Customer, error) {
	url := fmt.Sprintf("%s/internal/user/%s/customer", c.baseURL, userID)

	// Do GET to endpoint.
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode == http.StatusOK {
		resp := model.Customer{}
		rspBody, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()
		if err = json.Unmarshal(rspBody, &resp); err != nil {
			return nil, err
		}
		return &resp, nil
	}

	return nil, chassis.BuildErrorFromErrMsg(rsp)
}

// GetPayoutAccount gets the user payout account of a given owner, that can be a user or org.
func (c *RESTClient) GetPayoutAccount(ownerId string) (*model.PayoutAccount, error) {
	url := fmt.Sprintf("%s/internal/payout-account/%s", c.baseURL, ownerId)

	// Do GET to endpoint.
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode == http.StatusOK {
		resp := model.PayoutAccount{}
		rspBody, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()
		if err = json.Unmarshal(rspBody, &resp); err != nil {
			return nil, err
		}
		return &resp, nil
	}
	return nil, chassis.BuildErrorFromErrMsg(rsp)
}

// GetAddress gets the an address of a certain user.
func (c *RESTClient) GetAddress(userID, addressId string) (*model.Address, error) {
	url := fmt.Sprintf("%s/internal/user/%s/address/%s", c.baseURL, userID, addressId)

	// Do GET to endpoint.
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode == http.StatusOK {
		resp := model.Address{}
		rspBody, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()
		if err = resp.UnrestrictedUnmarshalJSON(rspBody); err != nil {
			return nil, err
		}
		return &resp, nil
	}

	return nil, chassis.BuildErrorFromErrMsg(rsp)
}

// GetAddress gets the an address of a certain user.
func (c *RESTClient) GetDefaultAddress(userID string) (*model.Address, error) {
	url := fmt.Sprintf("%s/internal/user/%s/address/default", c.baseURL, userID)

	// Do GET to endpoint.
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode == http.StatusOK {
		resp := model.Address{}
		rspBody, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()
		if err = resp.UnrestrictedUnmarshalJSON(rspBody); err != nil {
			return nil, err
		}
		return &resp, nil
	}

	return nil, chassis.BuildErrorFromErrMsg(rsp)
}


//GetNotificationInfo returns the contact information for email of an user or organisation
func (c *RESTClient) GetNotificationInfo(contactId string) (*model.EmailNotificationInfo, error) {
	url := fmt.Sprintf("%s/internal/notification-info/%s", c.baseURL, contactId)

	// Do GET to endpoint.
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode == http.StatusOK {
		resp := model.EmailNotificationInfo{}
		rspBody, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()
		if err = json.Unmarshal(rspBody, &resp); err != nil {
			return nil, err
		}
		return &resp, nil
	}

	return nil, chassis.BuildErrorFromErrMsg(rsp)
}


//GetUserByApiKey returns an user if there is a match with the passed api_key
func (c *RESTClient) GetUserByApiKey(apiKey, apiSecret string) (*model.User, error) {
	url := fmt.Sprintf("%s/internal/api-key/%s?secret=%s", c.baseURL, apiKey, apiSecret)

	// Do GET to endpoint.
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode == http.StatusOK {
		resp := model.User{}
		rspBody, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()
		if err = json.Unmarshal(rspBody, &resp); err != nil {
			return nil, err
		}
		return &resp, nil
	}

	return nil, chassis.BuildErrorFromErrMsg(rsp)
}

//GetDeliveryFees
func (c *RESTClient) GetDeliveryFees(ids []string) (*map[string]model.DeliveryFees, error) {
	if len(ids) == 0 {
		return nil, errors.New("empty array of IDs")
	}
	url := fmt.Sprintf("%s/internal/delivery-fees?ids=%s", c.baseURL, strings.Join(ids, ","))

	// Do GET to endpoint.
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode == http.StatusOK {
		resp := map[string]model.DeliveryFees{}
		rspBody, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()
		if err = json.Unmarshal(rspBody, &resp); err != nil {
			return nil, err
		}
		return &resp, nil
	}

	return nil, chassis.BuildErrorFromErrMsg(rsp)
}