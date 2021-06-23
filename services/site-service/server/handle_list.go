package server

import "net/http"

// Return the site list.
func (s *Server) list(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.sites, nil
}
