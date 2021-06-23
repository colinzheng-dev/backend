package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/item-service/client"
	"github.com/veganbase/backend/services/item-service/db"
)

func (s *Server) ids(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	ids, err := s.db.ItemIDs()
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (s *Server) searchInfo(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	id := chi.URLParam(r, "id")
	if id == "" {
		return chassis.BadRequest(w, "missing item id")
	}

	item, err := s.db.ItemByID(id)
	if err != nil {
		if err == db.ErrItemNotFound {
			return chassis.NotFoundWithMessage(w, err.Error())
		}
		return nil, err
	}

	resp := client.SearchInfo{
		Name:        item.Name,
		ItemType:    item.ItemType,
		Approval:    item.Approval,
		Description: item.Description,
		Tags:        item.Tags,
	}
	if c, ok := item.Attrs["content"]; ok {
		content, ok := c.(string)
		if ok {
			resp.Content = content
		}
	}
	if c, ok := item.Attrs["location"]; ok {
		loc, ok := c.(map[string]interface{})
		if ok {
			lat, ok1 := loc["latitude"]
			lon, ok2 := loc["longitude"]
			if ok1 && ok2 {
				latitude, ok1 := lat.(float64)
				longitude, ok2 := lon.(float64)
				if ok1 && ok2 {
					location := client.Location{
						Latitude:  latitude,
						Longitude: longitude,
					}
					resp.Location = &location
				}
			}
		}
	}

	return &resp, nil
}
