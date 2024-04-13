package tests

import (
	"database/sql"

	"git.rpjosh.de/RPJosh/workout/internal/api"
	"git.rpjosh.de/RPJosh/workout/internal/database"
	"git.rpjosh.de/RPJosh/workout/internal/models"
)

// GetDb returns a database connection to the MySQL / MariaDB
// database
func GetDb() *sql.DB {
	// Get the generic configuration of the app
	conf := models.GetAppConfig()
	api := &api.Api{
		Config: conf,
	}

	return api.GetDb()
}

func GetDbConnection() database.SqlConnection {
	return &database.DB{
		DB: GetDb(),
	}
}
