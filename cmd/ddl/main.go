package main

import (
	"database/sql"
	"os"
	"slices"

	"git.rpjosh.de/RPJosh/go-ddl-parser"
	"git.rpjosh.de/RPJosh/go-ddl-parser/structt"
	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/api"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/utils"
	"gopkg.in/yaml.v3"
)

type Config struct {
	structt.StructConfig `yaml:",inline"`

	IgnoreTables []string `yaml:"ignoreTables"`
}

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
	ddlConfig := &Config{}
	if err := parseDDLConfig(ddlConfig, "./cmd/ddl/ddl.yaml"); err != nil {
		logger.Fatal("Failed to parse struct configuration: %s", err)
	}

	// Filter tables
	for i := 0; i < len(tables); i++ {
		if slices.Contains(ddlConfig.IgnoreTables, tables[i].Name) {
			tables = utils.RemovePreserveOrder(&tables, i)
			i--
		}
	}

	// Use nullable types to improve JSON output
	ddlConfig.NullConfig = structt.NullConfig{
		Package: "github.com/guregu/null/v5",
		Prefix:  sql.NullString{Valid: true, String: "null."},
	}

	// Generate it
	if err := structt.CreateStructs(&ddlConfig.StructConfig, tables); err != nil {
		logger.Fatal("Failed to create structs: %s", err)
	}
}

// parseDDLConfig parses the given configuration file (.yaml file) to a StructConfig
func parseDDLConfig(conf *Config, file string) error {
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
