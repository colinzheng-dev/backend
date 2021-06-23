package main

import (
	"github.com/joeshaw/envdecode"
	"github.com/rs/zerolog/log"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/item-service/server"
)

var appname = "item-service"

func main() {
	cfg := server.Config{AppName: appname}
	err := envdecode.StrictDecode(&cfg)
	if err != nil {
		log.Fatal().Err(err).
			Msg("failed to process environment variables")
	}
	chassis.LogSetup(appname, cfg.DevMode)
	serv := server.NewServer(&cfg)
	go serv.HandleItemRank()
	go serv.HandleItemUpvotes()
	// content API handling
	go serv.HandleItemCreateOrUpdate()
	go serv.HandleItemDeletion()

	serv.Serve()
}
