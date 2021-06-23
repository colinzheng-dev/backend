package server

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/webhook-service/db"

	"time"
)

// Server is the server structure for the webhook service.
type Server struct {
	chassis.Server
	db      db.DB
}

// Config contains the configuration information needed to start
// the search service.
type Config struct {
	AppName              string
	DevMode              bool   `env:"DEV_MODE,default=false"`
	Project              string `env:"PROJECT_ID,default=dev"`
	DBURL                string `env:"DATABASE_URL,required"`
	Port                 int    `env:"PORT,default=8080"`
	Credentials          string `env:"CREDENTIALS_PATH"`
	SimultaneousMessages int    `env:"SIMULTANEOUS_MESSAGES,default=1"`
}

// NewServer creates the server structure for the search service.
func NewServer(cfg *Config) *Server {
	var err error
	// Common server initialisation.

	s := &Server{}
	s.Init(cfg.AppName, cfg.Project, cfg.Port, cfg.Credentials, s.routes())

	// Connect to webhook database.
	timeout, _ := context.WithTimeout(context.Background(), time.Second*10)

	s.db, err = db.NewPGClient(timeout, cfg.DBURL)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't connect to webhook database")
	}

	go s.ReceivedEvents()
	go s.HandleWebhookMessages()
	go s.retrySendEvents()

	return s
}

// SaveEvent routes messages to the server's database.
func (s *Server) SaveEvent(topic string, eventData interface{}, inTx func() error) error {
	return s.db.SaveEvent(topic, eventData, inTx)
}
