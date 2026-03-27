// Package testsExtra is a duplicate of the package [test] to avoid an import_cycle when
// your tested package is used inside router
package testsExtra

import (
	"database/sql"
	"fmt"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/models"
	"github.com/RPJoshL/go-logger"
	_ "github.com/go-sql-driver/mysql"
)

// Makes sure that a valid logger configuration is set
//
//nolint:gochecknoinits // Only for test packages
func init() {
	models.SetLoggerConfig()
}

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
