package server

import (
	"time"

	"github.com/rs/zerolog/log"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/site-service/events"
	"github.com/veganbase/backend/services/site-service/model"
)

// UpdateSites is a function that runs in a goroutine and regularly
// updates the sites list from the database.
func (s *Server) UpdateSites() {
	for range time.Tick(30 * time.Second) {
		sitesTmp, err := s.db.Sites()
		if err != nil {
			log.Error().Err(err).Msg("couldn't read sites from database")
			continue
		}
		if sitesEqual(s.sites, sitesTmp) {
			continue
		}

		s.setSites(sitesTmp)
		err = chassis.Emit(s, events.SiteUpdate, s.sites)
		if err != nil {
			log.Error().Err(err).Msg("failed emitting site update event")
		}
	}
}

// Set current site list (in its own function for locking purposes).
func (s *Server) setSites(ss map[string]*model.Site) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.sites = ss
}

// Determine whether two site maps are equal.
func sitesEqual(sites1, sites2 map[string]*model.Site) bool {
	if len(sites1) != len(sites2) {
		return false
	}
	for k, s1 := range sites1 {
		s2, ok := sites2[k]
		if !ok {
			return false
		}
		if s1.ID != s2.ID || s2.Name != s2.Name ||
			s1.URL != s2.URL || s1.EmailDomain != s2.EmailDomain ||
			s1.Signature != s2.Signature {
			return false
		}
	}
	return true
}
