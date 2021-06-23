package server

import (
	"context"
	"time"

	"github.com/veganbase/backend/services/user-service/model"

	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/user-service/db"
	"github.com/veganbase/backend/services/user-service/events"
	"googlemaps.github.io/maps"
)

// Server is the server structure for the user service.
type Server struct {
	chassis.Server
	db         db.DB
	avatarGen  func() string
	stripeKey  string
	mapsClient *maps.Client
	// Disabled encryption key, used for the SSO routes
	// encryptionKey []byte
}

// Config contains the configuration information needed to start
// the user service.
type Config struct {
	AppName      string
	DevMode      bool   `env:"DEV_MODE,default=false"`
	Project      string `env:"PROJECT_ID,default=dev"`
	DBURL        string `env:"DATABASE_URL,required"`
	Port         int    `env:"PORT,default=8080"`
	Credentials  string `env:"CREDENTIALS_PATH"`
	AvatarCount  int    `env:"AVATAR_COUNT,default=0"`
	AvatarFormat string `env:"AVATAR_FORMAT"`
	StripeKey    string `env:"STRIPE_KEY,required"`
	MapsKey      string `env:"GOOGLE_API_KEY,required"`
	// Disabled encryption key, used for the SSO routes
	// EncryptionKey string `env:"ENCRYPTION_KEY,required"`
}

// NewServer creates the server structure for the user service.
func NewServer(cfg *Config) *Server {
	// Common server initialisation.
	s := &Server{
		avatarGen: generateAvatar(cfg.AvatarFormat, cfg.AvatarCount),
	}
	s.Init(cfg.AppName, cfg.Project, cfg.Port, cfg.Credentials, s.routes())

	s.stripeKey = cfg.StripeKey

	// s.encryptionKey = []byte(cfg.EncryptionKey)

	// Connect to user database.
	timeout, _ := context.WithTimeout(context.Background(), time.Second*10)
	var err error
	s.db, err = db.NewPGClient(timeout, cfg.DBURL)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't connect to user database")
	}

	if cfg.MapsKey != "" {
		if s.mapsClient, err = maps.NewClient(maps.WithAPIKey(cfg.MapsKey)); err != nil {
			log.Fatal().Err(err).Msg("couldn't connect to google maps client")
		}
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

// Invalidate performs cache invalidate for user service clients.
func (s *Server) Invalidate(id string) {
	s.PubSub.Publish(events.UserCacheInvalTopic, id)
}
