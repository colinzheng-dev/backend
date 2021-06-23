package server

import (
	"strings"

	"github.com/rs/zerolog/log"
)

var sites = []string{
	"https://veganbase.com",
	"https://veganapi.com",
	"https://ethicalbuzz.com",
	"https://vegansunited.co",
}

// MaintainSiteInfo uses information from the site service to maintain:
//
// 1. the list of allowed origins for CORS checking, based on the
//    sites configured in the site service, plus a list of fixed
//    origins from the API gateway configuration;
//
// 2. a mapping from site URL to site name, used for determining the
//    site for a request from the request Origin header.
func (s *Server) MaintainSiteInfo() {
	s.updateCORS()
	s.updateSiteURLs()
	ch, _, err := s.siteSvc.SiteUpdates()
	if err != nil {
		log.Fatal().Err(err).
			Msg("can't start site update watcher")
	}
	for {
		<-ch
		s.updateCORS()
		s.updateSiteURLs()
	}
}

func (s *Server) updateCORS() {
	newCORSOrigins := map[string]bool{}
	msg := []string{}

	// Fixed origin list from gateway configuration.
	for _, origin := range s.configCORSOrigins {
		newCORSOrigins[origin] = true
		msg = append(msg, origin)
	}

	// Origins from site service.
	sites := s.siteSvc.Sites()
	for _, origin := range sites {
		newCORSOrigins[origin.URL] = true
		msg = append(msg, origin.URL)
	}

	s.muCORSOrigins.Lock()
	defer s.muCORSOrigins.Unlock()
	s.corsOrigins = newCORSOrigins
	log.Info().
		Str("origins", strings.Join(msg, " ")).
		Msg("allowed CORS origins updated")
}

// CORSOrigins gets the current allowed origins for CORS checking.
func (s *Server) CORSOrigins() map[string]bool {
	s.muCORSOrigins.RLock()
	defer s.muCORSOrigins.RUnlock()
	return s.corsOrigins
}

func (s *Server) updateSiteURLs() {
	newSiteURLs := map[string]string{}
	for name, site := range s.siteSvc.Sites() {
		newSiteURLs[site.URL] = name
	}

	s.muSiteURLs.Lock()
	defer s.muSiteURLs.Unlock()
	s.siteURLs = newSiteURLs
}

// SiteURLs gets the mapping from site URLs to site names.
func (s *Server) SiteURLs() map[string]string {
	s.muSiteURLs.RLock()
	defer s.muSiteURLs.RUnlock()
	return s.siteURLs
}
