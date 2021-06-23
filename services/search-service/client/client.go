package client

// Client is the service client API for the search service.
type Client interface {
	// Geo performs a geolocation search.
	Geo(latitude, longitude, dist float64) ([]string, error)

	// FullText performs a full-text search.
	FullText(q string) ([]string, error)

	// Region search items inside a region.
	Region(regionRef, regionType string) ([]string, error)

	CheckRegions(latitude, longitude float64, regions []int) (*bool, error)
}
