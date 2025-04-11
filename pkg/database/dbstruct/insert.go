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

// Insert statement in a database context
type Insert struct {

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

	// The last occurred error will be stored in this fild and is
	// only returned in "Run()".
	// If any error occurred, the (internal) processing is stopped
	err database.Error
}

// Insert will insert a single struct into the table defined within metadata.
// A pointer to a struct is expected (*struct).
// Embedded structs are not supported!
func (o *Operator) Insert(val any) *Insert {
	rtc := &Insert{
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

// InsertSlice will insert a list of structs into the table defined within metadata.
// A pointer to an array is expected (*[]struct).
// Embedded structs are not supported!
func (o *Operator) InsertSlice(val any) *Insert {
	rtc := &Insert{
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

// insertSlice is a wrapper for [InsertSlice] that
// accepts a [reflect.Value] instead of an interface
func (o *Operator) insertSlice(val reflect.Value) *Insert {
	rtc := &Insert{
		operator: o,
	}

	rtc.typ = val.Type().Elem()
	rtc.insertVal = val

	return rtc
}

// Selector sets a custom selector for columns which should be inserted
func (q *Insert) Selector(selector ColumnSelector) *Insert {
	q.columnSelector = selector
	q.customSelector = true
	return q
}

// Run executes the insert operation and returns all errors
// with the first inserted ID for an auto_increment column
func (q *Insert) Run() (int64, database.Error) {
	if q.err != nil {
		return 0, q.err
	}

	// Parse all fields. We expect a single table (level != 0)
	tbls, err := q.columnSelector.parseField(q.typ, 1, "")
	if err != nil {
		return 0, database.DatabaseError{
			Typ:      database.UnexpectedError,
			Err:      fmt.Errorf("failed to parse fields of struct %q: %w", q.typ, err),
			Response: errors.InternalError(),
		}
	}
	if len(tbls) < 1 {
		return 0, database.DatabaseError{
			Typ:      database.UnexpectedError,
			Err:      errors.New("no table received form parsing struct"),
			Response: errors.InternalError(),
		}
	}

	// Nothing to insert
	if q.insertVal.Len() == 0 {
		logger.Debug("Got no data to insert for table %q", getTableIdentifier(&tbls[0]))
		return 0, nil
	}

	// Insert 1 to 1 references first
	if q.columnSelector.ForeignKeyReference {
		if err := insert1To1Reference(tbls, q.operator, q.insertVal); err != nil {
			return 0, err
		}
	}

	// Build insert header
	insert := "INSERT INTO " + getTableIdentifier(&tbls[0]) + " (\n\t"
	ii := 0
	for _, c := range tbls[0].columns {
		// Skip non-insertable values. If we have only one column, it's the users fault!
		if c.PointedKeyReference != "" {
			continue
		}

		if ii != 0 {
			insert += ", "
		}
		insert += getColumnIdentifier(&tbls[0], &c)
		ii++
	}
	insert += "\n) VALUES"

	// Build insert data
	placeholders := make([]any, 0)
	primaryKeyPresent := false
	for rowI := range q.insertVal.Len() {
		ii = 0
		for colI, col := range tbls[0].columns {
			// Insert syntax preperatiosn
			if colI == 0 {
				if rowI != 0 {
					insert += ","
				}
				insert += "\n\t("
			}

			// If we have only one column, it's the users fault!
			if col.PointedKeyReference != "" {
				continue
			} else if col.IsPrimaryKey {
				primaryKeyPresent = true
			}
			if ii != 0 {
				insert += ", "
			}

			// Value to insert
			valueRef := q.insertVal.Index(rowI).Field(col.position)
			value := valueRef.Interface()
			// Mariadb will automatically assing an auto_increment for zero values
			if col.IsPrimaryKey && valueRef.IsZero() {
				value = nil
			}

			// If we have a 1:1 reference (which we don't create), the value to insert
			// has to be extracted from the referenced struct
			if col.isForeignKeyReference {
				position, err := getForeignKeyReferencePos(col, valueRef.Type())
				if err != nil {
					return 0, err
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
				insert += fmt.Sprintf("DEFAULT(%s)", getColumnIdentifier(&tbls[0], &col))
			} else {
				insert += "?"
				placeholders = append(placeholders, value)
			}

			ii++
		}
		insert += ")"
	}

	// Execute the insert (only if we have data to insert)
	insId := int64(0)
	if len(placeholders) > 0 && (len(placeholders) > q.insertVal.Len() || !primaryKeyPresent || !q.columnSelector.includePrimaryKeys) {
		res, err := q.operator.dbUtils.DB().Exec(insert, placeholders...)
		if err != nil {
			logger.Debug("Statement for failed insert:\n%s", insert)
			return 0, database.DatabaseError{
				Typ:      database.UnexpectedError,
				Err:      fmt.Errorf("failed to insert value: %w", err),
				Response: errors.InternalError(),
			}
		}

		// We will get the ID of the first inserted row (for MariaDB auto increment).
		// The ID of all other rows will ALWAYS be incremented by one (InnoDB)
		insId, _ = res.LastInsertId()
	}

	return insId, q.insertNTo1References(tbls, insId)
}

func (q *Insert) insertNTo1References(tbls []table, insId int64) database.Error {
	for _, col := range tbls[0].columns {
		if col.PointedKeyReference == "" || col.foreignKeyTable == nil || !q.columnSelector.PointedKeyReference {
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

		// Extract the field name to which the foreign key points to
		fieldName := coll.ForeignKeyReference
		lastPoint := strings.LastIndex(fieldName, ".")
		fieldName = structt.GetFieldName(fieldName[lastPoint+1:])

		// Build array with data we need to insert
		insArray := reflect.MakeSlice(reflect.SliceOf(col.foreignKeyTable.typ), 0, 0)

		// If we insert multiple rows at once with AUTO_INCREMENT, we also have multiple
		// primary keys. For MariaDB, they are incremented ALWAYS by once
		insIdCols := insId
		for rowI := range q.insertVal.Len() {
			field := q.insertVal.Index(rowI).Field(col.position)

			// Get identifier of the row we need to set for the foreign key.
			// This HAS TO BE the primary key of the table → use auto_increment
			// if last inserted ID it not zero.
			// But skip zero values. They wasn't counted up by MariaDb
			identifier := q.insertVal.Index(rowI).FieldByName(fieldName)
			if q.insertVal.Index(rowI).FieldByName(fieldName).IsZero() && insIdCols != 0 {
				identifier = reflect.ValueOf(int(insIdCols))
				insIdCols++
			}

			// Set this identifier for every element
			for i := range field.Len() {
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
			return database.DatabaseError{
				Typ:      database.UnexpectedError,
				Err:      fmt.Errorf("failed to insert pointed key reference %s: %w", qCopy.typ, errNew),
				Response: errors.InternalError(),
			}
		}
	}

	return nil
}

type oneToOneData struct {
	insertData   reflect.Value
	referencesId []reflect.Value
}

// insert1To1Reference inserts all 1:1 references if they do not
// exist yet
func insert1To1Reference(tbls []table, operator *Operator, data reflect.Value) database.Error {
	// Group insert values by full reference of destination table name
	groups := map[string]*oneToOneData{}

	for rowI := range data.Len() {
		for _, col := range tbls[0].columns {
			if !col.isForeignKeyReference {
				continue
			}

			refValue := data.Index(rowI).Field(col.position)
			group := getTableIdentifier(col.foreignKeyTable)

			// Only insert data where we don't have an ID but have data
			if refValue.Kind() == reflect.Pointer {
				if !refValue.IsValid() || refValue.IsZero() {
					continue
				}

				refValue = refValue.Elem()
			}

			position, err := getForeignKeyReferencePos(col, refValue.Type())
			if err != nil {
				return err
			}

			fieldVal := refValue.Field(position)
			if !fieldVal.IsZero() {
				continue
			}

			if _, exists := groups[group]; !exists {
				groups[group] = &oneToOneData{
					insertData: reflect.MakeSlice(reflect.SliceOf(refValue.Type()), 0, 0),
				}
			}

			// Set data
			groups[group].insertData = reflect.Append(groups[group].insertData, refValue)
			groups[group].referencesId = append(groups[group].referencesId, fieldVal)
		}
	}

	// Insert data per group
	for _, group := range groups {
		ins := operator.insertSlice(group.insertData)
		if id, err := ins.Run(); err != nil {
			return err
		} else {
			for i, ref := range group.referencesId {
				ref.Set(reflect.ValueOf(int(id) + i))
			}
		}
	}

	return nil
}

// getForeignKeyReferencePos returns the field position within the referenced struct
// for the foreign key constraint
func getForeignKeyReferencePos(col column, valueType reflect.Type) (position int, err database.Error) {
	// Get the referenced column position
	lastDot := strings.LastIndex(col.ForeignKeyReference, ".")
	referencedColumn := col.ForeignKeyReference[lastDot+1:]
	position = -1
	for _, cc := range col.foreignKeyTable.columns {
		if cc.Name == referencedColumn {
			position = cc.position
		}
	}

	// Referenced field not found
	if position == -1 {
		return 0, database.DatabaseError{
			Typ:      database.UnexpectedError,
			Err:      fmt.Errorf("didn't found referenced column %q in %s", referencedColumn, valueType),
			Response: errors.InternalError(),
		}
	}

	return position, nil
}
