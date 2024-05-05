// Duplicated package to avoid import_cycle when your tested
// package is used inside router
package testsExtra

import (
	"database/sql"
	"fmt"
	"time"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/models"
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
