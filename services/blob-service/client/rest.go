package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// RESTClient is a blob service client that connects via REST.
type RESTClient struct {
	baseURL string
}

// New creates a new blob service client.
func New(baseURL string) *RESTClient {
	return &RESTClient{baseURL}
}

// AddItemBlobs invokes the "add blob-item associations" method on the
// blob service.
func (c *RESTClient) AddItemBlobs(itemID string, blobIDs []string) error {
	// Do POST to endpoint.
	endpoint := fmt.Sprintf(c.baseURL+"/blob-item-assoc/%s", itemID)
	body, err := json.Marshal(blobIDs)
	rsp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	if rsp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code '%s' from blob service", rsp.Status)
	}
	return nil
}

// RemoveItemBlobs invokes the "remove blob-item association" method
// on the blob service. (An empty blob ID list means to remove all
// item/blob associations for the item.)
func (c *RESTClient) RemoveItemBlobs(itemID string, blobIDs []string) error {
	// Do DELETE to endpoint.
	endpoint := fmt.Sprintf(c.baseURL+"/blob-item-assoc/%s", itemID)
	body, err := json.Marshal(blobIDs)
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	rsp, err := client.Do(req)
	if rsp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code '%s' from blob service", rsp.Status)
	}
	return err
}
