package database

import (
	"database/sql"
	"fmt"
	"reflect"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

var _ Dbler = &Utils{}

// Utils contains a "sql.db" connection
// bundled with helper functions to make your work
// with the db easier
type Utils struct {

	// mainDb always represents the database that was provided
	// during initialization of the database utils.
	mainDb SqlConnection

	// Db is the current database connection that is used for all opertions.
	// This can either be a [SqlConnection] or [SqlTransaction]
	Db SqlConnection

	// IsTransaction states weather this databaseUtils is used as a transaction
	IsTransaction bool
}

// Dbler is an interface for generic methods of [Utils]. It allows you
// to overwrite the concrete type that is expected from other database tools
// like [dbstruct.NewOperator]
type Dbler interface {
	DB() SqlConnection
	NewTransactionInt() (Dbler, error)
	CommitTransaction() error
	RollbackTransaction() error
}

// NewUtils initializes a new instance of the utils
func NewUtils(db *sql.DB) *Utils {
	rtc := &Utils{
		Db:     &DB{DB: db},
		mainDb: &DB{DB: db},
	}

	return rtc
}

func NewUtilsByDb(db SqlConnection) *Utils {
	rtc := &Utils{
		Db:     db,
		mainDb: db,
	}

	return rtc
}

// isPointer returns weather "val" is a pointer to the given type
func isPointer(ref reflect.Value, typ reflect.Kind) error {
	if ref.Type().Kind() != reflect.Pointer {
		return errors.New("no pointer")
	} else if ref.IsNil() {
		return errors.New("nil pointer")
	} else if ref.Elem().Type().Kind() != typ {
		return fmt.Errorf("no %s", typ.String())
	}

	return nil
}

func (d *Utils) DB() SqlConnection {
	return d.Db
}

// NewTransaction creates a new instance of the DatabaseUtils
// that uses the created transaction
func (d *Utils) NewTransaction() (*Utils, error) {
	rtc := &Utils{
		mainDb:        d.Db,
		IsTransaction: true,
	}

	// Create transaction
	tx, err := d.Db.BeginTransaction()
	if err != nil {
		return nil, err
	}
	rtc.Db = tx

	return rtc, nil
}

func (d *Utils) NewTransactionInt() (Dbler, error) {
	return d.NewTransaction()
}

// CommitTransaction commits the current transaction if
// [isTransaction] is true.
// You can now longer use this instance of DatabaseUtils
// afterwards!
func (d *Utils) CommitTransaction() error {
	if !d.IsTransaction {
		logger.Warning("Called CommitTransaction() on no transcation")
		return errors.New("no transaction")
	}

	// Commit
	trans := d.Db.(SqlTransaction)
	d.IsTransaction = false
	return trans.Commit()
}

// RollbackTransaction rolls the current transaction back if
// [isTransaction] is true.
// You can now longer use this instance of DatabaseUtils
// afterwards!
func (d *Utils) RollbackTransaction() error {
	if !d.IsTransaction {
		logger.Warning("Called CommitTransaction() on no transcation")
		return errors.New("no transaction")
	}

	// Commit
	trans := d.Db.(SqlTransaction)
	d.IsTransaction = false
	err := trans.Rollback()
	if err != nil {
		logger.Warning("Failed to rollback transaction: %s", err)
	}

	return err
}

// QueryStruct executes the query and writes the result into *dst.
//
// Exactly a single row is expected to be returned from the database
func (d *Utils) QueryStruct(dst any, sql string, params ...any) Error {
	// Validate given type
	dstVal := reflect.ValueOf(dst)
	if err := isPointer(dstVal, reflect.Struct); err != nil {
		return DatabaseError{
			Typ:      UnexpectedError,
			Err:      fmt.Errorf("invalid type for dst given: %w", err),
			Response: errors.InternalError(),
		}
	}

	// Execute the statement
	rows, err := d.Db.Query(sql, params...)
	if err != nil {
		logger.Warning("Query for struct failed: %s", err)
		return DatabaseError{
			Typ:      UnexpectedError,
			Err:      fmt.Errorf("failed to query value: %w", err),
			Response: errors.InternalError(),
		}
	}
	defer rows.Close()

	// Fetch the next (and single) row
	columnNames, _ := rows.Columns()
	if rows.Next() {
		columns := mappDbColumns(dstVal, columnNames)

		// Scan the data into the struct
		if err := rows.Scan(columns...); err != nil {
			logger.Error("Query error for db: %s", err)
			return DatabaseError{
				Typ:      UnexpectedError,
				Err:      fmt.Errorf("failed to scan row: %w", err),
				Response: errors.InternalError(),
			}
		}
	} else if err := rows.Err(); err != nil {
		logger.Error("Query error for db (while getting next result): %s", err)
		return DatabaseError{
			Typ:      UnexpectedError,
			Err:      fmt.Errorf("failed to scan row: %w", err),
			Response: errors.InternalError(),
		}
	} else {
		return DatabaseError{
			Typ:      NoRows,
			Err:      errors.New("no data found in select"),
			Response: errors.NewError("No data found", 404),
		}
	}

	// Are there any remaining rows?
	if rows.Next() {
		// Get the count of them for debug purporses
		counter := 2
		for rows.Next() {
			counter++
		}

		return DatabaseError{
			Typ:      TooManyRows,
			Err:      fmt.Errorf("found %d rows instead of a single one", counter),
			Response: errors.NewError("Too many data found", 409),
		}
	}

	return nil
}

// QueryStructs executes the query and writes the result into the given array (*dst) of
// structs
func (d *Utils) QueryStructs(dst any, sql string, params ...any) Error {
	// Make sure that we got a slice
	dstType := reflect.TypeOf(dst)
	if dstType.Kind() != reflect.Pointer || dstType.Elem().Kind() != reflect.Slice || dstType.Elem().Elem().Kind() != reflect.Struct {
		return DatabaseError{
			Typ:      UnexpectedError,
			Err:      errors.New("expected a pointer to a slice containing structs for dst"),
			Response: errors.InternalError(),
		}
	}
	dstType = dstType.Elem().Elem()
	dstRef := reflect.ValueOf(dst).Elem()

	// Execute the sql statement
	rows, err := d.Db.Query(sql, params...)
	if err != nil {
		logger.Warning("Query for struct failed: %s", err)
		return DatabaseError{
			Typ:      UnexpectedError,
			Err:      fmt.Errorf("failed to query value: %w", err),
			Response: errors.InternalError(),
		}
	}
	defer rows.Close()

	// Fetch all rows into dst
	columnNames, _ := rows.Columns()
	for rows.Next() {
		dbRow := reflect.New(dstType)
		columns := mappDbColumns(dbRow, columnNames)

		// Scan elements
		if err := rows.Scan(columns...); err != nil {
			logger.Error("Query error for db: %s", err)
			return DatabaseError{
				Typ:      UnexpectedError,
				Err:      fmt.Errorf("failed to scan row: %w", err),
				Response: errors.InternalError(),
			}
		} else {
			dstRef.Set(reflect.Append(dstRef, dbRow.Elem()))
		}
	}

	if err := rows.Err(); err != nil {
		logger.Warning("Query for struct failed (while iterating): %s", err)
		return DatabaseError{
			Typ:      UnexpectedError,
			Err:      fmt.Errorf("failed to query value: %w", err),
			Response: errors.InternalError(),
		}
	}

	return nil
}

// QueryForValue will do a DB select which expects a single raw value.
// The value is written to dst.
//
// For structs, you should use the method "QueryForModel" or .Structs.Query
func (d *Utils) QueryForValue(dst any, sql string, params ...any) Error {
	// Execute select
	rows, err := d.Db.Query(sql, params...)
	if err != nil {
		logger.Error("Query error for db: %s", err)
		return DatabaseError{
			Typ:      UnexpectedError,
			Err:      fmt.Errorf("failed to query value: %w", err),
			Response: errors.InternalError(),
		}
	}
	defer rows.Close()

	// Fetch the next (and single) row
	if rows.Next() {
		err = rows.Scan(dst)
		if err != nil {
			logger.Error("Query error for db: %s", err)
			return DatabaseError{
				Typ:      UnexpectedError,
				Err:      fmt.Errorf("failed to scan row: %w", err),
				Response: errors.InternalError(),
			}
		}
	} else if err := rows.Err(); err != nil {
		logger.Error("Query error for db (while iterating): %s", err)
		return DatabaseError{
			Typ:      UnexpectedError,
			Err:      fmt.Errorf("failed to query value: %w", err),
			Response: errors.InternalError(),
		}
	} else {
		return DatabaseError{
			Typ:      NoRows,
			Err:      errors.New("no data found in select"),
			Response: errors.NewError("No data found", 404),
		}
	}

	// Are there any remaining rows?
	if rows.Next() {
		// Get the count of them for debug purporses
		counter := 2
		for rows.Next() {
			counter++
		}

		return DatabaseError{
			Typ:      TooManyRows,
			Err:      fmt.Errorf("found %d rows instead of a single one", counter),
			Response: errors.NewError("Too many data found", 409),
		}
	}

	return nil
}
