package tests

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/database"
	"github.com/RPJoshL/go-logger"
	_ "github.com/go-sql-driver/mysql"
)

// GetDb returns a database connection to the MySQL / MariaDB
// database
func GetDb() *sql.DB {
	// Get the generic configuration of the app
	conf := models.GetDbConfig()

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", conf.User, conf.Password, conf.Address, conf.Db))
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

// CommitDb commits the current database transaction
func CommitDb(t *testing.T, dbR *database.Utils, exit bool) {
	if db, ok := dbR.Db.(*database.TestDB); ok {
		if err := db.Commit(); err != nil {
			t.Fatalf("%s", "Failed to commit test db: "+err.Error())
		}

		dbR.Db, _ = database.NewTestDB(db.GetDb())

		if exit {
			t.Fatal("Committed to test database")
		} else {
			logger.Warning("Committed to test database")
		}
	} else {
		logger.Error("No test database found")
	}
}
