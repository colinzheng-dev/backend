package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/item-service/model"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

// RESTClient is a item service client that connects via REST.
type RESTClient struct {
	baseURL string
}

// New creates a new item service client.
func New(baseURL string) *RESTClient {
	return &RESTClient{baseURL}
}

// IDs invokes the user ids method on the item service.
func (c *RESTClient) IDs() ([]string, error) {
	// Do GET to endpoint.
	rsp, err := http.Get(c.baseURL + "/ids")
	if err != nil {
		return nil, err
	}

	// Decode response.
	resp := []string{}
	rspBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	err = json.Unmarshal(rspBody, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// SearchInfo invokes the user search_info method on the item service.
func (c *RESTClient) SearchInfo(id string) (*SearchInfo, error) {
	// Do GET to endpoint.
	rsp, err := http.Get(c.baseURL + "/search_info/" + id)
	if err != nil {
		return nil, err
	}

	// Decode response.
	resp := SearchInfo{}
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

// ItemInfo invokes the item GET/item/{item_id} method of the item service.
func (c *RESTClient) ItemInfo(id string) (*model.Item, error) {
	// Do GET to endpoint.
	rsp, err := http.Get(c.baseURL + "/item/" + id)
	if err != nil {
		return nil, err
	}

	// Decode response.
	if rsp.StatusCode == 200 {
		resp := model.Item{}

		rspBody, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()

		err = resp.UnrestrictedUnmarshalJSON(rspBody)
		if err != nil {
			return nil, err
		}

		return &resp, nil
	}
	return nil, chassis.BuildErrorFromErrMsg(rsp)
}

// ItemFullWithLinks invokes the item GET/item/{item_id}?links={linkType}:full method of the item service.
func (c *RESTClient) ItemFullWithLink(id, linkType string) (*model.ItemFullWithLink, error) {
	// Do GET to endpoint.
	url := c.baseURL + "/item/" + id
	if linkType != "" {
		url += "?links=" + linkType + ":full"
	}

	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	// Decode response.
	if rsp.StatusCode == 200 {
		resp := model.ItemFullWithLink{}

		rspBody, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()

		if err = resp.UnmarshalJSON(rspBody); err != nil {
			return nil, err
		}

		return &resp, nil
	}
	return nil, chassis.BuildErrorFromErrMsg(rsp)
}

// GetItems invokes the item GET/item/{item_id} method of the item service.
func (c *RESTClient) GetItems(ids []string, linkType string) (*[]model.ItemFullWithLink, error) {
	// Do GET to endpoint.
	url := c.baseURL + "/items?format=full&ids=" + strings.Join(ids, ",")
	if linkType != "" {
		url += "?links=" + linkType + ":full"
	}
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	// Decode response.
	if rsp.StatusCode == 200 {
		var resp []model.ItemFullWithLink
		//todo: with this unmarshal to interface{}, the attributes are alphabetically sorted
		//      maybe it is better to do an ItemFull.Unmarshal to be consistent
		rspBody, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return nil, err
		}
		defer rsp.Body.Close()

		//err = json.Unmarshal(rspBody, &resp)
		err = json.Unmarshal(rspBody, &resp)
		if err != nil {
			return nil, err
		}

		return &resp, nil
	}
	return nil, errors.New("item not found")
}

// Info invokes the item info method on the item service.
func (c *RESTClient) GetItemsInfo(ids []string) (map[string]*model.Info, error) {

	if len(ids) == 0 {
		return map[string]*model.Info{}, nil
	}

	// Do GET to endpoint.
	rsp, err := http.Get(c.baseURL + "/internal/info?ids=" + strings.Join(ids, ","))
	if err != nil {
		return nil, err
	}

	// Decode response.
	resp := map[string]*model.Info{}
	rspBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	if err = json.Unmarshal(rspBody, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *RESTClient) UpdateItemAvailability(itemId string, quantity int) error {

	// Do PATCH to endpoint.
	jsonBytes := []byte(`{"quantity":` + strconv.Itoa(quantity) + `}`)
	req, err := http.NewRequest(http.MethodPatch, c.baseURL+"/internal/item/"+itemId, bytes.NewBuffer(jsonBytes))

	client := &http.Client{}
	rsp, err := client.Do(req)

	if err != nil {
		return err
	}
	if rsp.StatusCode == http.StatusOK {
		return nil
	}
	return errors.New("update item availability failed")

}
