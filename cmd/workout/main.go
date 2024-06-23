package main

import (
	"os"
	"time"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/cmd/workout/args"
	"git.rpjosh.de/RPJosh/workout/internal/api"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/webserver"

	_ "time/tzdata"
)

func main() {
	defer logger.CloseFile()

	// Use UTC timezone globally
	os.Setenv("TZ", "UTC")

	// Get the generic configuration of the app
	conf := models.GetAppConfig()

	// Parse the command line
	if err := args.ParseArgs(conf, os.Args); err != nil {
		logger.Fatal("Unable to parse the command line")
	}

	logger.Debug("Using main CSS file: %q", conf.CssFileName)
	logger.Debug("Using 3dParty JS file: %s", conf.Js3dPartyFileName)

	// Set up the webserver
	webApp := webserver.WebServer[*models.AppConfig]{
		Logger:     logger.GetGlobalLogger(),
		Dependency: conf,
		Config: &webserver.WebConfig{
			Address:     conf.Address,
			ReadTimeout: 30 * time.Second,
		},
	}
	webApp.Setup(api.Routes)

	webApp.Start()
}
