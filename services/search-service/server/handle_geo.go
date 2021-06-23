package server

import (
	"errors"
	"github.com/go-chi/chi"
	"github.com/veganbase/backend/chassis"
	itModel "github.com/veganbase/backend/services/item-service/model"
	itTypes "github.com/veganbase/backend/services/item-service/model/types"
	itUtils "github.com/veganbase/backend/services/item-service/utils"
	"net/http"
	"net/url"
)

func (s *Server) geo(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return nil, errors.New("invalid query parameters")
	}

	var geo *[2]float64
	var dist *float64
	if err := chassis.GeoParam(qs, "geo", &geo); err != nil {
		return chassis.BadRequest(w, "invalid geo parameter")
	}
	if err := chassis.FloatParam(qs, "dist", &dist); err != nil {
		return chassis.BadRequest(w, "invalid dist parameter")
	}
	if dist == nil || geo == nil {
		return chassis.BadRequest(w, "must set both geo and dist parameters")
	}
	var itemType *itModel.ItemType
	if err := itUtils.ItemTypeParam(qs, &itemType); err != nil {
		return chassis.BadRequest(w, "invalid type parameter")
	}
	var approval *[]itTypes.ApprovalState
	if err := itUtils.ApprovalParam(qs, &approval); err != nil {
		return chassis.BadRequest(w, "invalid approval parameter")
	}

	res, err := s.db.Geo(geo[0], geo[1], *dist, itemType, approval)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) checkRegion(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	var regions *[]int
	var location *[2]float64

	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return nil, errors.New("invalid query parameters")
	}

	chassis.GeoParam(qs, "location", &location)
	chassis.IntSliceParam( qs, "regions", &regions)

	res, err := s.db.IsInsideRegions(location[0], location[1], *regions)
	if err != nil {
		return nil, err
	}

	return res, nil
}



func (s *Server) region(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	var regionType *string
	var region *string

	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return nil, errors.New("invalid query parameters")
	}

	if err := GetRegionType(qs, &regionType) ; err!= nil {
		return chassis.BadRequest(w, "invalid region parameter")
	}

	var itemType *itModel.ItemType
	if err := itUtils.ItemTypeParam(qs, &itemType); err != nil {
		return chassis.BadRequest(w, "invalid type parameter")
	}
	var approval *[]itTypes.ApprovalState
	if err := itUtils.ApprovalParam(qs, &approval); err != nil {
		return chassis.BadRequest(w, "invalid approval parameter")
	}

	chassis.StringParam(qs, "region", &region)

	res, err := s.db.GetItemsInsideRegion(*regionType, *region, itemType, approval)
	if err != nil {
		return nil, err
	}

	return res, nil
}



func (s *Server) Countries(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	res, err := s.db.Countries()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Server) States(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	countryID := chi.URLParam(r, "country_id")

	if countryID == "" {
		return chassis.BadRequest(w, "country ID is missing")
	}
	res, err := s.db.States(countryID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

