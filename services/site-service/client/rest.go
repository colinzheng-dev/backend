package client

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/veganbase/backend/chassis/pubsub"
	site_events "github.com/veganbase/backend/services/site-service/events"
	"github.com/veganbase/backend/services/site-service/model"
)

// RESTClient is a site service client that connects via REST.
type RESTClient struct {
	baseURL       string
	muSites       sync.RWMutex
	sites         SiteMap
	pubsub        pubsub.PubSub
	subName       string
	listenerCount int
	muListeners   sync.RWMutex
	listeners     map[int]chan bool
}

// New creates a new site service client.
func New(baseURL string, pubsub pubsub.PubSub, subname string) *RESTClient {
	client := &RESTClient{
		baseURL:       baseURL,
		sites:         SiteMap{},
		pubsub:        pubsub,
		subName:       subname,
		listenerCount: 1,
		listeners:     map[int]chan bool{},
	}
	go populate(client)
	return client
}

func populate(c *RESTClient) {
	// First populate the site map by doing a REST call.
	populateFromREST(c)

	// Now wait for update events.
	if c.pubsub != nil {
		handleUpdates(c)
	}
}

func populateFromREST(c *RESTClient) {
	first := true
	for {
		if !first {
			time.Sleep(10 * time.Second)
		}
		first = false

		// GET site map from site service endpoint.
		url := c.baseURL + "/sites"
		rsp, err := http.Get(url)
		if err != nil {
			log.Error().Err(err).Str("url", url).
				Msg("couldn't retrieve sites map")
			continue
		}

		// Decode response.
		sites := map[string]*model.Site{}
		rspBody, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			log.Error().Err(err).
				Msg("couldn't read sites map response body")
			continue
		}
		defer rsp.Body.Close()
		err = json.Unmarshal(rspBody, &sites)
		if err != nil {
			log.Error().Err(err).Str("body", string(rspBody)).
				Msg("couldn't decode sites map response body")
			continue
		}

		// Set site map.
		c.setSites(sites)
		c.notifyListeners()
		return
	}
}

func (c *RESTClient) setSites(sites SiteMap) {
	c.muSites.Lock()
	defer c.muSites.Unlock()
	c.sites = sites
}

func (c *RESTClient) notifyListeners() {
	c.muListeners.RLock()
	defer c.muListeners.RUnlock()
	for _, listener := range c.listeners {
		// Non-blocking send: if the channel is full, we've notified the
		// listener before but they've not read from the channel yet.
		select {
		case listener <- true:
		default:
		}
	}
}

func handleUpdates(c *RESTClient) {
	subCh, _, err := c.pubsub.Subscribe(site_events.SiteUpdate, c.subName, pubsub.Fanout)
	if err != nil {
		log.Fatal().Err(err).
			Msg("unable to subscribe to site updates")
	}

	for {
		newSiteData := <-subCh
		newSites := SiteMap{}
		err := json.Unmarshal(newSiteData, &newSites)
		if err != nil {
			log.Error().Err(err).
				Msg("unmarshalling site list update")
			continue
		}
		c.setSites(newSites)
		c.notifyListeners()
	}
}

// Sites returns the site list (which is populated and updated
// asynchronously based on events from the site service).
func (c *RESTClient) Sites() SiteMap {
	c.muSites.RLock()
	defer c.muSites.RUnlock()
	return c.sites
}

// SiteUpdates returns a channel that sends a value whenever the site
// list is updated, along with a cancel function to stop watching for
// updates.
func (c *RESTClient) SiteUpdates() (chan bool, func(), error) {
	if c.pubsub == nil {
		return nil, nil, ErrNoPubSub
	}

	c.muListeners.Lock()
	defer c.muListeners.Unlock()

	id := c.listenerCount
	c.listenerCount++

	ch := make(chan bool)
	c.listeners[id] = ch
	fn := func() { delete(c.listeners, id) }

	return ch, fn, nil
}
