package server

import (
	"context"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/api-gateway/db"
	site "github.com/veganbase/backend/services/site-service/client"
	user "github.com/veganbase/backend/services/user-service/client"
)

// Server is the server structure for the blob service.
type Server struct {
	chassis.Server
	db                db.DB
	siteSvc           site.Client
	siteSvcURL        *url.URL
	userSvc           user.Client
	userSvcURL        *url.URL
	blobSvcURL        *url.URL
	itemSvcURL        *url.URL
	categorySvcURL    *url.URL
	cartSvcURL        *url.URL
	socialSvcURL      *url.URL
	purchaseSvcURL    *url.URL
	paymentSvcURL     *url.URL
	webhookSvcURL     *url.URL
	searchSvcURL      *url.URL
	secureSession     bool
	muCORSOrigins     sync.RWMutex
	corsOrigins       map[string]bool
	configCORSOrigins []string
	relaxCORSKey      string
	muSiteURLs        sync.RWMutex
	siteURLs          map[string]string
	disabledCSRF      bool
}

// Config contains the configuration information needed to start
// the blob service.
type Config struct {
	AppName            string
	DevMode            bool   `env:"DEV_MODE,default=false"`
	Project            string `env:"PROJECT_ID,default=dev"`
	DBURL              string `env:"DATABASE_URL,required"`
	Port               int    `env:"PORT,default=8080"`
	Credentials        string `env:"CREDENTIALS_PATH"`
	UserServiceURL     string `env:"USER_SERVICE_URL,default=http://user-service"`
	BlobServiceURL     string `env:"BLOB_SERVICE_URL,default=http://blob-service"`
	ItemServiceURL     string `env:"ITEM_SERVICE_URL,default=http://item-service"`
	SiteServiceURL     string `env:"SITE_SERVICE_URL,default=http://site-service"`
	CategoryServiceURL string `env:"CATEGORY_SERVICE_URL,default=http://category-service"`
	CartServiceURL     string `env:"CART_SERVICE_URL,default=http://cart-service"`
	SocialServiceURL   string `env:"SOCIAL_SERVICE_URL,default=http://social-service"`
	PurchaseServiceURL string `env:"PURCHASE_SERVICE_URL,default=http://purchase-service"`
	PaymentServiceURL  string `env:"PAYMENT_SERVICE_URL,default=http://payment-service"`
	WebhookServiceURL  string `env:"WEBHOOK_SERVICE_URL,default=http://webhook-service"`
	SearchServiceURL   string `env:"SEARCH_SERVICE_URL,default=http://search-service"`
	CSRFSecret         string `env:"CSRF_SECRET"`
	CORSOrigins        string `env:"CORS_ORIGINS"`
	RelaxCORSKey       string `env:"RELAX_CORS_KEY,default=no"`
	DisableCSRF        bool   `env:"DISABLE_CSRF,default=false"`
}

// NewServer creates the server structure for the blob service.
func NewServer(cfg *Config) *Server {
	// Backend service URL parsing.
	userSvcURL := chassis.CheckURL(cfg.UserServiceURL, "user service")
	blobSvcURL := chassis.CheckURL(cfg.BlobServiceURL, "blob service")
	itemSvcURL := chassis.CheckURL(cfg.ItemServiceURL, "item service")
	siteSvcURL := chassis.CheckURL(cfg.SiteServiceURL, "site service")
	categorySvcURL := chassis.CheckURL(cfg.CategoryServiceURL, "category service")
	cartSvcURL := chassis.CheckURL(cfg.CartServiceURL, "cart service")
	socialSvcURL := chassis.CheckURL(cfg.SocialServiceURL, "social service")
	purchaseSvcURL := chassis.CheckURL(cfg.PurchaseServiceURL, "purchase service")
	paymentSvcURL := chassis.CheckURL(cfg.PaymentServiceURL, "payment service")
	webhookSvcURL := chassis.CheckURL(cfg.WebhookServiceURL, "webhook service")
	searchSvcURL := chassis.CheckURL(cfg.SearchServiceURL, "webhook service")

	// Fixed CORS origin list from environment.
	corsOrigins := []string{}
	if len(cfg.CORSOrigins) > 0 {
		corsOrigins = strings.Split(cfg.CORSOrigins, ",")
	}

	// Common server initialisation.
	s := &Server{
		userSvcURL:        userSvcURL,
		blobSvcURL:        blobSvcURL,
		itemSvcURL:        itemSvcURL,
		siteSvcURL:        siteSvcURL,
		categorySvcURL:    categorySvcURL,
		socialSvcURL:      socialSvcURL,
		cartSvcURL:        cartSvcURL,
		purchaseSvcURL:    purchaseSvcURL,
		paymentSvcURL:     paymentSvcURL,
		webhookSvcURL:     webhookSvcURL,
		searchSvcURL:      searchSvcURL,
		secureSession:     !cfg.DevMode,
		corsOrigins:       map[string]bool{},
		configCORSOrigins: corsOrigins,
		relaxCORSKey:      cfg.RelaxCORSKey,
		siteURLs:          map[string]string{},
		disabledCSRF:      cfg.DisableCSRF,
	}

	s.Init(cfg.AppName, cfg.Project, cfg.Port, cfg.Credentials, s.routes(cfg.DevMode, cfg.CSRFSecret))
	s.siteSvc = site.New(cfg.SiteServiceURL, s.PubSub, s.AppName)
	var err error
	s.userSvc, err = user.New(cfg.UserServiceURL, s.PubSub, s.AppName)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't initialise user service client")
	}

	// Connect to database.
	timeout, _ := context.WithTimeout(context.Background(), time.Second*10)
	s.db, err = db.NewPGClient(timeout, cfg.DBURL)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't connect to API gateway database")
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
