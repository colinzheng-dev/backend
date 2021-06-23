package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// RESTClient is a search service client that connects via REST.
type RESTClient struct {
	baseURL string
}

// New creates a new search service client.
func New(baseURL string) *RESTClient {
	return &RESTClient{
		baseURL: baseURL,
	}
}

// Geo performs a geolocation search.
func (c *RESTClient) Geo(latitude, longitude, dist float64) ([]string, error) {
	return c.idList(fmt.Sprintf("%s/search/geo?geo=%f,%f&dist=%f",
		c.baseURL, latitude, longitude, dist))
}

// FullText performs a full-text search.
func (c *RESTClient) FullText(q string) ([]string, error) {
	return c.idList(fmt.Sprintf("%s/search/full_text?q=%s", c.baseURL, url.QueryEscape(q)))
}

// Region search items inside a region.
func (c *RESTClient) Region(regionRef, regionType string) ([]string, error) {
	requestUrl := fmt.Sprintf("%s/search/region?region_type=%s&region=%s", c.baseURL, regionType, regionRef)
	return c.idList(requestUrl)
}

// CheckRegions checks if a certain point is inside a list of regions.
func (c *RESTClient) CheckRegions(latitude, longitude float64, regions []int) (*bool, error) {
	var refs []string
	for _, ref := range regions {
		refs = append(refs, strconv.Itoa(ref))
	}

	requestUrl := fmt.Sprintf("%s/search/check-region?location=%f,%f&regions=%s",
		c.baseURL, latitude, longitude, strings.Join(refs, ",") )

	rsp, err := http.Get(requestUrl)
	if err != nil {
		return nil, err
	}

	// Decode response.
	var resp bool
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

func (c *RESTClient) idList(url string) ([]string, error) {
	// GET request.
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	// Decode response.
	ids := []string{}
	rspBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	err = json.Unmarshal(rspBody, &ids)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

