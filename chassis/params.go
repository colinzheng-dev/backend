package chassis

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Sorting struct {
	Field string
	Order string
}

type Pagination struct {
	Page uint
	PerPage uint
}

// IntParam extracts an integer URL query parameter.
func IntParam(qs url.Values, k string, dst *uint) error {
	s := qs.Get(k)
	if s != "" {
		i, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		*dst = uint(i)
	}
	return nil
}

// FloatParam extracts a floating point URL query parameter.
func FloatParam(qs url.Values, k string, dst **float64) error {
	s := qs.Get(k)
	if s != "" {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		if f >= 0.0 {
			*dst = &f
		}
	}
	return nil
}

// StringParam extracts a string URL query parameter.
func StringParam(qs url.Values, k string, dst **string) {
	s := qs.Get(k)
	if s != "" {
		val := s
		*dst = &val
	}
}

// StringSliceParam extracts a string URL query parameter and transforms it into a slice.
func StringSliceParam(qs url.Values, k string, dst **[]string) {
	var ids []string
	rawIds := qs.Get(k)
	if rawIds != "" {
		ids = strings.Split(rawIds, ",")
	}
	*dst = &ids
}

// IntSliceParam extracts an int list from URL query parameter and transforms it into a slice.
func IntSliceParam(qs url.Values, k string, dst **[]int) {
	var ids []int
	rawIds := qs.Get(k)
	if rawIds != "" {
		stringIds := strings.Split(rawIds, ",")
		for _, id :=range stringIds {
			if v, err := strconv.Atoi(id); err == nil {
				ids = append(ids, v)
			}
		}
	}
	*dst = &ids
}

// GeoParam extracts a latitude, longitude coordinate pair from a URL
// query parameter.
func GeoParam(qs url.Values, k string, dst **[2]float64) error {
	s := qs.Get(k)
	if s == "" {
		return nil
	}
	coords := strings.Split(s, ",")
	if len(coords) != 2 {
		return errors.New("invalid geo value")
	}
	lat, err := strconv.ParseFloat(coords[0], 64)
	if err != nil {
		return err
	}
	lon, err := strconv.ParseFloat(coords[1], 64)
	if err != nil {
		return err
	}
	geo := [2]float64{lat, lon}
	*dst = &geo
	return nil
}

// PaginationParams retrieves pagination parameters.
func PaginationParams(qs url.Values, page, perPage *uint) error {
	*page = 1
	*perPage = 30
	if err := IntParam(qs, "page", page); err != nil {
		return err
	}
	if err := IntParam(qs, "per_page", perPage); err != nil {
		return err
	}
	return nil
}

const (
	FunctionSortPattern  = `((asc|desc)\([A-Za-z_]+\))`
	AttributeSortPattern = `([A-Za-z_]+\.(asc|desc))`
)

func SortingParam(qs url.Values, k string, dst **Sorting) error {
	sortingParams := Sorting{}
	s := qs.Get(k)
	if s == "" {
		return nil
	}
	if ok, _ := regexp.MatchString(FunctionSortPattern, s); ok {
		broken := strings.Split(s, "(")
		sortingParams.Field = broken[1][:len(broken)-1]
		sortingParams.Order = broken[0]
	} else if ok, _ := regexp.MatchString(AttributeSortPattern, s); ok {
		broken := strings.Split(s, ".")
		sortingParams.Field = broken[0]
		sortingParams.Order = broken[1]
	} else {
		return errors.New("sorting param not matching any pattern")
	}

	*dst = &sortingParams
	return nil
}
