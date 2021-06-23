package server

import (
	"context"
	"time"

	"github.com/veganbase/backend/services/purchase-service/model"

	"github.com/rs/zerolog/log"

	"github.com/veganbase/backend/chassis"
	cart "github.com/veganbase/backend/services/cart-service/client"
	item "github.com/veganbase/backend/services/item-service/client"
	payment "github.com/veganbase/backend/services/payment-service/client"
	"github.com/veganbase/backend/services/purchase-service/db"
	search "github.com/veganbase/backend/services/search-service/client"
	user "github.com/veganbase/backend/services/user-service/client"
)

// Server is the server structure for the purchase.
type Server struct {
	chassis.Server
	db           db.DB
	itemSvc      item.Client
	cartSvc      cart.Client
	paymentSvc   payment.Client
	userSvc      user.Client
	searchSvc    search.Client
	imageBaseURL string
	// redisClient  chassis.RedisClient
	isDevMode bool
}

// Config contains the configuration information needed to start
// the purchase service.
type Config struct {
	AppName           string
	DevMode           bool   `env:"DEV_MODE,default=false"`
	Project           string `env:"PROJECT_ID,default=dev"`
	DBURL             string `env:"DATABASE_URL,required"`
	Port              int    `env:"PORT,default=8089"`
	Credentials       string `env:"CREDENTIALS_PATH"`
	ItemServiceURL    string `env:"ITEM_SERVICE_URL,default=http://item-service"`
	CartServiceURL    string `env:"CART_SERVICE_URL,default=http://cart-service"`
	PaymentServiceURL string `env:"PAYMENT_SERVICE_URL,default=http://payment-service"`
	UserServiceURL    string `env:"USER_SERVICE_URL,default=http://user-service"`
	SearchServiceURL  string `env:"SEARCH_SERVICE_URL,default=http://search-service"`
	// RedisAddress      string `env:"REDIS_CLIENT,default=http://redis-service"`
	// RedisPWD          string `env:"REDIS_PWD,required"`
}

// NewServer creates the server structure for the puchase service.
func NewServer(cfg *Config) *Server {
	// Backend service URL parsing.

	chassis.CheckURL(cfg.ItemServiceURL, "item service")
	chassis.CheckURL(cfg.CartServiceURL, "cart service")
	chassis.CheckURL(cfg.PaymentServiceURL, "payment service")
	chassis.CheckURL(cfg.UserServiceURL, "user service")
	chassis.CheckURL(cfg.SearchServiceURL, "search service")

	// Common server initialisation.
	s := &Server{}
	s.Init(cfg.AppName, cfg.Project, cfg.Port, cfg.Credentials, s.routes())
	var err error

	s.itemSvc = item.New(cfg.ItemServiceURL)
	s.cartSvc = cart.New(cfg.CartServiceURL)
	s.paymentSvc = payment.New(cfg.PaymentServiceURL)
	s.searchSvc = search.New(cfg.SearchServiceURL)

	s.isDevMode = cfg.DevMode

	if s.userSvc, err = user.New(cfg.UserServiceURL, s.PubSub, s.AppName); err != nil {
		log.Fatal().Err(err).Msg("couldn't initiate user-service client")
	}

	// Connect to purchase database.
	timeout, _ := context.WithTimeout(context.Background(), time.Second*10)
	s.db, err = db.NewPGClient(timeout, cfg.DBURL)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't connect to user database")
	}
	// XXX: I've disabled the redis client and the APIs associated with it to make
	// the base purchase service work. Recurring subscriptions cannot be configured until
	// this is resolved, which requires creating a new redis instance in the k8s cluster
	// s.redisClient = chassis.NewRedisClient(cfg.RedisAddress, cfg.RedisPWD)
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
