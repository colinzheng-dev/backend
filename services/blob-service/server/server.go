package server

import (
	"context"
	"time"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/chassis/storage"
	"github.com/veganbase/backend/services/blob-service/db"
	"github.com/veganbase/backend/services/blob-service/model"

	"github.com/rs/zerolog/log"
)

// Server is the server structure for the blob service.
type Server struct {
	chassis.Server
	db           db.DB
	blobstore    storage.Storage
	imageBaseURL string
}

// Config contains the configuration information needed to start
// the blob service.
type Config struct {
	AppName      string
	DevMode      bool   `env:"DEV_MODE,default=false"`
	Project      string `env:"PROJECT_ID,default=dev"`
	DBURL        string `env:"DATABASE_URL,required"`
	Port         int    `env:"PORT,default=8080"`
	Credentials  string `env:"CREDENTIALS_PATH"`
	BucketName   string `env:"BUCKET_NAME"`
	ImageBaseURL string `env:"IMAGE_BASE_URL"`
}

// NewServer creates the server structure for the blob service.
func NewServer(cfg *Config) *Server {
	// Common server initialisation.
	s := &Server{imageBaseURL: cfg.ImageBaseURL}
	s.Init(cfg.AppName, cfg.Project, cfg.Port, cfg.Credentials, s.routes())

	// Connect to blob database.
	timeout, _ := context.WithTimeout(context.Background(), time.Second*10)
	var err error
	s.db, err = db.NewPGClient(timeout, cfg.DBURL)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't connect to blob database")
	}

	// Connect to blob storage.
	if !cfg.DevMode {
		s.blobstore, err = storage.NewGoogleClient(s.Ctx, cfg.Credentials, cfg.BucketName)
	} else {
		blobs := map[string]*model.Blob{}
		s.blobstore = storage.NewMockStorage(&blobs)
	}
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't connect to blob storage")
	}

	return s
}

// Publish is a placeholder for event publication (not used here yet).
func (s *Server) Publish(topic string, eventData interface{}) error {
	return nil
}

// SaveEvent saves and publishes an event.
func (s *Server) SaveEvent(topic string, eventData interface{}, inTx func() error) error {
	return s.db.SaveEvent(topic, eventData, inTx)
}
