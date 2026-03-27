package database

import (
	"context"
	"database/sql"
	"fmt"

	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"git.rpjosh.de/RPJosh/workout/pkg/utils"
	"github.com/RPJoshL/go-logger"
)

// Make sure that SQL implements SqlConnection
var _ SqlConnection = &DB{}
var _ SqlTransaction = &DBTransaction{}
var _ SqlConnection = &TestDB{}
var _ SqlTransaction = &TestDBTransaction{}

// SqlConnection is a generic interface wrapper for a sql.DB connection.
// The underlaying object should either be a *sql.DB or *sql.Transaction.
type SqlConnection interface {

	// Query executes a query that returns rows, typically a SELECT.
	// The args are for any placeholder parameters in the query.
	//
	// Query uses [context.Background] internally; to specify the context, use
	// [DB.QueryContext].
	Query(query string, args ...any) (*sql.Rows, error)

	// Exec executes a query without returning any rows.
	// The args are for any placeholder parameters in the query.
	//
	// Exec uses [context.Background] internally; to specify the context, use
	// [DB.ExecContext].
	Exec(query string, args ...any) (sql.Result, error)

	// BeginTransaction starts a new transaction for the db.
	// You cannot nest transactions!
	BeginTransaction() (SqlTransaction, error)

	// GetDb returns the underlaying database connection.
	// You should NOT USE THIS for any queries or statements!
	GetDb() *sql.DB
}

// SqlTransaction is a wrapper around a *sql.Tx connection object.
// It provides additional features like rolling or committing a transaction
type SqlTransaction interface {
	SqlConnection

	// Commit commits the transaction.
	Commit() error

	// Rollback aborts the transaction.
	Rollback() error
}

// DB is a wrapper around a default *sql.DB that implements the
// SqlConnection interface
type DB struct {
	*sql.DB
}

func (d *DB) GetDb() *sql.DB {
	return d.DB
}

// DBTransaction is a wrapper around a *sql.Tx transaction
type DBTransaction struct {
	*sql.Tx

	db *sql.DB
}

func (d *DB) BeginTransaction() (SqlTransaction, error) {
	tx, err := d.Begin()
	if err != nil {
		return nil, err
	} else {
		// Create a unique
		return &DBTransaction{
			Tx: tx,
			db: d.DB,
		}, nil
	}
}

func (d *DBTransaction) BeginTransaction() (SqlTransaction, error) {
	logger.Error("Nested transaction are not supported. Only use this function for tests!")

	return nil, errors.New("nested transactions are not supported")
}
func (d *DBTransaction) GetDb() *sql.DB {
	return d.db
}

// TestDB is a wrapper around a *sql.DB that creates a
// database connection for unit testing.
// It's fully based on transactions. Every "databaseConnection"
// created by "NewTestDB" is fully independent to each other
type TestDB struct {
	// Internal field. Don't use it
	*sql.Tx

	db *sql.DB
}

// TestDBTransaction is a wrapper around a *sql.Tx transaction
type TestDBTransaction struct {
	*sql.Tx

	db *sql.DB

	// Identifier of the savepoint created before
	// the transaction was started
	identifier string

	// reference to db struct
	dbTest *TestDB
}

// NewTestDB creates a new independent connection to the database
func NewTestDB(db *sql.DB) (*TestDB, error) {
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		// We use a high isolation level to not read changes made
		// in another transaction (which shouldn't be the case either)
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, err
	}

	return &TestDB{
		db: db,
		Tx: tx,
	}, nil
}

// Close should be called by your unit tests to destroy and cleanup
// the underlaying transaction and the database
func (d *TestDB) Close() {
	// Always rollback the transaction
	if err := d.Rollback(); err != nil {
		logger.Debug("Failed to close transaction of TestDB: %s", err)
	}
	// Don't close db connection
	// if err := d.db.Close(); err != nil {
	//	logger.Debug("Failed to close db of TestDB: %s", err)
	//}
}

func (d *TestDB) GetDb() *sql.DB {
	return d.db
}

func (d *TestDB) BeginTransaction() (SqlTransaction, error) {
	rtc := &TestDBTransaction{
		db:     d.db,
		Tx:     d.Tx,
		dbTest: d,
	}

	// To create "nested" transactions, we use the "SAVEPOINT" feature of the db
	rtc.identifier, _ = utils.GenerateRandomString(12)

	if _, err := d.Exec("SAVEPOINT " + rtc.identifier); err != nil {
		logger.Warning("Failed to create savepoint for TestDB: %s", err)
		return nil, errors.New("failed to create savepoint")
	}

	return rtc, nil
}

func (d *TestDBTransaction) BeginTransaction() (SqlTransaction, error) {
	return d.dbTest.BeginTransaction()
}

func (d *TestDBTransaction) GetDb() *sql.DB {
	return d.db
}

func (d *TestDBTransaction) Commit() error {
	// We commit nothing, because we are using a single transaction for tests only
	return nil
}

func (d *TestDBTransaction) Rollback() error {
	// Rollback to the previously created savepoint
	_, err := d.Exec("ROLLBACK TO SAVEPOINT " + d.identifier)
	return err
}

// CreateTable creates a table with the provided column configuration via
// an ddl statement and returns the table name.
// The table name is random and generated by this function
func CreateTable(db SqlConnection, statement string) (string, error) {
	name, _ := utils.GenerateRandomString(8)
	name = "ddl_test_" + name
	return CreateTableWithName(db, statement, name)
}

// CreateTableWithName creates a table withe the provided name
// via an ddl statement and returns the table name
func CreateTableWithName(db SqlConnection, statement, name string) (string, error) {
	sqll := fmt.Sprintf("CREATE TABLE %s (%s)", name, statement)

	_, err := db.Exec(sqll)
	if err != nil {
		logger.Debug("Create statement: %s", sqll)
	}
	return name, err
}
func DropTable(db SqlConnection, tableName string) error {
	_, err := db.Exec("DROP TABLE " + tableName)
	return err
}
