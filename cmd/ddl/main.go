package main

import (
	"os"

	"git.rpjosh.de/RPJosh/go-ddl-parser"
	"git.rpjosh.de/RPJosh/go-ddl-parser/structt"
	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/api"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"gopkg.in/yaml.v3"
)

func main() {
	defer logger.CloseFile()

	// Get the generic configuration of the app
	conf := models.GetAppConfig()
	api := &api.Api{
		Config: conf,
	}

	// Get all tables within current schema
	mariadb := ddl.NewMariaDb(api.GetDb())
	tables, err := mariadb.GetTables(conf.Db.Db)
	if err != nil {
		logger.Fatal("Failed to fetch tables from mariadb: %s", err)
	}

	// Get ddl configuration
	ddlConfig := &structt.StructConfig{}
	if err := parseDDLConfig(ddlConfig, "./cmd/ddl/ddl.yaml"); err != nil {
		logger.Fatal("Failed to parse struct configuration: %s", err)
	}

	// Generate it
	if err := structt.CreateStructs(ddlConfig, tables); err != nil {
		logger.Fatal("Failed to create structs: %s", err)
	}

}

// parseDDLConfig parses the given configuration file (.yaml file) to a StructConfig
func parseDDLConfig(conf *structt.StructConfig, file string) error {
	if file == "" {
		return nil
	}

	dat, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(dat, conf); err != nil {
		return err
	}

	return nil
}
