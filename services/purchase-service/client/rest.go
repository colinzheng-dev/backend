package client

import (
	"bytes"
	"encoding/json"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/purchase-service/model"
	"github.com/veganbase/backend/services/purchase-service/model/types"
	"io/ioutil"
	"net/http"
)

// RESTClient is a purchase service client that connects via REST.
type RESTClient struct {
	baseURL string
}

// New creates a new item service client.
func New(baseURL string) *RESTClient {
	return &RESTClient{baseURL}
}

// GetPurchaseInfo retrieves a purchase in it's full format by calling /internal/purchase/{pur_id}?format=full
func (c *RESTClient) GetPurchaseInfo(purchaseId string) (*model.FullPurchase, error) {
	// Do GET to endpoint.
	rsp, err := http.Get(c.baseURL + "/internal/purchase/" + purchaseId + "?format=full")
	if err != nil {
		return nil, err
	}

	// Decode response.
	if rsp.StatusCode == http.StatusOK {
		resp := model.FullPurchase{}
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
	return nil, chassis.BuildErrorFromErrMsg(rsp)
	}

// UpdatePurchaseStatus invokes the purchase-service internal path PATCH /internal/purchase/{pur_id} to
// update the purchase status with the desired value
func (c *RESTClient) UpdatePurchaseStatus(purchaseId string, status string) (*model.Purchase, error) {

	// Do PATCH to endpoint.
	var purchaseStatus types.PurchaseStatus
	//check if status passed is valid
	if err := purchaseStatus.FromString(status);err != nil {
		return nil, err
	}

	jsonBytes := []byte(`{"status":"` + purchaseStatus.String() + `"}`)
	req, err := http.NewRequest(http.MethodPatch, c.baseURL+"/internal/purchase/"+purchaseId, bytes.NewBuffer(jsonBytes))

	client := &http.Client{}
	rsp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	if rsp.StatusCode == http.StatusOK {
		// Decode response.
		purchase := model.Purchase{}
		rspBody, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()
		if err = json.Unmarshal(rspBody, &purchase); err != nil {
			return nil, err
		}

		return &purchase, nil
	}
	return nil, chassis.BuildErrorFromErrMsg(rsp)
}

// UpdateOrderStatus invokes the purchase-service internal path PATCH /internal/order/{ord_id} to
// update the order payment status with the desired value
func (c *RESTClient) UpdateOrderPaymentStatus(orderId string, status string) (*model.Order, error) {

	// Do PATCH to endpoint.
	var paymentStatus types.PaymentStatus
	//check if status passed is valid
	if err := paymentStatus.FromString(status); err != nil {
		return nil, err
	}

	jsonBytes := []byte(`{"payment_status":"` + paymentStatus.String() + `"}`)
	req, err := http.NewRequest(http.MethodPatch, c.baseURL+"/internal/order/"+orderId, bytes.NewBuffer(jsonBytes))

	client := &http.Client{}
	rsp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	if rsp.StatusCode == http.StatusOK {
		// Decode response.
		order := model.Order{}
		rspBody, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()
		if err = json.Unmarshal(rspBody, &order); err != nil {
			return nil, err
		}

		return &order, nil
	}
	return nil, chassis.BuildErrorFromErrMsg(rsp)
}


// UpdateBookingStatus invokes the purchase-service internal path PATCH /internal/booking/{bok_id} to
// update the booking payment status with the desired value
func (c *RESTClient) UpdateBookingPaymentStatus(bookingId string, status string) (*model.Booking, error) {

	// Do PATCH to endpoint.
	var paymentStatus types.PaymentStatus
	//check if status passed is valid
	if err := paymentStatus.FromString(status);err != nil {
		return nil, err
	}

	jsonBytes := []byte(`{"payment_status":"` + paymentStatus.String() + `"}`)
	req, err := http.NewRequest(http.MethodPatch, c.baseURL+"/internal/booking/"+bookingId, bytes.NewBuffer(jsonBytes))

	client := &http.Client{}
	rsp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	if rsp.StatusCode == http.StatusOK {
		// Decode response.
		bk := model.Booking{}
		rspBody, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()
		if err = json.Unmarshal(rspBody, &bk); err != nil {
			return nil, err
		}

		return &bk, nil
	}
	return nil, chassis.BuildErrorFromErrMsg(rsp)
}

func (c *RESTClient) UserBoughtItem(itemId, userId string) (*bool, error) {
	rsp, err := http.Get(c.baseURL + "/internal/purchase/item-bought?item_id=" +itemId + "&user_id=" + userId)
	if err != nil {
		return nil, err
	}

	// Decode response.
	if rsp.StatusCode == http.StatusOK {
		var resp bool
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
	return nil, chassis.BuildErrorFromErrMsg(rsp)
}