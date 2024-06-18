package main

import (
	"os"

	"git.rpjosh.de/RPJosh/go-logger"
	"gopkg.in/yaml.v3"
)

type Config struct {
	User   UserConfig   `yaml:"user"`
	App    AppConfig    `yaml:"app"`
	Logger LoggerConfig `yaml:"logger"`
}

type UserConfig struct {
	Name     string `yaml:"name"`
	Password string `yaml:"password"`
}
type AppConfig struct {
	Url       string `yaml:"url"`
	Directory string `yaml:"directory"`
}
type LoggerConfig struct {
	Level string `yaml:"level"`
}

// GetConfig returns the main configuration of the app.
// It panics if an invalid or missing conifiguration was found
func GetConfig(path string) Config {
	rtc := Config{}

	fileContent, err := os.ReadFile(path)
	if err != nil {
		logger.Fatal("Failed to read configuration file: %s", err)
	}

	if err := yaml.Unmarshal(fileContent, &rtc); err != nil {
		logger.Fatal("Failed to parse configuration file %q: %s", path, err)
	}

	// Check for required variables
	if rtc.App.Url == "" || rtc.App.Directory == "" || rtc.User.Name == "" || rtc.User.Password == "" {
		logger.Fatal("User and app details are required")
	}

	// Configure logger
	level := logger.LevelInfo
	if rtc.Logger.Level != "" {
		level = logger.GetLevelByName(rtc.Logger.Level)
	}
	logger.SetGlobalLogger(logger.GetLoggerFromEnv(&logger.Logger{
		ColoredOutput: true,
		Level:         level,
		PrintSource:   true,
		File:          &logger.FileLogger{},
	}))

	return rtc
}
