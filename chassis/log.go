package chassis

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LogSetup performs logging setup common to all services.
func LogSetup(appname string, dev bool) {
	baselog := zerolog.New(os.Stdout)
	if dev {
		baselog = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout})
	}
	applog := baselog.With().Timestamp().Str("service", appname).Logger()
	log.Logger = applog
}
