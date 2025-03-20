package dbstruct

import (
	"fmt"
	"reflect"
	"strings"

	"git.rpjosh.de/RPJosh/go-ddl-parser/structt"
	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/pkg/database"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

// Update statement in a database context
type Update struct {

	// Internal reference to an operator
	operator *Operator

	// Fields to exclude to query
	columnSelector ColumnSelector

	// Whether the selector was set
	customSelector bool

	// Type of the struct to insert
	typ reflect.Type

	// Value to insert (slice of types)
	insertVal reflect.Value

	// The last occured error will be stored in this fild and is
	// only returned in "Run()".
	// If any error occured, the (internal) processing is stopped
	err database.Error
}

// Update will update a single row of the database.
// A pointer to a struct is expected (*struct).
// Embedded structs are not supported!
func (o *Operator) Update(val any) *Update {
	rtc := &Update{
		operator: o,
	}

	// Check for pointer
	if !rtc.columnSelector.isTableStruct(reflect.TypeOf(val)) {
		rtc.err = database.DatabaseError{
			Typ:      database.UnexpectedError,
			Err:      fmt.Errorf("invalid type for insert value given. Expected a pointer to a table struct: %s", reflect.TypeOf(val)),
			Response: errors.InternalError(),
		}
		return rtc
	}

	rtc.typ = reflect.TypeOf(val).Elem()
	rtc.insertVal = reflect.MakeSlice(reflect.SliceOf(rtc.typ), 0, 0)
	rtc.insertVal = reflect.Append(rtc.insertVal, reflect.ValueOf(val).Elem())

	return rtc
}

// UpdateSlice will update a list of rows in the database.
// A pointer to an array is expected (*[]struct).
// Embedded structs are not supported!
func (o *Operator) UpdateSlice(val any) *Update {
	rtc := &Update{
		operator: o,
	}

	// Check for pointer
	t := reflect.TypeOf(val)
	if isPointerType(t, reflect.Slice) != nil || !rtc.columnSelector.isArrayToTableStruct(t.Elem()) {
		rtc.err = database.DatabaseError{
			Typ:      database.UnexpectedError,
			Err:      fmt.Errorf("invalid type for insert value given. Expected a pointer to an array of table structs: %s", t),
			Response: errors.InternalError(),
		}
		return rtc
	}

	rtc.typ = t.Elem().Elem()
	rtc.insertVal = reflect.ValueOf(val).Elem()

	return rtc
}

// Selector sets a custom selector for columns which should be inserted
func (q *Update) Selector(selector ColumnSelector) *Update {
	q.columnSelector = selector
	q.customSelector = true
	return q
}

// Run executes the insert operation and returns all errors
func (q *Update) Run() database.Error {
	if q.err != nil {
		return q.err
	}

	// A primary key is needed to identify the row to update
	q.columnSelector.includePrimaryKeys = true

	// Parse all fields. We expect a single table (level != 0)
	tbls, err := q.columnSelector.parseField(q.typ, 1, "")
	if err != nil {
		return database.DatabaseError{
			Typ:      database.UnexpectedError,
			Err:      fmt.Errorf("failed to parse fields of struct %q: %s", q.typ, err),
			Response: errors.InternalError(),
		}
	}
	if len(tbls) < 1 {
		return database.DatabaseError{
			Typ:      database.UnexpectedError,
			Err:      fmt.Errorf("no table received form parsing struct"),
			Response: errors.InternalError(),
		}
	}

	// Nothing to insert
	if q.insertVal.Len() == 0 {
		logger.Debug("Got no data to insert for table %q", getTableIdentifier(&tbls[0]))
		return nil
	}

	// Insert 1:1 references
	if q.columnSelector.ForeignKeyReference {
		if err := insert1To1Reference(tbls, q.operator, q.insertVal); err != nil {
			return err
		}
	}

	// We expect exactly a primary key
	primKeys := 0
	primKeyJoin := ""
	for _, c := range tbls[0].columns {
		if c.IsPrimaryKey {
			primKeys++
			primKeyJoin = fmt.Sprintf("ON vals.%s = tbl.%s", c.Name, c.Name)
		}
	}
	if primKeys != 1 {
		return database.DatabaseError{
			Typ:      database.UnexpectedError,
			Err:      fmt.Errorf("expected to find exactly a single primary key. Found %d", primKeys),
			Response: errors.InternalError(),
		}
	}

	// Build update statement with values
	update := "UPDATE " + getTableIdentifier(&tbls[0]) + " tbl \nJOIN ("
	placeholders := make([]any, 0)
	for rowI := 0; rowI < q.insertVal.Len(); rowI++ {
		ii := 0
		for colI, col := range tbls[0].columns {
			// Syntax preperatiosn
			if colI == 0 {
				if rowI != 0 {
					update += "\n\tUNION ALL"
				}
				update += "\n\tSELECT "
			}

			// Skip pointed key reference
			if col.PointedKeyReference != "" {
				continue
			}

			if ii != 0 {
				update += ", "
			}

			// Value to insert
			value := q.insertVal.Index(rowI).Field(col.position).Interface()

			// If we have a 1:1 reference (which we don't create), the value to insert
			// has to be extracted from the referenced struct
			if col.isForeignKeyReference {
				// Get the referenced column position
				lastDot := strings.LastIndex(col.ForeignKeyReference, ".")
				referencedColumn := col.ForeignKeyReference[lastDot+1:]
				position := -1
				for _, cc := range col.foreignKeyTable.columns {
					if cc.Name == referencedColumn {
						position = cc.position
					}
				}

				// Referenced field not found
				if position == -1 {
					return database.DatabaseError{
						Typ:      database.UnexpectedError,
						Err:      fmt.Errorf("didn't found referenced column %q in %s", referencedColumn, reflect.ValueOf(value).Type()),
						Response: errors.InternalError(),
					}
				}

				refValue := reflect.ValueOf(value)
				if refValue.Kind() == reflect.Pointer {
					if !refValue.IsValid() || refValue.IsZero() {
						value = nil
					} else {
						value = refValue.Elem().Field(position).Interface()
					}
				} else {
					value = refValue.Field(position).Interface()
				}
			}

			// We use a default value if the column is not nullable
			// and the user provided a "null" value (go-ddl only sets the field
			// to a sql.Null type if the column is not nullable)
			if col.HasDefaultValue && isZero(value) {
				update += fmt.Sprintf(
					`(SELECT a.def FROM (SELECT COUNT(*), DEFAULT(%s) AS "def" FROM %s) a) AS %q`,
					col.Name, getTableIdentifier(&tbls[0]), col.Name,
				)
			} else {
				update += fmt.Sprintf("? AS %q", col.Name)
				placeholders = append(placeholders, value)
			}

			ii = ii + 1
		}
	}

	// Add join and set statements
	update += "\n) vals " + primKeyJoin + "\n SET "
	ii := 0
	for _, c := range tbls[0].columns {
		if c.IsPrimaryKey || c.PointedKeyReference != "" {
			continue
		}

		// Add colon
		if ii != 0 {
			update += ", "
		}

		// Add set statement
		update += fmt.Sprintf("tbl.%s = vals.%s", c.Name, c.Name)

		ii++
	}

	// Execute the update
	_, err = q.operator.dbUtils.DB().Exec(update, placeholders...)
	if err != nil {
		logger.Debug("Statement for failed update:\n%s", update)
		return database.DatabaseError{
			Typ:      database.UnexpectedError,
			Err:      fmt.Errorf("failed to update value: %s", err),
			Response: errors.InternalError(),
		}
	}

	// Add n:1 relationships (Delete + Insert again)
	if q.columnSelector.PointedKeyReference {

		// Use transactions if something fails
		transaction, err := q.operator.dbUtils.NewTransactionInt()
		if err != nil {
			return database.DatabaseError{
				Typ:      database.UnexpectedError,
				Err:      fmt.Errorf("failed to create transaction for 1:n reference %s", err),
				Response: errors.InternalError(),
			}
		}

		// Find columns we need to update
		for _, col := range tbls[0].columns {
			if col.PointedKeyReference == "" {
				continue
			}

			// Find the referenced field
			coll := column{}
			for _, cc := range col.foreignKeyTable.columns {
				if getColumnIdentifier(col.foreignKeyTable, &cc) == col.PointedKeyReference {
					coll = cc
					break
				}
			}
			if coll.fieldName == "" || coll.ForeignKeyReference == "" {
				return database.DatabaseError{
					Typ:      database.UnexpectedError,
					Err:      fmt.Errorf("no referenced field found for %q in %q", col.PointedKeyReference, col.foreignKeyTable.typ),
					Response: errors.InternalError(),
				}
			}

			// Get table and column from where to delete the rows from
			delColumnName := col.PointedKeyReference
			delLastPoint := strings.LastIndex(delColumnName, ".")
			delTableName := col.PointedKeyReference[:delLastPoint]
			delStatement := fmt.Sprintf("DELETE FROM %s WHERE %s IN (", delTableName, delColumnName)
			delPlaceholder := []any{}

			// Extract the field name to which the foreign key points to
			fieldName := coll.ForeignKeyReference
			lastPoint := strings.LastIndex(fieldName, ".")
			fieldName = structt.GetFieldName(fieldName[lastPoint+1:])

			// Values to insert again after deleting the old ones
			//insVal := reflect.MakeSlice(reflect.SliceOf(col.foreignKeyTable.typ), 0, 0)

			// Build delete statement
			for rowI := 0; rowI < q.insertVal.Len(); rowI++ {
				identifier := q.insertVal.Index(rowI).FieldByName(fieldName).Interface()
				if rowI != 0 {
					delStatement += ","
				}
				delStatement += "?"
				delPlaceholder = append(delPlaceholder, identifier)

				//insVal = reflect.Append(insVal, q.insertVal.Index(rowI).FieldByName(col.fieldName))
			}
			delStatement += ")"

			// Delete values
			res, err := transaction.DB().Exec(delStatement, delPlaceholder...)
			if err != nil {
				logger.Debug("Failed delete statement:\n%s", delStatement)
				transaction.RollbackTransaction()
				return database.DatabaseError{
					Typ:      database.UnexpectedError,
					Err:      fmt.Errorf("failed to delete 1:n reference: %s", err),
					Response: errors.InternalError(),
				}
			}
			affectedRows, _ := res.RowsAffected()
			logger.Trace("Deleted %d rows from %q for 1:n update (for field %q)", affectedRows, delTableName, col.fieldName)

			// Insert data again
			//ins := q.operator.InsertSlice(insVal.Addr().Interface())
			ins := q.operator.insertSlice(q.insertVal)
			tmpOperator := *ins.operator
			ins.operator = &tmpOperator
			ins.operator.dbUtils = transaction

			// Only update the field with the 1:n relationship
			//ins.Selector(ColumnSelector{ IncludeColumns: []string{ col.fieldName + "|#" + structt.GetFieldName("") } })
			ins.Selector(ColumnSelector{IncludeColumns: []string{"*|" + delTableName}, PointedKeyReference: true})
			ins.columnSelector.includePrimaryKeys = true
			if _, err := ins.Run(); err != nil {
				transaction.RollbackTransaction()
				return err
			}

			// Build array with data we need to insert

			/*
				insArray := reflect.MakeSlice(reflect.SliceOf(col.foreignKeyTable.typ), 0, 0)

				// If we insert multiple rows at once with AUTO_INCREMENT, we also have multiple
				// primary keys. For MariaDB, they are incremented ALWAYS by once
				insIdCols := insId
				for rowI := 0; rowI < q.insertVal.Len(); rowI++ {
					field := q.insertVal.Index(rowI).Field(col.position)

					// Get identifier of the row we need to set for the foreign key.
					// This HAS TO BE the primary key of the table → use auto_increment
					// if last inserted ID it not zero
					identifier := q.insertVal.Index(rowI).FieldByName(fieldName)
					if insIdCols != 0 {
						identifier = reflect.ValueOf(int(insIdCols))
						insIdCols++
					}

					// Set this identifier for every element
					for i := 0; i < field.Len(); i++ {
						sF := field.Index(i).FieldByName(coll.fieldName)
						sF.Set(identifier)
						insArray = reflect.Append(insArray, field.Index(i))
					}
				}

				// Copy this insert struct to apply customizations
				qCopy := *q
				qCopy.typ = insArray.Type().Elem()
				qCopy.insertVal = insArray
				_, errNew := qCopy.Run()
				if errNew != nil {
					return 0, databaseErr{
						Typ:      UnexpectedError,
						Err:      fmt.Errorf("failed to insert pointed key reference %s: %s", qCopy.typ, errNew),
						Response: errors.InternalError(),
					}
				}
			*/
		}

		transaction.CommitTransaction()
	}

	return nil
}
