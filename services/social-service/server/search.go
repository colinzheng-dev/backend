package server

import (
	"errors"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/social-service/db"
	"github.com/veganbase/backend/services/social-service/model"
	"net/http"
	"net/url"
	"strings"
)
// ResultFormat is a wrapper around a string describing the result
// format from a search, either full results or summary results.
type ResultFormat string

// Constant values for results formats.
const (
	SummaryResults ResultFormat = "summary"
	FullResults    ResultFormat = "full"
)

// SearchParams represents all the possible query parameters on search routes.
type SearchParams struct {
	db.DatabaseParams
	Format    ResultFormat
}


// Params processes all the possible query parameters for a search route.
func Params(r *http.Request, allowOwner bool) (*SearchParams, error) {
	qs, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return nil, errors.New("invalid query parameters")
	}
	ps := SearchParams{	DatabaseParams: db.DatabaseParams{Pagination: &chassis.Pagination{},}}

	// Pagination parameters.
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

	// Filtering parameters.
	if ps.PostTypes, err = PostTypeParam(qs); err != nil {
		return nil, err
	}

	if err = chassis.SortingParam(qs,"sort_by", &ps.SortBy); err != nil {
		return nil, err
	}

	return &ps, nil
}


func PostTypeParam(qs url.Values) ([]string, error) {
	postTypes := []string{}

	s := qs.Get("type")
	ss := strings.Split(s, ",")

	for _, s := range ss {
		if s != "" {
			postType := model.UnknownPost
			if err := postType.FromString(s); err != nil {
				return nil, err
			}

			postTypes = append(postTypes, postType.String())
		}
	}

	return postTypes, nil
}