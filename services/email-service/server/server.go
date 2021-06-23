package server

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/email-service/db"
	"github.com/veganbase/backend/services/email-service/mailer"
	"github.com/veganbase/backend/services/email-service/model"
	site "github.com/veganbase/backend/services/site-service/client"
)

// Server is the server structure for the user service.
type Server struct {
	chassis.Server
	db         db.DB
	mailer     mailer.Mailer
	muTopic    sync.RWMutex
	topics     map[string]*model.TopicInfo
	muSubs     sync.Mutex
	subClosers map[string]chan bool
	muxCh      chan subEvent
	siteSvc    site.Client
}

type subEvent struct {
	topic string
	data  []byte
}

// Config contains the configuration information needed to start
// the user service.
type Config struct {
	AppName            string
	DevMode            bool   `env:"DEV_MODE,default=false"`
	Project            string `env:"PROJECT_ID,default=dev"`
	DBURL              string `env:"DATABASE_URL,required"`
	Port               int    `env:"PORT,default=8080"`
	Credentials        string `env:"CREDENTIALS_PATH"`
	MJPublicKey        string `env:"MAILJET_API_KEY_PUBLIC"`
	MJPrivateKey       string `env:"MAILJET_API_KEY_PRIVATE"`
	SimultaneousEmails int    `env:"SIMULTANEOUS_EMAILS,default=10"`
	SiteServiceURL     string `env:"SITE_SERVICE_URL,default=http://site-service"`
}

// NewServer creates the server structure for the user service.
func NewServer(cfg *Config) *Server {
	// Backend service URL parsing.
	chassis.CheckURL(cfg.SiteServiceURL, "site service")

	// Common server initialisation.
	s := &Server{
		topics:     map[string]*model.TopicInfo{},
		subClosers: map[string]chan bool{},
	}
	s.InitSimple(cfg.AppName, cfg.Project, cfg.Port, cfg.Credentials)
	s.siteSvc = site.New(cfg.SiteServiceURL, s.PubSub, s.AppName)

	// Connect to email service database.
	timeout, _ := context.WithTimeout(context.Background(), time.Second*10)
	var err error
	s.db, err = db.NewPGClient(timeout, cfg.DBURL)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't connect to email database")
	}

	// Initialise mailer.
	if cfg.MJPublicKey == "" || cfg.MJPrivateKey == "" ||
		cfg.MJPublicKey == "dev" || cfg.MJPrivateKey == "dev" {
		log.Info().Msg("using development mailer")
		s.mailer = mailer.NewDevMailer()
	} else {
		s.mailer, err = mailer.NewMailjetMailer(cfg.MJPublicKey, cfg.MJPrivateKey)
		if err != nil {
			log.Fatal().Err(err).Msg("couldn't connect to Mailjet")
		}
	}

	// Set up email topic subscription multiplexing channel and email
	// event processor.
	s.muxCh = make(chan subEvent, cfg.SimultaneousEmails)
	go s.sender()

	return s
}

// Publish is a placeholder for event publication (not used here yet).
func (s *Server) Publish(topic string, eventData interface{}) error {
	return s.PubSub.Publish(topic, eventData)
}

// SaveEvent saves and publishes an event.
func (s *Server) SaveEvent(topic string, eventData interface{}, inTx func() error) error {
	return s.db.SaveEvent(topic, eventData, inTx)
}
