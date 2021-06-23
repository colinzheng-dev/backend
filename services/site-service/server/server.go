package server

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/site-service/db"
	"github.com/veganbase/backend/services/site-service/events"
	"github.com/veganbase/backend/services/site-service/model"

	"time"
)

// Server is the server structure for the user service.
type Server struct {
	chassis.Server
	db    db.DB
	lock  sync.RWMutex
	sites map[string]*model.Site
}

// Config contains the configuration information needed to start
// the user service.
type Config struct {
	AppName     string
	DevMode     bool   `env:"DEV_MODE,default=false"`
	Project     string `env:"PROJECT_ID,default=dev"`
	DBURL       string `env:"DATABASE_URL,required"`
	Port        int    `env:"PORT,default=8080"`
	Credentials string `env:"CREDENTIALS_PATH"`
}

// NewServer creates the server structure for the user service.
func NewServer(cfg *Config) *Server {
	// Common server initialisation.
	s := &Server{}
	s.Init(cfg.AppName, cfg.Project, cfg.Port, cfg.Credentials, s.routes())

	// Connect to site database.
	timeout, _ := context.WithTimeout(context.Background(), time.Second*10)
	var err error
	s.db, err = db.NewPGClient(timeout, cfg.DBURL)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't connect to site database")
	}

	s.sites, err = s.db.Sites()
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't read sites from database")
	}
	err = chassis.Emit(s, events.SiteUpdate, s.sites)
	if err != nil {
		log.Error().Err(err).Msg("failed emitting site update event")
	}

	return s
}

// Publish routes messages to the server's Pub/Sub stream.
func (s *Server) Publish(topic string, eventData interface{}) error {
	return s.PubSub.Publish(topic, eventData)
}

// SaveEvent routes messages to the server's database.
func (s *Server) SaveEvent(topic string, eventData interface{}, inTx func() error) error {
	return s.db.SaveEvent(topic, eventData, inTx)
}
