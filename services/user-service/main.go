package main

import (
	"fmt"

	"github.com/joeshaw/envdecode"
	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/user-service/server"
)

const appname = "user-service"

func main() {
	cfg := server.Config{AppName: appname}
	err := envdecode.StrictDecode(&cfg)
	if err != nil {
		log.Fatal().Err(err).
			Msg(fmt.Sprintf("failed to process environment variables: %s", err.Error()))
	}
	chassis.LogSetup(appname, cfg.DevMode)
	serv := server.NewServer(&cfg)
	serv.Serve()
}
