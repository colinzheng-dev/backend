package client

import (
	"bytes"
	"encoding/json"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/cart-service/model"
	"io/ioutil"
	"net/http"
)

// RESTClient is a cart service client that connects via REST.
type RESTClient struct {
	baseURL string
}

// New creates a new item service client.
func New(baseURL string) *RESTClient {
	return &RESTClient{baseURL}
}

// Active invokes the cart GET /cart/{user_id}/active method on the cart service.
func (c *RESTClient) Active(userId string) (*model.FullCart, error) {
	// Do GET to endpoint.
	rsp, err := http.Get(c.baseURL + "/internal/cart/" + userId + "/active")
	if err != nil {
		return nil, err
	}

	// Decode response.
	if rsp.StatusCode == http.StatusOK {
		resp := model.FullCart{}
		rspBody, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()

		if 	err = resp.UnmarshalJSON(rspBody); err != nil {
			return nil, err
		}

		return &resp, nil
	}
	return nil, chassis.BuildErrorFromErrMsg(rsp)
}

// FinishCart invokes the cart-service path PATCH /cart/{cart_id}
func (c *RESTClient) FinishCart(cartId string) (*model.Cart, error) {

	// Do PATCH to endpoint.
	jsonBytes := []byte(`{"cart_status":"complete"}`)
	req, err := http.NewRequest(http.MethodPatch, c.baseURL + "/internal/cart/"+ cartId, bytes.NewBuffer(jsonBytes))

	client := &http.Client{}
	rsp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	// Decode response.
	if rsp.StatusCode == http.StatusOK {
		cart := model.Cart{}
		rspBody, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()
		if err = json.Unmarshal(rspBody, &cart); err != nil {
			return nil, err
		}

		return &cart, nil
	}
	return nil, chassis.BuildErrorFromErrMsg(rsp)
}