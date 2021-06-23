package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/item-service/model"
	"github.com/veganbase/backend/services/item-service/model/types"
	"github.com/veganbase/backend/services/item-service/utils"
)

// ErrSearchNotFound is the error returned when an attempt is made to
// access a search route that isn't allowed and should turn up as a
// 404 Not Found.
var ErrSearchNotFound = errors.New("search route not found")

// ResultFormat is a wrapper around a string describing the result
// format from a search, either full results or summary results.
type ResultFormat string

// Constant values for results formats.
const (
	SummaryResults ResultFormat = "summary"
	FullResults    ResultFormat = "full"
)

// SearchParams represents all the possible query parameters on search
// routes.
type SearchParams struct {
	ItemTypes   []model.ItemType
	Q           *string
	Format      ResultFormat
	Geo         *[2]float64
	Dist        *float64
	Pagination  *chassis.Pagination
	Approval    *[]types.ApprovalState
	Owner       *string
	Links       LinkInfo
	Ids         *[]string
	Collections *[]string
	SortBy      *chassis.Sorting
}

// Params processes all the possible query parameters for a search
// route.
func Params(r *http.Request, allowOwner bool) (*SearchParams, error) {
	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return nil, errors.New("invalid query parameters")
	}
	ps := SearchParams{}

	// Pagination parameters.
	ps.Pagination =	&chassis.Pagination{}
	if err := chassis.PaginationParams(qs, &ps.Pagination.Page, &ps.Pagination.PerPage); err != nil {
		return nil, err
	}

	// Format parameter.
	switch qs.Get("format") {
	case "full":
		ps.Format = FullResults
	case "summary", "":
		ps.Format = SummaryResults
	default:
		return nil, errors.New("invalid format parameter")
	}

	// Text search parameter.
	chassis.StringParam(qs, "q", &ps.Q)

	// Geo-search parameters.
	if err := chassis.FloatParam(qs, "dist", &ps.Dist); err != nil {
		return nil, err
	}
	if err := chassis.GeoParam(qs, "geo", &ps.Geo); err != nil {
		return nil, err
	}
	if (ps.Dist == nil) != (ps.Geo == nil) {
		return nil, errors.New("must set both geo and dist parameters")
	}

	// Filtering parameters.
	ps.ItemTypes, err = utils.ItemsTypeParam(qs)
	if err != nil {
		return nil, err
	}
	if err = utils.ApprovalParam(qs, &ps.Approval); err != nil {
		return nil, err
	}
	if allowOwner {
		chassis.StringParam(qs, "owner", &ps.Owner)
	}

	// Parse any embedded link parameters.
	ps.Links, err = ParseLinkParams(qs)
	if err != nil {
		return nil, err
	}

	chassis.StringSliceParam(qs, "ids", &ps.Ids)
	chassis.StringSliceParam(qs, "collections", &ps.Collections)
	chassis.SortingParam(qs, "sort_by", &ps.SortBy)

	return &ps, nil
}

// String converts SearchParams values to a string for debugging.
func (ps SearchParams) String() string {
	ss := []string{}
	if len(ps.ItemTypes) > 0 {
		itemTypes := []string{}
		for _, item := range ps.ItemTypes {
			itemTypes = append(itemTypes, item.String())
		}
		ss = append(ss, "type="+strings.Join(itemTypes, ","))
	}
	if ps.Q != nil {
		ss = append(ss, "q="+*ps.Q)
	}
	ss = append(ss, "format="+string(ps.Format))
	if ps.Geo != nil {
		ss = append(ss, fmt.Sprintf("geo=[%f,%f]", (*ps.Geo)[0], (*ps.Geo)[1]))
	}
	if ps.Dist != nil {
		ss = append(ss, fmt.Sprintf("dist=%f", *ps.Dist))
	}
	if ps.Pagination.Page != 0 {
		ss = append(ss, fmt.Sprintf("page=%d", ps.Pagination.Page))
	}
	if ps.Pagination.PerPage != 0 {
		ss = append(ss, fmt.Sprintf("per_page=%d", ps.Pagination.PerPage))
	}
	if ps.Approval != nil {
		apps := []string{}
		for _, app := range *ps.Approval {
			apps = append(apps, app.String())
		}
		ss = append(ss, "approval="+strings.Join(apps, ","))
	}
	if ps.Owner != nil {
		ss = append(ss, "user="+*ps.Owner)
	}
	if ps.Links != nil {
		ss = append(ss, "links="+ps.Links.String())
	}
	return strings.Join(ss, " ")
}
