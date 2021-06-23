package main

import (
	"github.com/joeshaw/envdecode"
	"github.com/rs/zerolog/log"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/purchase-service/server"
)

var appname = "purchase-service"

func main() {
	cfg := server.Config{AppName: appname}
	err := envdecode.StrictDecode(&cfg)
	if err != nil {
		log.Fatal().Err(err).
			Msg("failed to process environment variables")
	}
	chassis.LogSetup(appname, cfg.DevMode)
	serv := server.NewServer(&cfg)
	go serv.ScheduleItemSubscriptionProcessingJobs()

	serv.Serve()
}
