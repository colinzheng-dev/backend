package server

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/veganbase/backend/chassis"
	blob "github.com/veganbase/backend/services/blob-service/client"
	category "github.com/veganbase/backend/services/category-service/client"
	"github.com/veganbase/backend/services/item-service/db"
	"github.com/veganbase/backend/services/item-service/model"
	search "github.com/veganbase/backend/services/search-service/client"
	social "github.com/veganbase/backend/services/social-service/client"
	user "github.com/veganbase/backend/services/user-service/client"
)

// Server is the server structure for the item service.
type Server struct {
	chassis.Server
	db           db.DB
	blobSvc      blob.Client
	userSvc      user.Client
	searchSvc    search.Client
	categorySvc  category.Client
	socialSvc    social.Client
	imageBaseURL string
	// merchantID                uint64 // hardcoded into content_api.go
	contentAPICredentialsFile string
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
	BlobServiceURL     string `env:"BLOB_SERVICE_URL,default=http://blob-service"`
	UserServiceURL     string `env:"USER_SERVICE_URL,default=http://user-service"`
	SearchServiceURL   string `env:"SEARCH_SERVICE_URL,default=http://search-service"`
	CategoryServiceURL string `env:"CATEGORY_SERVICE_URL,default=http://category-service"`
	SocialServiceURL   string `env:"SOCIAL_SERVICE_URL,default=http://social-service"`
	ImageBaseURL       string `env:"IMAGE_BASE_URL"`
	// MerchantID                uint64 `env:"MERCHANT_ID"`
	ContentAPICredentialsFile string `env:"CONTENT_API_CREDENTIALS_FILE"`
}

// NewServer creates the server structure for the user service.
func NewServer(cfg *Config) *Server {
	// Backend service URL parsing.
	chassis.CheckURL(cfg.BlobServiceURL, "blob service")
	chassis.CheckURL(cfg.UserServiceURL, "user service")
	chassis.CheckURL(cfg.SearchServiceURL, "search service")
	chassis.CheckURL(cfg.CategoryServiceURL, "category service")

	// Common server initialisation.
	s := &Server{
		blobSvc:      blob.New(cfg.BlobServiceURL),
		searchSvc:    search.New(cfg.SearchServiceURL),
		imageBaseURL: cfg.ImageBaseURL,
		// merchantID:                cfg.MerchantID,
		contentAPICredentialsFile: cfg.ContentAPICredentialsFile,
	}
	s.Init(cfg.AppName, cfg.Project, cfg.Port, cfg.Credentials, s.routes())
	var err error
	s.userSvc, err = user.New(cfg.UserServiceURL, s.PubSub, s.AppName)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't initialise user service client")
	}
	s.categorySvc = category.New(cfg.CategoryServiceURL, s.PubSub, s.AppName)
	s.socialSvc = social.New(cfg.SocialServiceURL)
	// Connect to item database.
	timeout, _ := context.WithTimeout(context.Background(), time.Second*10)
	s.db, err = db.NewPGClient(timeout, cfg.DBURL)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't connect to user database")
	}

	// Load JSON validation schemas.
	model.LoadSchemas(s.categorySvc)

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
