package server

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/chassis"
	item "github.com/veganbase/backend/services/item-service/client"
	"github.com/veganbase/backend/services/search-service/db"

	"time"
)

// Server is the server structure for the search service.
type Server struct {
	chassis.Server
	db      db.DB
	itemSvc item.Client
}

// Config contains the configuration information needed to start
// the search service.
type Config struct {
	AppName        string
	DevMode        bool   `env:"DEV_MODE,default=false"`
	Project        string `env:"PROJECT_ID,default=dev"`
	DBURL          string `env:"DATABASE_URL,required"`
	Port           int    `env:"PORT,default=8080"`
	Credentials    string `env:"CREDENTIALS_PATH"`
	ItemServiceURL string `env:"ITEM_SERVICE_URL,default=http://item-service"`
}

// NewServer creates the server structure for the search service.
func NewServer(cfg *Config) *Server {
	// Backend service URL parsing.
	chassis.CheckURL(cfg.ItemServiceURL, "item service")

	// Common server initialisation.
	s := &Server{
		itemSvc: item.New(cfg.ItemServiceURL),
	}
	s.Init(cfg.AppName, cfg.Project, cfg.Port, cfg.Credentials, s.routes())

	// Connect to search database.
	timeout, _ := context.WithTimeout(context.Background(), time.Second*10)
	var err error
	s.db, err = db.NewPGClient(timeout, cfg.DBURL)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't connect to search database")
	}

	return s
}

// SaveEvent routes messages to the server's database.
func (s *Server) SaveEvent(topic string, eventData interface{}, inTx func() error) error {
	return s.db.SaveEvent(topic, eventData, inTx)
}
