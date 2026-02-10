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

var Version string

func main() {
	defer logger.CloseFile()

	// Use UTC timezone globally
	if err := os.Setenv("TZ", "UTC"); err != nil {
		logger.Warning("Failed to normalize the time zone to UTC: %s", err)
	}

	// Get the generic configuration of the app
	conf := models.GetAppConfig(Version)

	// Parse the command line
	if err := args.ParseArgs(conf, os.Args, Version); err != nil {
		logger.Fatal("Unable to parse the command line")
	}

	logger.Debug("Using main CSS file: %q", conf.CssFileName)
	logger.Debug("Using 3dParty JS file: %s", conf.Js3dPartyFileName)

	// Set up the webserver
	webApp := webserver.WebServer[*models.AppConfig]{
		Logger:     logger.GetGlobalLogger(),
		Dependency: conf,
		Config: &webserver.WebConfig{
			Address: conf.Address,
			// WearOS is really slow over bluetooth!
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 80 * time.Second,
		},
	}
	webApp.Setup(api.Routes)

	webApp.Start()
}
