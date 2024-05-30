package database

import (
	"database/sql"
	"fmt"
	"reflect"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

// DatabaseUtils contains a "sql.db" connection
// bundled with helper functions to make your work
// with the db easier
type DatabaseUtils struct {

	// mainDb always represents the database that was provided
	// during initialization of the database utils.
	mainDb SqlConnection

	// Db is the current database connection that is used for all opertions.
	// This can either be a [SqlConnection] or [SqlTransaction]
	Db SqlConnection

	// IsTransaction states weather this databaseUtils is used as a transaction
	IsTransaction bool

	Struct *StructOperator
}

// NewDatabaseUtils initializes a new instance of the utils
func NewDatabaseUtils(db *sql.DB) *DatabaseUtils {
	rtc := &DatabaseUtils{
		Db:     &DB{DB: db},
		mainDb: &DB{DB: db},
	}
	rtc.Struct = &StructOperator{
		dbUtils: rtc,
	}

	return rtc
}

func NewDatabaseUtilsByDb(db SqlConnection) *DatabaseUtils {
	rtc := &DatabaseUtils{
		Db:     db,
		mainDb: db,
	}
	rtc.Struct = &StructOperator{
		dbUtils: rtc,
	}

	return rtc
}

// isPointer returns weather "val" is a pointer to the given type
func isPointer(ref reflect.Value, typ reflect.Kind) error {
	if ref.Type().Kind() != reflect.Pointer {
		return fmt.Errorf("no pointer")
	} else if ref.IsNil() {
		return fmt.Errorf("nil pointer")
	} else if ref.Elem().Type().Kind() != typ {
		return fmt.Errorf("no %s", typ.String())
	}

	return nil
}

// isPointerType returns weather "val" is a pointer to the given type
func isPointerType(ref reflect.Type, typ reflect.Kind) error {
	if ref.Kind() != reflect.Pointer {
		return fmt.Errorf("no pointer")
	} else if ref.Elem().Kind() != typ {
		return fmt.Errorf("no %s", typ.String())
	}

	return nil
}

// NewTransaction creates a new instance of the DatabaseUtils
// that uses the created transaction
func (d *DatabaseUtils) NewTransaction() (*DatabaseUtils, error) {
	rtc := &DatabaseUtils{
		mainDb:        d.Db,
		IsTransaction: true,
	}
	rtc.Struct = &StructOperator{dbUtils: rtc}

	// Create transaction
	tx, err := d.Db.BeginTransaction()
	if err != nil {
		return nil, err
	}
	rtc.Db = tx

	return rtc, nil
}

// CommitTransaction commits the current transaction if
// [isTransaction] is true.
// You can now longer use this instance of DatabaseUtils
// afterwards!
func (d *DatabaseUtils) CommitTransaction() error {
	if !d.IsTransaction {
		logger.Warning("Called CommitTransaction() on no transcation")
		return fmt.Errorf("no transaction")
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
func (d *DatabaseUtils) RollbackTransaction() error {
	if !d.IsTransaction {
		logger.Warning("Called CommitTransaction() on no transcation")
		return fmt.Errorf("no transaction")
	}

	// Commit
	trans := d.Db.(SqlTransaction)
	d.IsTransaction = false
	return trans.Rollback()
}

// QueryStruct executes the query and writes the result into *dst.
//
// Exactly a single row is expected to be returned from the database
func (d *DatabaseUtils) QueryStruct(dst any, sql string, params ...any) DatabaseError {

	// Validate given type
	dstVal := reflect.ValueOf(dst)
	if err := isPointer(dstVal, reflect.Struct); err != nil {
		return databaseErr{
			Typ:      UnexpectedError,
			Err:      fmt.Errorf("invalid type for dst given: %s", err),
			Response: errors.InternalError(),
		}
	}

	// Execute the statement
	rows, err := d.Db.Query(sql, params...)
	if err != nil {
		logger.Warning("Query for struct failed: %s", err)
		return databaseErr{
			Typ:      UnexpectedError,
			Err:      fmt.Errorf("failed to query value: %s", err),
			Response: errors.InternalError(),
		}
	}
	defer rows.Close()

	// Fetch the next (and single) row
	columnNames, _ := rows.Columns()
	if rows.Next() {
		columns := d.mappDbColumns(dstVal, columnNames)

		// Scan the data into the struct
		if err := rows.Scan(columns...); err != nil {
			logger.Error("Query error for db: %s", err)
			return databaseErr{
				Typ:      UnexpectedError,
				Err:      fmt.Errorf("failed to scan row: %s", err),
				Response: errors.InternalError(),
			}
		}
	} else {
		return databaseErr{
			Typ:      NoRows,
			Err:      fmt.Errorf("no data found in select"),
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

		return databaseErr{
			Typ:      TooManyRows,
			Err:      fmt.Errorf("found %d rows instead of a single one", counter),
			Response: errors.NewError("Too many data found", 409),
		}
	}

	return nil
}

// QueryStructs executes the query and writes the result into the given array (*dst) of
// structs
func (d *DatabaseUtils) QueryStructs(dst any, sql string, params ...any) DatabaseError {

	// Make sure that we got a slice
	dstType := reflect.TypeOf(dst)
	if dstType.Kind() != reflect.Pointer || dstType.Elem().Kind() != reflect.Slice || dstType.Elem().Elem().Kind() != reflect.Struct {
		return databaseErr{
			Typ:      UnexpectedError,
			Err:      fmt.Errorf("expected a pointer to a slice containing structs for dst"),
			Response: errors.InternalError(),
		}
	}
	dstType = dstType.Elem().Elem()
	dstRef := reflect.ValueOf(dst).Elem()

	// Execute the sql statement
	rows, err := d.Db.Query(sql, params...)
	if err != nil {
		logger.Warning("Query for struct failed: %s", err)
		return databaseErr{
			Typ:      UnexpectedError,
			Err:      fmt.Errorf("failed to query value: %s", err),
			Response: errors.InternalError(),
		}
	}
	defer rows.Close()

	// Fetch all rows into dst
	columnNames, _ := rows.Columns()
	for rows.Next() {
		dbRow := reflect.New(dstType)
		columns := d.mappDbColumns(dbRow, columnNames)

		// Scan elements
		if err := rows.Scan(columns...); err != nil {
			logger.Error("Query error for db: %s", err)
			return databaseErr{
				Typ:      UnexpectedError,
				Err:      fmt.Errorf("failed to scan row: %s", err),
				Response: errors.InternalError(),
			}
		} else {
			dstRef.Set(reflect.Append(dstRef, dbRow.Elem()))
		}
	}

	return nil
}

// QueryForValue will do a DB select which expects a single raw value.
// The value is written to dst.
//
// For structs, you should use the method "QueryForModel" or .Structs.Query
func (d *DatabaseUtils) QueryForValue(dst any, sql string, params ...any) DatabaseError {

	// Execute select
	rows, err := d.Db.Query(sql, params...)
	if err != nil {
		logger.Error("Query error for db: %s", err)
		return databaseErr{
			Typ:      UnexpectedError,
			Err:      fmt.Errorf("failed to query value: %s", err),
			Response: errors.InternalError(),
		}
	}
	defer rows.Close()

	// Fetch the next (and single) row
	if rows.Next() {
		err = rows.Scan(dst)
		if err != nil {
			logger.Error("Query error for db: %s", err)
			return databaseErr{
				Typ:      UnexpectedError,
				Err:      fmt.Errorf("failed to scan row: %s", err),
				Response: errors.InternalError(),
			}
		}
	} else {
		return databaseErr{
			Typ:      NoRows,
			Err:      fmt.Errorf("no data found in select"),
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

		return databaseErr{
			Typ:      TooManyRows,
			Err:      fmt.Errorf("found %d rows instead of a single one", counter),
			Response: errors.NewError("Too many data found", 409),
		}
	}

	return nil
}
