package tests

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/database"
	_ "github.com/go-sql-driver/mysql"
)

// GetDb returns a database connection to the MySQL / MariaDB
// database
func GetDb() *sql.DB {
	// Get the generic configuration of the app
	conf := models.GetAppConfig()

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", conf.Db.User, conf.Db.Password, conf.Db.Address, conf.Db.Db))
	if err != nil {
		logger.Fatal("Failed to open DB connection: %s", err)
	}

	// Set performance setttings
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(6)
	db.SetMaxIdleConns(6)

	return db
}

// GetDbConnection returns a test db connection
func GetDbConnection(t *testing.T) database.SqlConnection {
	db, err := database.NewTestDB(GetDb())
	if err != nil {
		logger.Fatal("Failed to create connection to test database: %s", err)
	}

	// Automatically rollback transaction when test is finished
	t.Cleanup(func() { _ = db.Rollback() })

	return db
}
