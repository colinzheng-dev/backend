package client

import (
	"encoding/json"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/social-service/model"
	"io/ioutil"
	"net/http"
)

// RESTClient is a social service client that connects via REST.
type RESTClient struct {
	baseURL string
}

// New creates a new item service client.
func New(baseURL string) *RESTClient {
	return &RESTClient{baseURL}
}

// GetUpvotesCount invokes the item GET /internal/items/upvotes method of the social service.
func (c *RESTClient) GetUpvotesCount() (*map[string]model.UpvoteQuantityInfo, error) {
	// Do GET to endpoint.

	url := c.baseURL + "/internal/upvotes"

	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	// Decode response.
	if rsp.StatusCode == 200 {
		resp := map[string]model.UpvoteQuantityInfo{}
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


// GetUpvotesCount invokes the item GET /internal/items/upvotes method of the social service.
func (c *RESTClient) GetUserUpvotes(sessionUser string) (*map[string]bool, error) {

	url := c.baseURL + "/internal/user/" + sessionUser + "/upvotes"
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	// Decode response.
	if rsp.StatusCode == 200 {
		//result := make(map[string]int)
		resp := map[string]bool{}
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
// GetUpvotesCount invokes the item GET /internal/item/{item_iod}/rank method of the social service.
func (c *RESTClient) GetOverallRank(itemId string) (*float64, error) {
	// Do GET to endpoint.
	url := c.baseURL + "/internal/item/" + itemId + "/rank"
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	// Decode response.
	if rsp.StatusCode == 200 {
		var resp float64
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