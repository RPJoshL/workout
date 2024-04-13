package models

import (
	"strings"

	"git.rpjosh.de/RPJosh/go-logger"
	rpFiles "git.rpjosh.de/RPJosh/workout"
	"git.rpjosh.de/RPJosh/workout/pkg/utils"
)

// DbConfig contains configuration options for a mysql db
type DbConfig struct {

	// Address, port and database name (like 127.0.0.1:3306/myDbName)
	Address string

	// User for connecting to the database
	User string

	// Db to connect to
	Db string

	// Password of the user
	Password string
}

// AppConfig contains the generic configuration options for the app.
type AppConfig struct {

	// Address on which the WebServer should listen on
	Address string

	// Full qualified domain of the site
	FQDN string

	// Enabled some development specific function
	DevMode bool

	// File name of the main CSS file to use (the filename is dynamic to not cache this important file)
	CssFileName string

	// Private JWT key
	JWTKey []byte

	// Database configuration to use
	Db DbConfig
}

// GetAppConfig fetches all configuration options from the current environment
// variables. It panics if not all information were provided correctly
func GetAppConfig() *AppConfig {

	// Apply logger configuration
	logger.SetGlobalLogger(logger.GetLoggerFromEnv(&logger.Logger{
		ColoredOutput: true,
		Level:         logger.LevelInfo,
		PrintSource:   true,
		File:          &logger.FileLogger{},
	}))

	config := &AppConfig{
		Address:     utils.GetEnvString("SERVER_ADDRESS", "0.0.0.0:4020"),
		FQDN:        utils.RequireEnvString("SERVER_FQDN"),
		DevMode:     utils.GetEnvBool("DEV_MODE", false),
		JWTKey:      []byte(utils.RequireEnvSecret("JWT_KEY")),
		CssFileName: getCssFileName(),
		Db: DbConfig{
			Address:  utils.RequireEnvString("DB_ADDRESS"),
			User:     utils.RequireEnvString("DB_USER"),
			Password: utils.RequireEnvSecret("DB_PASSWORD"),
			Db:       utils.RequireEnvString("DB_DB"),
		},
	}

	return config
}

func getCssFileName() string {
	files, err := rpFiles.Static.ReadDir("static/css")
	if err != nil {
		logger.Fatal("Failed to find static folder")
	}

	// Get first ".css" file
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".css") && strings.HasPrefix(file.Name(), "pages") {
			return file.Name()
		}
	}

	logger.Fatal("Didn't found a file named 'pages*.css' inside static directory")
	return ""
}
