package server

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

// Sync performs the initial synchronisation of the search database
// with the items database.
func (s *Server) Sync() {
	fmt.Println("Synchronising with item service database...")

	var ids []string
	var err error
	for  {
		for {
			ids, err = s.itemSvc.IDs()
			if err == nil {
				break
			}

			s.LogError("item-sync", "couldn't get ID list from item service during sync", err, true)
			log.Info().Msg("trying to acquire item ID list again in 5 minutes")
			time.Sleep(5 * time.Minute)
		}

		for _, id := range ids {
			log.Info().Str("id", id).Msg("syncing with item service")
			searchInfo, err := s.itemSvc.SearchInfo(id)
			if err != nil {
				s.LogError("item-sync", "getting information from item service for ID " + id, err, true)
				continue
			}

			if searchInfo.Location != nil {
				s.addLocationInfo(id, searchInfo)
			}

			if err := s.addFullTextInfo(id, searchInfo); err != nil {
				s.LogError("item-sync", "adding text information for ID " + id, err, true)
				continue
			}
		}
		wait()
	}
}

//TODO: MAKE IT A SCHEDULED ROUTINE TO EXECUTE BASED ON PRE-SET TIME
func wait() {
	period := time.Tick(5* time.Minute)
	for range period {
		break
	}
}
func (s *Server) processItemUpdate(id string) {
	searchInfo, err := s.itemSvc.SearchInfo(id)
	if err != nil {
		//TODO: MAY BE A GOOD IDEA TO LOG ERRORS ON DATABASE
		log.Error().Err(err).
			Msgf("getting information from item service for ID '%s'", id)
		return
	}

	if searchInfo.Location != nil {
		s.addLocationInfo(id, searchInfo)
	}
	s.addFullTextInfo(id, searchInfo)
}

func (s *Server) processItemDelete(id string) {
	s.db.ItemRemoved(id)
}
