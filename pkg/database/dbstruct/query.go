package dbstruct

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"git.rpjosh.de/RPJosh/go-ddl-parser"
	"git.rpjosh.de/RPJosh/go-ddl-parser/structt"
	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/pkg/database"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"git.rpjosh.de/RPJosh/workout/pkg/utils"
)

type Query struct {

	// Internal reference to an operator
	operator *Operator

	// The destination type is an array instead
	// of a single struct
	isArray bool

	// Destination to write the result to:
	//  - Single value: pointer to a struct (*struct)
	//  - Multiple values: array of structs ([]struct)
	dst reflect.Value

	// Type of the struct to write the data to (single + multiple)
	typ reflect.Type

	// The last occurred error will be stored in this fild and is
	// only returned in "Run()".
	// If any error occurred, the (internal) processing is stopped
	err database.Error

	// Fields to exclude to query
	columnSelector ColumnSelector

	// Whether the selector was set
	customSelector bool

	// Custom columns to add to the query.
	// The name of a column is used as an index.
	// The value is the matching select value
	customColumns map[string]map[string]string

	// Custom where statement
	whereStatement string
	// Placeholders for the where statement
	wherePlaceholder []any

	// Order by statements set by user indexed by the table name in UPPER_CASE.
	// An empty string is specified for the root
	orderBy map[string]string

	// Custom order by statement
	customOrderBy             string
	customOrderByPlaceholders []any

	// Internal number of rows in a count query
	count int

	// Name of the sub table queried for n:1 relationships
	subTable string

	// Custom JOIN statement to append
	customJoin            string
	customJoinPlaceholder []any

	// Is used to set a nil value for 1:1 references when the foreign
	// key is null
	nilMapper []nilMapper
}

// nilMapper contains information to set a nil value to a mapped column
// by scanf
type nilMapper struct {
	// Reflection reference to the struct field for setting the nil value
	val reflect.Value

	// Whether the value is null
	isNull int
}

// Query analyzes the provided struct and selects all fields tagged
// with "dbColumn" from one or multiple tables and writes the result
// into *dst.
// If **dst is a nil pointer, the underlaying value to which the pointer points
// to will be initialized
//
// Exactly a single row is expected to be returned from the database.
func (o *Operator) Query(dst any) *Query {
	rtc := &Query{
		operator:      o,
		customColumns: map[string]map[string]string{},
		orderBy:       map[string]string{},
	}

	// Validate given type
	dstVal, err := isPointer(dst, reflect.Struct, true)
	if err != nil {
		rtc.err = database.DatabaseError{
			Typ:      database.UnexpectedError,
			Err:      fmt.Errorf("invalid type for dst given: %w", err),
			Response: errors.InternalError(),
		}
	}
	rtc.dst = dstVal
	rtc.typ = dstVal.Type().Elem()

	return rtc
}

// QuerySlice is like [Query] but writes multiple rows from the database
// to an array (*[]dst).
func (o *Operator) QuerySlice(dst any) *Query {
	rtc := &Query{
		isArray:       true,
		operator:      o,
		customColumns: map[string]map[string]string{},
		orderBy:       map[string]string{},
	}

	// Make sure that we got a slice
	dstType := reflect.TypeOf(dst)
	if dstType.Kind() != reflect.Pointer || dstType.Elem().Kind() != reflect.Slice || dstType.Elem().Elem().Kind() != reflect.Struct {
		rtc.err = database.DatabaseError{
			Typ:      database.UnexpectedError,
			Err:      errors.New("expected a pointer to a slice containing structs for dst"),
			Response: errors.InternalError(),
		}
		return rtc
	}
	rtc.dst = reflect.ValueOf(dst).Elem()
	rtc.typ = dstType.Elem().Elem()

	return rtc
}

// Selector adds a column selector for the query
func (q *Query) Selector(selector ColumnSelector) *Query {
	q.columnSelector = selector
	q.customSelector = true
	return q
}

// Custom adds a custom WHERE condition to the statement.
// You mustn't use this together with [Column()].
func (w *Where) Custom(statement string, values ...any) *Where {
	w.customWhere = strings.TrimSpace(statement)
	w.customValue = values

	return w
}

// IfNotZero adds this expression to the query if "value" and "operator" is not zero.
// Only generic go types like string or int are supported
func (w *Where) IfNotZero() *Query {
	if !w.isZero(w.val) && w.operator != "" {
		w.Add()
	}

	return w.query
}

// IfAllNotZero adds this expression if all provided values are
// NOT zero. Only generic go types are supported
func (w *Where) IfAllNotZero(vals ...any) *Query {
	for _, val := range vals {
		if w.isZero(val) {
			return w.query
		}
	}

	w.Add()
	return w.query
}

// isZero returns weather the provided value contains
// a default value for it's type
func (w *Where) isZero(val any) bool {
	refVal := reflect.ValueOf(val)

	// Default go type zero check
	if refVal.IsZero() {
		return true
	}

	// Array length > 0
	if (refVal.Kind() == reflect.Slice || refVal.Kind() == reflect.Array) && refVal.Len() == 0 {
		return true
	}

	return false
}

// OrderBy adds a custom order by statement to the select query.
// The name of a column and the sort order is expected ("COL", "ASC", "COL2", "DESC").
//
// The first value identifies the table name to add to the order by statement (for 1:n relationships).
// Provide an empty string for the root table
func (q *Query) OrderBy(tableName string, vals ...string) *Query {
	tableName = strings.ToUpper(tableName)
	orderBy := q.orderBy[tableName]

	for i := 0; i < len(vals); i++ {
		// Extract column name
		columnName := vals[i]
		if strings.Count(columnName, "|") == 1 {
			pipe := strings.Index(columnName, "|")
			columnName = columnName[pipe+1:]
		}

		if i == len(vals)-1 {
			logger.Warning("No sort direction provided for column %q", columnName)
			break
		}

		// Add column to statement
		if orderBy != "" {
			orderBy += ", "
		}
		orderBy += columnName + " " + vals[i+1]
		i++
	}

	q.orderBy[tableName] = orderBy
	return q
}

// CustomOrderBy adds an custom order by statement to the query.
// It is appended to the "default" statements defined by [OrderBy]
func (q *Query) CustomOrderBy(statement string, placeholders ...any) *Query {
	q.customOrderBy = statement
	q.customOrderByPlaceholders = placeholders

	return q
}

// CustomColumn adds a custom column to select from the table.
//
// The first value identifies the table name to add this custom select for.
// Provide an empty string for the root table
func (q *Query) CustomColumn(tableName, fieldName, sel string) *Query {
	tableName = strings.ToUpper(tableName)
	selVal := q.customColumns[tableName]
	if selVal == nil {
		selVal = map[string]string{}
	}
	selVal[fieldName] = sel

	q.customColumns[tableName] = selVal
	return q
}

// CustomJoin appends the provided JOIN statement after the
// automatically generated ones
func (q *Query) CustomJoin(join string, placeholders ...any) *Query {
	if q.customJoin != "" {
		q.customJoin += "\n" + join
		q.customJoinPlaceholder = append(q.customJoinPlaceholder, placeholders...)
	} else {
		q.customJoin = join
		q.customJoinPlaceholder = placeholders
	}

	return q
}

// GetWhereStatement returns the previously build SQL statement and
// placeholders of the WHERE condition
func (q *Query) GetWhereStatement() (statement string, placeholder []any) {
	return q.whereStatement, q.wherePlaceholder
}

// Count only counts the number of rows and doesn't
// write anything into [dst]
func (q *Query) Count() (int, database.Error) {
	err := q.run(true)
	return q.count, err
}

// Run executes the query and fetches the data from the database
// into *dst.
//
// Any error that occurred druing building the query or while fetching
// the data is returned here
func (q *Query) Run() database.Error {
	return q.run(false)
}

func (q *Query) run(onlyCount bool) database.Error {
	if q.err != nil {
		return q.err
	}

	// Apply default selector values
	if !q.customSelector {
		q.columnSelector.ForeignKeyReference = true
	}

	// Parse all struct fields
	tbls, err := q.columnSelector.parseField(q.typ, 0, "")
	if err != nil {
		return database.DatabaseError{
			Typ:      database.UnexpectedError,
			Err:      fmt.Errorf("failed to parse fields of struct %q: %w", q.typ, err),
			Response: errors.InternalError(),
		}
	}

	// Get select statement to execute (use a dummy struct to not initialize fields)
	dummy := reflect.New(q.typ)
	sel := "SELECT\n"
	var join, from, orderBy string
	for i, t := range tbls {
		// reflect.New creates a pointer
		writeTo := dummy.Elem()
		var rootWriteTo reflect.Value
		if t.fieldName != "" {
			writeTo = writeTo.FieldByName(t.fieldName)
			if writeTo.Kind() == reflect.Pointer {
				writeTo = writeTo.Elem()
			}

			rootWriteTo = dummy.Elem()
		}
		selAdd, joinAdd, _ := q.getColumns(t, writeTo, &rootWriteTo, &tbls, false)
		if q.err != nil {
			return q.err
		}

		// Initialize from statement
		if i == 0 {
			from = getTableIdentifier(&t)
		} else if from != getTableIdentifier(&t) {
			join = "JOIN " + getTableIdentifier(&t)
		}

		// Append select data
		sel += selAdd
		join += joinAdd

		// Order by PrimaryKeys
		for _, c := range t.columns {
			if c.IsPrimaryKey {
				if orderBy == "" {
					orderBy = getColumnIdentifier(&t, &c)
				} else {
					orderBy += ", " + getColumnIdentifier(&t, &c)
				}
			}
		}
	}

	// Add custom joins
	if q.customJoin != "" {
		join += "\n" + q.customJoin + "\n"
		q.wherePlaceholder = append(q.customJoinPlaceholder, q.wherePlaceholder...)
	}

	// Get custom order by statement
	custOrderBy := q.orderBy[q.subTable]
	if custOrderBy != "" {
		orderBy = "\nORDER BY " + custOrderBy
	} else if orderBy != "" && q.customOrderBy == "" {
		orderBy = "\nORDER BY " + orderBy
	}
	if q.customOrderBy != "" {
		if orderBy != "" {
			orderBy += ", "
		} else {
			orderBy = "\nORDER BY "
		}
		orderBy += q.customOrderBy
		q.wherePlaceholder = append(q.wherePlaceholder, q.customOrderByPlaceholders...)
	}

	if q.whereStatement != "" {
		q.whereStatement = "\n" + q.whereStatement
	}

	// Only count number of rows
	if onlyCount {
		sel = "SELECT COUNT(*) "
	} else {
		// Always select at least a single value.
		// We also need this to close any comma
		sel += "\t1"
	}

	// Execute the statement
	fullSelect := fmt.Sprintf("%s\nFROM %s\n%sWHERE 1=1%s%s", sel, from, join, q.whereStatement, orderBy)
	rows, err := q.operator.dbUtils.DB().Query(fullSelect, q.wherePlaceholder...)
	if err != nil {
		logger.Debug("Select for failed query:\n%s", fullSelect)
		return database.DatabaseError{
			Typ:      database.UnexpectedError,
			Err:      fmt.Errorf("failed to query value: %w", err),
			Response: errors.InternalError(),
		}
	}
	defer rows.Close()

	if err := q.extractQueryResult(onlyCount, rows, tbls, fullSelect); err != nil {
		return err
	}

	return q.queryNTo1References(tbls)
}

func (q *Query) extractQueryResult(onlyCount bool, rows *sql.Rows, tbls []table, fullSelect string) database.Error {
	// We always get a single result for count(*)
	if onlyCount {
		rows.Next()
		if err := rows.Scan(&q.count); err != nil {
			return database.DatabaseError{
				Typ:      database.UnexpectedError,
				Err:      fmt.Errorf("failed to scan row: %w", err),
				Response: errors.InternalError(),
			}
		}

		return nil
	}

	// Expect a single value
	var tmp any
	if !q.isArray && rows.Next() {
		var columns []any
		for _, t := range tbls {
			// Get value to write to
			writeTo := q.dst.Elem()
			var rootWriteTo reflect.Value
			if t.fieldName != "" {
				writeTo = writeTo.FieldByName(t.fieldName)
				if writeTo.Kind() == reflect.Pointer {
					writeTo = writeTo.Elem()
				}

				rootWriteTo = q.dst.Elem()
			}

			_, _, mappedAdd := q.getColumns(t, writeTo, &rootWriteTo, &tbls, false)
			columns = append(columns, mappedAdd...)
		}
		columns = append(columns, &tmp)

		// Scan the data into the struct
		if err := rows.Scan(columns...); err != nil {
			logger.Error("Query error for db: %s", err)
			return database.DatabaseError{
				Typ:      database.UnexpectedError,
				Err:      fmt.Errorf("failed to scan row: %w", err),
				Response: errors.InternalError(),
			}
		}
	} else if !q.isArray {
		return database.DatabaseError{
			Typ:      database.NoRows,
			Err:      errors.New("no data found in select"),
			Response: errors.NewError("No data found", 404),
		}
	}
	// Are there any remaining rows?
	if !q.isArray && rows.Next() {
		// Get the count of them for debug purporses
		counter := 2
		for rows.Next() {
			counter++
		}

		return database.DatabaseError{
			Typ:      database.TooManyRows,
			Err:      fmt.Errorf("found %d rows instead of a single one", counter),
			Response: errors.NewError("Too many data found", 409),
		}
	}

	// Multiple values
	i := 0
	for rows.Next() {
		i++
		dst := reflect.New(q.typ)
		var columns []any
		for _, t := range tbls {
			// Get value to write to
			writeTo := dst.Elem()
			var rootWriteTo reflect.Value
			if t.fieldName != "" {
				writeTo = writeTo.FieldByName(t.fieldName)
				if writeTo.Kind() == reflect.Pointer {
					writeTo = writeTo.Elem()
				}

				rootWriteTo = dst.Elem()
			}
			_, _, mappedAdd := q.getColumns(t, writeTo, &rootWriteTo, &tbls, false)
			columns = append(columns, mappedAdd...)
		}
		columns = append(columns, &tmp)

		// Scan elements
		if err := rows.Scan(columns...); err != nil {
			logger.Error("Query error for db: %s", err)
			return database.DatabaseError{
				Typ:      database.UnexpectedError,
				Err:      fmt.Errorf("failed to scan row: %w", err),
				Response: errors.InternalError(),
			}
		} else {
			q.dst.Set(reflect.Append(q.dst, dst.Elem()))
		}

		// Limit max number of rows
		if i > 100000 {
			logger.Warning("Received maximum result size of 100.000 rows. Aborting")
			logger.Debug("Select statement:\n%s", fullSelect)
			_ = rows.Close()
			break
		}
	}

	// Map nil values for 1:1 reference
	for _, n := range q.nilMapper {
		if n.isNull == 1 {
			n.val.Set(reflect.Zero(n.val.Type()))
		}
	}

	return nil
}

func (q *Query) queryNTo1References(tbls []table) database.Error {
	if q.columnSelector.PointedKeyReference && (!q.isArray || q.dst.Len() > 0) {
		// Find all fields with a pointed key reference
		for _, t := range tbls {
			var subError error
			if q.isArray {
				valArray := []reflect.Value{}
				useAllSelect := !q.columnSelector.PointedKeyReferenceAsync
				var wg sync.WaitGroup

				for i := range q.dst.Len() {
					var thisVal reflect.Value
					if t.fieldName != "" {
						// Query embedded
						thisVal = q.dst.Index(i).FieldByName(t.fieldName).Addr()
					} else {
						thisVal = q.dst.Index(i).Addr()
					}

					if useAllSelect {
						valArray = append(valArray, thisVal)
					} else {
						wg.Add(1)
						go func() {
							if err := q.findAllPointedReferences(t, []reflect.Value{thisVal}); err != nil {
								logger.Warning("Failed to select pointed key reference: %s", err)
							}
							wg.Done()
						}()
					}
				}
				wg.Wait()

				if useAllSelect {
					subError = q.findAllPointedReferences(t, valArray)
				}
			} else {
				valArray := []reflect.Value{}
				if t.fieldName != "" {
					// Query embedded
					valArray = append(valArray, q.dst.FieldByName(t.fieldName))
				} else {
					valArray = append(valArray, q.dst)
				}

				subError = q.findAllPointedReferences(t, valArray)
			}

			if subError != nil {
				return database.DatabaseError{
					Typ:      database.UnexpectedError,
					Err:      fmt.Errorf("failed to scan pointed key reference: %w", subError),
					Response: errors.InternalError(),
				}
			}
		}
	}

	return nil
}

// findAllPointedReferences queries struct fields with a tag "PointedForeignKey"
// to resolve n:1 relationships and queries the data from the db.
// Val is expected to be a *struct represented by table
func (q *Query) findAllPointedReferences(t table, values []reflect.Value) error {
	// Loop through all fields and find pointed key references
	for _, c := range t.columns {
		if c.PointedKeyReference == "" {
			// Nothing found
			continue
		}

		// Find field
		col := column{}
		for _, cc := range c.foreignKeyTable.columns {
			if getColumnIdentifier(c.foreignKeyTable, &cc) == c.PointedKeyReference {
				col = cc
				break
			}
		}
		if col.fieldName == "" || col.ForeignKeyReference == "" {
			// Suspress error if all fieds from the pointed table were ignored
			tableIdentifier1 := c.PointedKeyReference[0:strings.LastIndex(c.PointedKeyReference, ".")]
			tableIdentifier2 := tableIdentifier1
			if strings.Contains(tableIdentifier2, ".") {
				tableIdentifier2 = tableIdentifier2[strings.LastIndex(tableIdentifier2, ".")+1:]
			}
			for _, ex := range q.columnSelector.ExcludeColumns {
				if ex == "*|"+tableIdentifier1 || ex == "*|"+tableIdentifier2 {
					return nil
				}
			}

			return fmt.Errorf("no referenced field found for %q in %q", c.PointedKeyReference, c.foreignKeyTable.typ)
		}

		// Extract the field name to which the foreign key points to
		fieldName := col.ForeignKeyReference
		lastPoint := strings.LastIndex(fieldName, ".")
		fieldName = structt.GetFieldName(fieldName[lastPoint+1:])

		// Copy this query struct to apply customizations
		qCopy := *q
		qCopy.isArray = true
		qCopy.wherePlaceholder = []any{}
		qCopy.whereStatement = fmt.Sprintf("\tAND %s IN (", getColumnIdentifier(c.foreignKeyTable, &col))
		// Get the table name to fetch
		qCopy.subTable = strings.ToUpper(c.foreignKeyTable.Table)

		// Destination to write this value to
		for i, val := range values {
			field := val.Elem().Field(c.position)
			identValue := val.Elem().FieldByName(fieldName)

			// Initialize type
			if i == 0 {
				qCopy.typ = field.Type().Elem()
				// We use the first element as a buffer
				qCopy.dst = field.Addr().Elem()
			} else {
				qCopy.whereStatement += ", "
			}

			// Add to where statement
			qCopy.wherePlaceholder = append(qCopy.wherePlaceholder, identValue.Interface())
			qCopy.whereStatement += "?"
		}
		qCopy.whereStatement += ")"

		// Query it!
		if err := qCopy.Run(); err != nil {
			return err
		}

		// Map the array elements to the correct field again
		if len(values) != 1 {
			// Initialize holder values
			holders := make([]pointedReferenceHolder, len(values))
			for i, val := range values {
				field := val.Elem().Field(c.position)
				identValue := val.Elem().FieldByName(fieldName)

				holders[i] = pointedReferenceHolder{
					Field:      field,
					Identity:   identValue,
					FieldSlice: reflect.MakeSlice(field.Type(), 0, 0),
				}
			}

			// Loop through all values
			for i := range qCopy.dst.Len() {
				elValue := qCopy.dst.Index(i).FieldByName(col.fieldName)

				for o, val := range holders {
					if val.Identity.Equal(elValue) {
						holders[o].FieldSlice = reflect.Append(holders[o].FieldSlice, qCopy.dst.Index(i))
					}
				}
			}

			// Set fields
			for _, holder := range holders {
				holder.Field.Set(holder.FieldSlice)
			}
		}
	}

	return nil
}

// pointedReferenceHolder is a placeholder struct to store information
// for a specific 1:n field
type pointedReferenceHolder struct {
	// Field of the struct to set the array value to
	Field reflect.Value

	// Identity value
	Identity reflect.Value

	// Slice containing the temporary elements
	FieldSlice reflect.Value
}

// getColumns returns a list of columns (comma and \n separated)
// of the provided table.
//
// Any join statement that is needed to select
// these columns will be added.
//
// If "withDefaults" is provided, all null values are replaced with the datatypes
// default value (as in go)
//
// Because the mechanism of creating a select statement and mapping
// the column values to a struct is the same, this function initializes
// an array of pointers that are pointing to the matching
// fields of dst
func (q *Query) getColumns(tbl table, dst reflect.Value, root *reflect.Value, tbls *[]table, withDefaults bool) (sel, join string, mapped []any) {
	if q.err != nil {
		return
	}

	for _, c := range tbl.columns {
		if c.PointedKeyReference != "" {
			// We select this value later (array of []struct)
		} else if c.ForeignKeyReference != "" && c.isForeignKeyReference {
			// 1:1 relationship → we expect a pointer to a struct
			field := dst.Field(c.position)
			if field.Type().Kind() != reflect.Pointer || field.Type().Elem().Kind() != reflect.Struct {
				q.err = database.DatabaseError{
					Typ:      database.UnexpectedError,
					Err:      fmt.Errorf("expected a pointer to a struct for table type %s.%s", tbl.Table, c.Name),
					Response: errors.InternalError(),
				}
				return
			}

			// The pointer can be nil → initialize a new struct
			var structVal reflect.Value
			if field.IsNil() {
				structVal = reflect.New(field.Type().Elem())
				// Set field to pointer
				field.Set(structVal)
				structVal = structVal.Elem()
			} else {
				structVal = field.Elem()
			}

			if q.columnSelector.ForeignKeyReference {
				lastDot := strings.LastIndex(c.ForeignKeyReference, ".")
				referencedTable := c.ForeignKeyReference[0:lastDot]
				referencedColumn := c.ForeignKeyReference[lastDot+1:]

				// When two references to the same table exists, we cannot use the full qualified table name.
				// We use for that a random ID as a table alias
				alias := ""
				aliasId := ""
				aliasRef := c.ForeignKeyReference
				sameTypeCount := 0
				for _, tbl := range *tbls {
					sameTypeCount += cntSameForeignKeyReference(tbl, c.foreignKeyTable)
				}
				if sameTypeCount > 1 {
					// Include table name for a better debug experience
					refTable := referencedTable
					if strings.Contains(refTable, ".") {
						refTable = refTable[strings.LastIndex(refTable, ".")+1:]
						refTable = strings.ReplaceAll(refTable, ".", "-")
					}

					aliasId = "alias__" + refTable + "__" + utils.WithoutError(utils.GenerateRandomString(8))
					alias = " AS " + aliasId
					aliasRef = aliasId + "." + referencedColumn
				}

				sourceField := getColumnIdentifier(&tbl, &c)
				join += fmt.Sprintf("LEFT JOIN %s%s ON %s = %s\n", referencedTable, alias, aliasRef, sourceField)

				refTableName := referencedTable
				refTableSchema := ""
				lastDotTwo := strings.LastIndex(referencedTable, ".")
				if lastDotTwo != -1 {
					refTableSchema = referencedTable[0:lastDotTwo]
					refTableName = referencedTable[lastDotTwo+1:]
				}
				tbl := table{
					MetadataTag: structt.MetadataTag{
						Schema: refTableSchema,
						Table:  refTableName,
					},
					typ:     structVal.Type(),
					columns: c.foreignKeyTable.columns,
				}
				addSel, addJoin, addMapped := q.getColumns(tbl, structVal, nil, tbls, true)

				// Replace the table identifier with the generated alias name
				if aliasId != "" {
					addSel = strings.ReplaceAll(addSel, referencedTable, aliasId)
				}

				// Add a custom field to identify a nullish foreign key
				addSel += "\t" + sourceField + " IS NULL,\n"
				q.nilMapper = append(q.nilMapper, nilMapper{val: field})
				addMapped = append(addMapped, &q.nilMapper[len(q.nilMapper)-1].isNull)

				sel += addSel
				join += addJoin
				mapped = append(mapped, addMapped...)
			} else {
				// Include only the ID of the struct in select and reference
				// the struct ID
				sel += "\t" + getColumnIdentifier(&tbl, &c) + ",\n"

				// Find the position of the referenced column in the struct
				lastDot := strings.LastIndex(c.ForeignKeyReference, ".")
				referencedColumn := c.ForeignKeyReference[lastDot+1:]
				position := -1
				for _, cc := range c.foreignKeyTable.columns {
					if cc.Name == referencedColumn {
						position = cc.position
					}
				}

				// Referenced field not found
				if position == -1 {
					q.err = database.DatabaseError{
						Typ:      database.UnexpectedError,
						Err:      fmt.Errorf("didn't found referenced column %q in %s", referencedColumn, structVal.Type()),
						Response: errors.InternalError(),
					}
					return
				}

				// Get the pointer to the referenced column
				mapped = append(mapped, structVal.Field(position).Addr().Interface())
			}
		} else {
			// Simple select field
			sel += "\t" + withDefaultValue(&tbl, &c, dst.Field(c.position), withDefaults) + ",\n"
			mapped = append(mapped, dst.Field(c.position).Addr().Interface())
		}
	}

	// Add any custom columns
	if customMap, ok := q.customColumns[q.subTable]; ok {
		for fieldName, sellecto := range customMap {
			dstVal := dst

			_, found := dstVal.Type().FieldByName(fieldName)
			if !found {
				// If an embedded struct was used, we also try to fetch the value
				// of the embedded type
				err := database.DatabaseError{
					Typ:      database.UnexpectedError,
					Err:      fmt.Errorf("didn't found custom field %q in %q", fieldName, dst.Type().Name()),
					Response: errors.InternalError(),
				}

				if root != nil && root.IsValid() {
					if _, found := root.Type().FieldByName(fieldName); found {
						dstVal = *root
					} else {
						q.err = err
						return
					}
				} else {
					q.err = err
					return
				}
			}

			// Simple select field
			sel += "\t" + sellecto + ",\n"
			mapped = append(mapped, dstVal.FieldByName(fieldName).Addr().Interface())
		}
	}

	return
}

// cntSameForeignKeyReference counts the number of foreign key references to the provided
// table
func cntSameForeignKeyReference(table table, ref *table) (rtc int) {
	if ref == nil {
		return 0
	}

	for _, c := range table.columns {
		if c.foreignKeyTable != nil && c.foreignKeyTable.Schema == ref.Schema && c.foreignKeyTable.Table == ref.Table {
			rtc++
		}
	}

	return
}

// withDefaultValue will return a colum select value with a COALESC
// for the provided column identifier if "withDefault" is true.
//
// This is required for an 1:1 reference where the foreign key can be null.
// To not throw an exception while scanning, null values (from the db) are replaced
// by the go zero type
func withDefaultValue(tbl *table, col *column, field reflect.Value, withDefault bool) string {
	identifier := getColumnIdentifier(tbl, col)
	if !withDefault {
		return identifier
	}

	typ := field.Type()
	if typ.Kind() == reflect.Pointer {
		typ = field.Elem().Type()
	}

	defaultValue := ""
	switch typ.Kind() {
	case reflect.Int, reflect.Float64:
		defaultValue = "0"
	case reflect.String:
		defaultValue = "''"
	case reflect.Struct:
		switch typ {
		case reflect.TypeOf(time.Time{}):
			defaultValue = "'0000-00-00'"
		case reflect.TypeOf(ddl.Location{}):
			// DDL location supports nil values
			return identifier
		}
	}

	if defaultValue != "" {
		return fmt.Sprintf("COALESCE(%s, %s)", identifier, defaultValue)
	} else {
		logger.Trace("Could not determine data type for adding COALESC statement: %s", field.Type().PkgPath())
		return identifier
	}
}
