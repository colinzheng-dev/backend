package client

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/veganbase/backend/chassis/pubsub"
	"github.com/veganbase/backend/services/category-service/events"
	category_events "github.com/veganbase/backend/services/category-service/events"
)

// RESTClient is a category service client that connects via REST.
type RESTClient struct {
	baseURL      string
	muCategories sync.RWMutex
	categories   events.CategoryMap
	membership   map[string]map[string]bool
	pubsub       pubsub.PubSub
	subName      string
}

// New creates a new category service client.
func New(baseURL string, pubsub pubsub.PubSub, subname string) *RESTClient {
	client := &RESTClient{
		baseURL:    baseURL,
		categories: events.CategoryMap{},
		membership: map[string]map[string]bool{},
		pubsub:     pubsub,
		subName:    subname,
	}
	populate(client)
	return client
}

// Categories returns the category list (which is populated and
// updated asynchronously based on events from the category service).
func (c *RESTClient) Categories() events.CategoryMap {
	c.muCategories.RLock()
	defer c.muCategories.RUnlock()
	return c.categories
}

// IsValidLabel determines whether a label is valid for a category.
func (c *RESTClient) IsValidLabel(catName string, label string) bool {
	vals, ok := c.membership[catName]
	if !ok {
		return false
	}
	return vals[label]
}

func populate(c *RESTClient) {
	// Populate the category map by doing a REST call.
	populateFromREST(c)

	// Now wait for update events.
	if c.pubsub != nil {
		go handleUpdates(c)
	}
}

func populateFromREST(c *RESTClient) {
	first := true
outer:
	for {
		if !first {
			time.Sleep(10 * time.Second)
		}
		first = false

		// GET category list from service endpoint.
		url := c.baseURL + "/categories"
		rsp, err := http.Get(url)
		if err != nil {
			log.Error().Err(err).Str("url", url).
				Msg("couldn't retrieve category list")
			continue
		}

		// Decode response.
		cats := map[string]interface{}{}
		rspBody, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			log.Error().Err(err).
				Msg("couldn't read category list response body")
			continue
		}
		defer rsp.Body.Close()
		err = json.Unmarshal(rspBody, &cats)
		if err != nil {
			log.Error().Err(err).Str("body", string(rspBody)).
				Msg("couldn't decode category list response body")
			continue
		}

		// Now get the contents for each category.
		categories := map[string]*events.Category{}
		for name := range cats {
			cat, err := c.getOneCategory(name)
			if err != nil {
				log.Error().Err(err).Str("name", name).
					Msg("failed to get category information")
				continue outer
			}
			categories[name] = cat
		}

		// Set category map.
		c.setCategories(categories)
		return
	}
}

func (c *RESTClient) setCategories(categories events.CategoryMap) {
	mem := map[string]map[string]bool{}
	for n, cat := range categories {
		mem[n] = map[string]bool{}
		for label := range *cat {
			mem[n][label] = true
		}
	}
	c.muCategories.Lock()
	defer c.muCategories.Unlock()
	c.categories = categories
	c.membership = mem
}

func (c *RESTClient) setCategory(name string, entries events.Category) {
	mem := map[string]bool{}
	for label := range entries {
		mem[label] = true
	}
	c.muCategories.Lock()
	defer c.muCategories.Unlock()
	c.categories[name] = &entries
	c.membership[name] = mem
}

func (c *RESTClient) getOneCategory(name string) (*events.Category, error) {
	url := c.baseURL + "/category/" + name
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	// Decode response.
	cat := events.Category{}
	rspBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	err = json.Unmarshal(rspBody, &cat)
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func handleUpdates(c *RESTClient) {
	subCh, _, err := c.pubsub.Subscribe(category_events.CategoryUpdate, c.subName, pubsub.Fanout)
	if err != nil {
		log.Fatal().Err(err).
			Msg("unable to subscribe to category updates")
	}

	for {
		catInfoData := <-subCh
		catInfo := events.CategoryUpdateInfo{}
		err := json.Unmarshal(catInfoData, &catInfo)
		if err != nil {
			log.Error().Err(err).
				Msg("unmarshalling category list update")
			continue
		}
		c.setCategory(catInfo.Name, catInfo.Entries)
	}
}
