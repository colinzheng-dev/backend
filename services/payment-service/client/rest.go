package client

import (
	"bytes"
	"encoding/json"
	"github.com/veganbase/backend/services/payment-service/model"
	purchase "github.com/veganbase/backend/services/purchase-service/model"

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

// CreatePaymentIntent invokes the payment internal path POST /internal/payment-intent to
// create a new payment. This method will call stripe's API to create a new intent and return
// it's summary information a PaymentIntentView model.
func (c *RESTClient) CreatePaymentIntent(purchase purchase.Purchase) (*[]model.PaymentIntent, error) {
	jsonBytes, err := json.Marshal(purchase)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/internal/payment-intent", bytes.NewBuffer(jsonBytes))

	client := &http.Client{}
	rsp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	// Decode response.
	pi := []model.PaymentIntent{}
	rspBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	if err = json.Unmarshal(rspBody, &pi); err != nil {
		return nil, err
	}

	return &pi, nil
}


