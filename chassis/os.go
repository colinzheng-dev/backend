package chassis

import (
	"os"
	"strconv"

	"github.com/rs/zerolog/log"
)

// MustEnv looks up a required environment variable.
func MustEnv(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		log.Fatal().
			Str("variable", key).
			Msg("required environment variable not set")
	}
	return v
}

// EnvInt looks up a required integer-valued environment variable,
// with a default.
func EnvInt(key string, def int) int {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	vi, err := strconv.Atoi(v)
	if err != nil {
		log.Fatal().
			Str("variable", key).
			Msg("environment variable must be an integer")
	}
	return vi
}
