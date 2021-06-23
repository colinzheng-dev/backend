package server

import (
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/item-service/model"
	"net/http"
	"net/url"
	"strings"
)

func (s *Server) getItemTypeSummaryInfo(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	info, err := s.db.GetItemTypeInfo()
	if err != nil {
		return nil, err
	}
	return parseResponse(info), nil
}

func parseResponse (info *[]model.ItemTypeInfo) map[string]interface{} {
	responseMap := make(map[string]interface{})

	for _, i := range *info {
		responseMap[i.ItemType] = model.ItemTypeInfo{
			Quantity: i.Quantity,
		}
	}

	return responseMap
}

func (s *Server) getItemsBasicInfo(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return chassis.BadRequest(w, "invalid query parameters")
	}

	idparam := qs.Get("ids")
	if idparam == "" {
		return chassis.BadRequest(w, "invalid items ID list")
	}

	ids := strings.Split(idparam, ",")
	info, err := s.db.Info(ids)
	if err != nil {
		return nil, err
	}

	return info, nil
}
