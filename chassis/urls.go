package chassis

import (
	"net/url"

	"github.com/rs/zerolog/log"
)

// CheckURL parses URLs for service endpoints.
func CheckURL(inURL string, name string) *url.URL {
	url, err := url.Parse(inURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse " + name + " URL")
	}
	return url
}
