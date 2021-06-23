package server

import (
	"context"
	"github.com/stripe/stripe-go"
	site "github.com/veganbase/backend/services/site-service/client"
	"regexp"

	"time"

	"github.com/rs/zerolog/log"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/payment-service/db"
	purchase "github.com/veganbase/backend/services/purchase-service/client"
	user "github.com/veganbase/backend/services/user-service/client"
)

// Server is the server structure for the purchase.
type Server struct {
	chassis.Server
	db           db.DB
	userSvc      user.Client
	siteSvc      site.Client
	purchaseSvc  purchase.Client
	imageBaseURL string
	stripeKey    string
	webhookSecret   string
	livemode bool
}

// Config contains the configuration information needed to start
// the payment service.
type Config struct {
	AppName            string
	DevMode            bool   `env:"DEV_MODE,default=false"`
	Project            string `env:"PROJECT_ID,default=dev"`
	DBURL              string `env:"DATABASE_URL,required"`
	Port               int    `env:"PORT,default=8091"`
	Credentials        string `env:"CREDENTIALS_PATH"`
	UserServiceURL     string `env:"USER_SERVICE_URL,default=http://user-service"`
	SiteServiceURL     string `env:"SITE_SERVICE_URL,default=http://site-service"`
	PurchaseServiceURL string `env:"PURCHASE_SERVICE_URL,default=http://purchase-service"`
	StripeKey          string `env:"STRIPE_KEY,required"`
	WebhookSecret      string `env:"WEBHOOK_SECRET_KEY,required"`
}

// NewServer creates the server structure for the payment service.
func NewServer(cfg *Config) *Server {
	// Backend service URL parsing.

	chassis.CheckURL(cfg.UserServiceURL, "user service")
	chassis.CheckURL(cfg.PurchaseServiceURL, "purchase service")
	chassis.CheckURL(cfg.SiteServiceURL, "site service")

	// Common server initialisation.
	s := &Server{}
	s.Init(cfg.AppName, cfg.Project, cfg.Port, cfg.Credentials, s.routes())

	s.stripeKey = cfg.StripeKey
	s.webhookSecret = cfg.WebhookSecret

	s.livemode = true
	matched, _ := regexp.MatchString(`\w+(_test_)\w+`, s.stripeKey)
	if matched {
		s.livemode = false
	}

	var err error
	if s.userSvc, err = user.New(cfg.UserServiceURL, s.PubSub, s.AppName); err != nil {
		log.Fatal().Err(err).Msg("couldn't initiate user-service client")
	}
	s.purchaseSvc = purchase.New(cfg.PurchaseServiceURL)
	s.siteSvc = site.New(cfg.SiteServiceURL, s.PubSub, s.AppName)

	// Connect to payment's database.
	timeout, _ := context.WithTimeout(context.Background(), time.Second*10)

	s.db, err = db.NewPGClient(timeout, cfg.DBURL)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't connect to user database")
	}

	stripe.Key = s.stripeKey
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
