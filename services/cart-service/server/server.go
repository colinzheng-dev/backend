package server

import (
	"context"
	"github.com/veganbase/backend/services/cart-service/model"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/cart-service/db"
	item "github.com/veganbase/backend/services/item-service/client"
	search "github.com/veganbase/backend/services/search-service/client"
	user "github.com/veganbase/backend/services/user-service/client"
)

// Server is the server structure for the item service.
type Server struct {
	chassis.Server
	db           db.DB
	itemSvc      item.Client
	userSvc      user.Client
	searchSvc    search.Client
	imageBaseURL string
}

// Config contains the configuration information needed to start
// the item service.
type Config struct {
	AppName          string
	DevMode          bool   `env:"DEV_MODE,default=false"`
	Project          string `env:"PROJECT_ID,default=dev"`
	DBURL            string `env:"DATABASE_URL,default=postgres://cart_service@db/vb_carts?sslmode=disable"`
	Port             int    `env:"PORT,default=8086"`
	Credentials      string `env:"CREDENTIALS_PATH"`
	ItemServiceURL   string `env:"ITEM_SERVICE_URL,default=http://item-service"`
	UserServiceURL   string `env:"USER_SERVICE_URL,default=http://user-service"`
	SearchServiceURL string `env:"SEARCH_SERVICE_URL,default=http://search-service"`
	ImageBaseURL     string `env:"IMAGE_BASE_URL"`
}

// NewServer creates the server structure for the user service.
func NewServer(cfg *Config) *Server {
	// Backend service URL parsing.

	chassis.CheckURL(cfg.ItemServiceURL, "item service")

	// Common server initialisation.
	s := &Server{
		imageBaseURL: cfg.ImageBaseURL,
	}
	s.Init(cfg.AppName, cfg.Project, cfg.Port, cfg.Credentials, s.routes())

	var err error
	if s.userSvc, err = user.New(cfg.UserServiceURL, s.PubSub, s.AppName); err != nil {
		log.Fatal().Err(err).Msg("couldn't initiate user-service client")
	}

	s.itemSvc = item.New(cfg.ItemServiceURL)
	s.searchSvc = search.New(cfg.SearchServiceURL)
	// Connect to cart database.
	timeout, _ := context.WithTimeout(context.Background(), time.Second*10)
	s.db, err = db.NewPGClient(timeout, cfg.DBURL)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't connect to user database")
	}
	// Load JSON validation schemas.
	model.LoadSchemas()

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
