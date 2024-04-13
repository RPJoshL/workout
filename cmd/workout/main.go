package main

import (
	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/go-webserver/webserver"
	"git.rpjosh.de/RPJosh/workout/internal/api"
	"git.rpjosh.de/RPJosh/workout/internal/models"
)

func main() {
	defer logger.CloseFile()

	// Get the generic configuration of the app
	conf := models.GetAppConfig()
	logger.Debug("Using main CSS file: %q", conf.CssFileName)

	// Set up the webserver
	webApp := webserver.WebServer[*models.AppConfig]{
		Logger:     logger.GetGlobalLogger(),
		Dependency: conf,
		Config: &webserver.WebConfig{
			Address: conf.Address,
		},
	}
	webApp.Setup(api.Routes)

	webApp.Start()
}
