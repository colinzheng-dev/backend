package server

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/category-service/db"

	"time"
)

// TODO: MAKE THIS MORE LIKE THE SITE SERVICE? I.E. COLLECT CATEGORY
// INFORMATION ONCE AND ONLY UPDATE IT WHEN THE DATABASE CHANGES...

// Server is the server structure for the category service.
type Server struct {
	chassis.Server
	db db.DB
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

	// Connect to category database.
	timeout, _ := context.WithTimeout(context.Background(), time.Second*10)
	var err error
	s.db, err = db.NewPGClient(timeout, cfg.DBURL)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't connect to category database")
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
