package server

import (
	item "github.com/veganbase/backend/services/item-service/client"
)

func (s *Server) addLocationInfo(id string, info *item.SearchInfo) error {
	return s.db.AddGeo(id, info.ItemType, info.Approval,
		info.Location.Latitude, info.Location.Longitude)
}

func (s *Server) addFullTextInfo(id string, info *item.SearchInfo) error {
	return s.db.AddFullText(id, info.ItemType, info.Approval,
		info.Name, info.Description, info.Content, info.Tags)
}
