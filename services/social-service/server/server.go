package server

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/veganbase/backend/chassis"
	blob "github.com/veganbase/backend/services/blob-service/client"
	item "github.com/veganbase/backend/services/item-service/client"
	pur "github.com/veganbase/backend/services/purchase-service/client"
	user "github.com/veganbase/backend/services/user-service/client"
	"github.com/veganbase/backend/services/social-service/db"
	"github.com/veganbase/backend/services/social-service/model"
)

// Server is the server structure for the item service.
type Server struct {
	chassis.Server
	db           db.DB
	itemSvc      item.Client
	blobSvc      blob.Client
	purSvc       pur.Client
	userSvc      user.Client
	imageBaseURL string
}

// Config contains the configuration information needed to start
// the item service.
type Config struct {
	AppName            string
	DevMode            bool   `env:"DEV_MODE,default=false"`
	Project            string `env:"PROJECT_ID,default=dev"`
	DBURL              string `env:"DATABASE_URL,required"`
	Port               int    `env:"PORT,default=8080"`
	Credentials        string `env:"CREDENTIALS_PATH"`
	ItemServiceURL     string `env:"ITEM_SERVICE_URL,default=http://item-service"`
	PurchaseServiceURL string `env:"PURCHASE_SERVICE_URL,default=http://purchase-service"`
	BlobServiceURL     string `env:"BLOB_SERVICE_URL,default=http://blob-service"`
	UserServiceURL     string `env:"USER_SERVICE_URL,default=http://user-service"`
	ImageBaseURL       string `env:"IMAGE_BASE_URL"`
}

// NewServer creates the server structure for the user service.
func NewServer(cfg *Config) *Server {

	// Common server initialisation.
	s := &Server{imageBaseURL: cfg.ImageBaseURL,}
	s.Init(cfg.AppName, cfg.Project, cfg.Port, cfg.Credentials, s.routes())
	chassis.CheckURL(cfg.BlobServiceURL, "blob service")
	chassis.CheckURL(cfg.ItemServiceURL, "item service")
	chassis.CheckURL(cfg.PurchaseServiceURL, "purchase service")
	// Connect to social database.
	var err error
	s.userSvc, err = user.New(cfg.UserServiceURL, s.PubSub, s.AppName)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't initialise user service client")
	}
	timeout, _ := context.WithTimeout(context.Background(), time.Second*10)
	s.itemSvc = item.New(cfg.ItemServiceURL)
	s.blobSvc = blob.New(cfg.BlobServiceURL)
	s.purSvc = pur.New(cfg.PurchaseServiceURL)

	s.db, err = db.NewPGClient(timeout, cfg.DBURL)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't connect to social database")
	}
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
