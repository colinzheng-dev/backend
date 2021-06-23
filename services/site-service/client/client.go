package client

import (
	"errors"

	"github.com/veganbase/backend/services/site-service/model"
)

// ErrNoPubSub is the error returned when an attempt is made to watch
// for site updates without pub/sub being set up.
var ErrNoPubSub = errors.New("pub/sub needs to be set up to watch for updates")

// SiteMap is a map from site names to site information.
type SiteMap map[string]*model.Site

// Client is the service client API for the site service.
type Client interface {
	// Sites returns the current site list (which is populated and
	// updated asynchronously based on events from the site service).
	Sites() SiteMap

	// SiteUpdates returns a channel that sends a value whenever the
	// site list is updated, along with a cancel function to stop
	// watching for updates.
	SiteUpdates() (chan bool, func(), error)
}
