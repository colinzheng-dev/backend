package server

import (
	"context"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/chassis"
	item "github.com/veganbase/backend/services/item-service/client"
	purchase "github.com/veganbase/backend/services/purchase-service/client"
	"github.com/veganbase/backend/services/shipping-service/integrations/shippypro"
	user "github.com/veganbase/backend/services/user-service/client"
)

type Server struct {
	chassis.Server
	AppName        string
	Ctx            context.Context
	Srv            *http.Server
	userSvc        user.Client
	purchaseSvc    purchase.Client
	itemSvc        item.Client
	shippyProSvc   *shippypro.Shippo
	apiOrderID     int
	integrationKey string
}

type Config struct {
	//Connector configurations
	AppName string
	DevMode bool   `env:"DEV_MODE,default=false"`
	Project string `env:"PROJECT_ID,default=dev"`
	// DBURL              string `env:"DATABASE_URL,required"`
	Port               int    `env:"PORT,default=8094"`
	Credentials        string `env:"CREDENTIALS_PATH"`
	UserServiceURL     string `env:"USER_SERVICE_URL,default=http://user-service"`
	PurchaseServiceURL string `env:"PURCHASE_SERVICE_URL,default=http://purchase-service"`
	ItemServiceURL     string `env:"ITEM_SERVICE_URL,default=http://item-service"`
	APIOrdersID        int    `env:"API_ORDER_ID,required"`
	// Integration              string `env:"INTEGRATION,default=magento"`
	// IntegrationURL           string `env:"INTEGRATION_URL,required"`
	// Username                 string `env:"USERNAME"`
	// Pwd                      string `env:"PWD"`
	IntegrationKey string `env:"INTEGRATION_KEY,required"`
}

// NewServer creates the server structure for the payment service.
func NewServer(cfg *Config) *Server {
	var err error
	s := &Server{}
	s.InitSimple(cfg.AppName, cfg.Project, cfg.Port, cfg.Credentials)
	s.integrationKey = cfg.IntegrationKey
	s.apiOrderID = cfg.APIOrdersID

	// Connect to connector's database.
	// timeout, _ := context.WithTimeout(context.Background(), time.Second*10)

	// if s.db, err = db.NewPGClient(timeout, cfg.DBURL); err != nil {
	// 	log.Fatal().Err(err).Msg("couldn't connect to connector's database")
	// }
	// switch strings.ToLower(cfg.Integration) {
	// case "magento":
	// 	//loading configurations on the database
	// 	//TODO: VALIDATE USER, PWD, PAYMENT_METHODS AND USE_GUEST_CARTS
	// 	rawAddr, err := s.db.ConfigurationsByType("default-address")
	// 	var address mag.Address

	// 	if err != nil {
	// 		if err == db.ErrConfigurationSettingsNotFound { //ignore if the configuration is not set
	// 			log.Info().Err(err).Msg("default-address was not loaded")
	// 		} else {
	// 			log.Fatal().Err(err).Msg("error accessing the database")
	// 		}
	// 	} else {
	// 		if err = json.Unmarshal(rawAddr.Entries[0].Value, &address); err != nil {
	// 			log.Fatal().Err(err).Msg("error unmarshalling default-address configuration")
	// 		}
	// 	}
	// 	if s.Integration, err = magento.New(cfg.IntegrationURL, cfg.DefaultPaymentMethodName, cfg.DefaultPaymentMethod,
	// 		cfg.DefaultShippingMethod, cfg.Username, cfg.Pwd, cfg.IntegrationKey, &address, cfg.UseGuestCarts, cfg.DevMode); err != nil {
	// 		log.Fatal().Err(err).Msg("couldn't initialize Magento's integration")
	// 	}
	// 	go s.SyncItems()
	// 	go s.ProcessOrders()
	// case "shippo":
	// 	//loading configurations on the database
	// 	rawAddr, err := s.db.ConfigurationsByType("default-address")
	// 	var addressInput ship.AddressInput
	// 	var address *ship.Address
	// 	if err != nil {
	// 		if err == db.ErrConfigurationSettingsNotFound { //ignore if the configuration is not set
	// 			log.Info().Err(err).Msg("default-address was not loaded")
	// 		} else {
	// 			log.Fatal().Err(err).Msg("error accessing the database")
	// 		}
	// 	} else {
	// 		if err = json.Unmarshal(rawAddr.Entries[0].Value, &addressInput); err != nil {
	// 			log.Fatal().Err(err).Msg("error unmarshalling default-address configuration")
	// 		}
	// 		addr := ship.Address{AddressInput: addressInput}
	// 		address = &addr
	// 	}
	// 	if s.Integration, err = shippo.New(cfg.IntegrationURL, cfg.IntegrationKey, cfg.DevMode, &s.trackingChannel, address, s.db); err != nil {
	// 		log.Fatal().Err(err).Msg("couldn't initialize Magento's integration")
	// 	}
	// 	go s.ProcessOrders()
	// default:
	// 	log.Fatal().Msg("integration type " + cfg.Integration + " is not implemented")
	// }
	s.purchaseSvc = purchase.New(cfg.PurchaseServiceURL)
	s.itemSvc = item.New(cfg.ItemServiceURL)

	if s.userSvc, err = user.New(cfg.UserServiceURL, s.PubSub, s.AppName); err != nil {
		log.Fatal().Err(err).Msg("couldn't initiate user-service client")
	}

	if s.shippyProSvc, err = shippypro.New("https://www.shippypro.com/api", cfg.IntegrationKey, cfg.DevMode); err != nil {
		log.Fatal().Err(err).Msg("couldn't initialize shippy pro")
	}
	return s
}
