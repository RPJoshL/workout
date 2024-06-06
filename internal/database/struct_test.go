package database

import (
	"fmt"
	"testing"

	tests "git.rpjosh.de/RPJosh/workout/internal/tests/extra"
	"github.com/google/go-cmp/cmp"
)

type TestQuery struct {
}

// TestQueryTypes tests allowed dst types that can be passed
// to the query function (single + multiple)
func TestQueryTypes(t *testing.T) {
	dbUtils := NewDatabaseUtils(tests.GetDb())

	// Allowed value
	var testQuery TestQuery
	if query := dbUtils.Struct.Query(&testQuery); query.err != nil {
		t.Errorf("Value should be allowed for Query(): %s", query.err)
	}
	// Empty pointer
	var testQueryNil *TestQuery
	if query := dbUtils.Struct.Query(testQueryNil); query.err == nil {
		t.Errorf("Empty pointer should be disallowed for Query()")
	}
	// Random type
	var testQueryString string
	if query := dbUtils.Struct.Query(&testQueryString); query.err == nil {
		t.Errorf("Data type should be disallowed for Query()")
	}

	// Array
	var testQueryArray []TestQuery
	if query := dbUtils.Struct.QuerySlice(&testQueryArray); query.err != nil {
		t.Errorf("Value should be allowed for QuerySlice(): %s", query.err)
	}
	// Array pointer
	var testQueryArrayPointer []*TestQuery
	if query := dbUtils.Struct.QuerySlice(&testQueryArrayPointer); query.err == nil {
		t.Errorf("Slice with pointer types should be disallowed for QuerySlice()")
	}
	// StringArray
	var testQueryArrayString []string
	if query := dbUtils.Struct.QuerySlice(&testQueryArrayString); query.err == nil {
		t.Errorf("Array of data type should be disallowed for QuerySlice()")
	}

}

type TestParseSimple struct {
	ID   int    `dbColumn:"Column:id,PrimaryKey"`
	Col1 string `dbColumn:"Column:col_1"`

	ExcludedField string

	DbMetadata_ any `dbMetadata:"Table:DDL_FIXED_TABLE_NAME_SIMPLE"`
}

func TestQuerySimple(t *testing.T) {
	dbUtils := NewDatabaseUtils(tests.GetDb())
	tblName, err := CreateTableWithName(dbUtils.Db, `id INT PRIMARY KEY, col_1 VARCHAR(100)`, "DDL_FIXED_TABLE_NAME_SIMPLE")
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	defer DropTable(dbUtils.Db, tblName)

	// Expected table
	expected := &TestParseSimple{
		ID:            200,
		Col1:          "Wario fights Joshi",
		ExcludedField: "Untouched value",
	}

	// Insert data
	if _, err := dbUtils.Db.Exec(
		fmt.Sprintf("INSERT INTO %s VALUES (?, ?)", tblName), expected.ID, expected.Col1,
	); err != nil {
		t.Errorf("Failed to insert data: %s", err)
	}

	// Get data
	got := &TestParseSimple{ExcludedField: "Untouched value"}
	if err := dbUtils.Struct.Query(got).Run(); err != nil {
		t.Errorf("Failed to select values: %s", err)
	}

	// Compare structs
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Mismatch of Query() (-want +got):\n%s", diff)
	}

	// Insert a second row
	expected2 := &TestParseSimple{
		ID:   100,
		Col1: "But Mario is here",
	}
	if _, err := dbUtils.Db.Exec(
		fmt.Sprintf("INSERT INTO %s VALUES (?, ?)", tblName), expected2.ID, expected2.Col1,
	); err != nil {
		t.Errorf("Failed to insert data 2: %s", err)
	}

	// Get data to fail
	if err := dbUtils.Struct.Query(got).Run(); err == nil {
		t.Errorf("Expected error for two rows in Query()")
	}

	// Query multiple columns
	gotArray := []TestParseSimple{}
	if err := dbUtils.Struct.QuerySlice(&gotArray).Run(); err != nil {
		t.Errorf("Failed to select values: %s", err)
	}
	expected.ExcludedField = ""
	expectedArray := []TestParseSimple{
		*expected2, *expected,
	}

	// Compare structs
	if diff := cmp.Diff(expectedArray, gotArray); diff != "" {
		t.Errorf("Mismatch of QuerySlice() (-want +got):\n%s", diff)
	}
}

type TestQueryWhereS struct {
	ID   int    `dbColumn:"Column:id,PrimaryKey"`
	Col2 string `dbColumn:"Column:col_2"`

	DbMetadata_ any `dbMetadata:"Table:DDL_FIXED_TABLE_NAME_WHERE"`
}

func TestQueryWhere(t *testing.T) {
	dbUtils := NewDatabaseUtils(tests.GetDb())
	tblName, err := CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY, col_2 VARCHAR(100)`,
		"DDL_FIXED_TABLE_NAME_WHERE",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	defer DropTable(dbUtils.Db, tblName)

	// Insert test data
	if _, err := dbUtils.Db.Exec(
		fmt.Sprintf("INSERT INTO %s VALUES (?, ?)", tblName), 10, "10. Part",
	); err != nil {
		t.Errorf("Failed to insert data 2: %s", err)
	}
	if _, err := dbUtils.Db.Exec(
		fmt.Sprintf("INSERT INTO %s VALUES (?, ?)", tblName), 20, "20. Part",
	); err != nil {
		t.Errorf("Failed to insert data 2: %s", err)
	}

	// Test basic where
	got := &TestQueryWhereS{}
	q := dbUtils.Struct.Query(got)
	q.Where().Column(tblName+".id", "=", 10).Add()
	if err := q.Run(); err != nil {
		t.Errorf("Failed to select values: %s", err)
	}

	expected := &TestQueryWhereS{
		ID:   10,
		Col2: "10. Part",
	}

	// Compare structs
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Mismatch of simple where statement (-want +got):\n%s", diff)
	}

	// Test empty
	gotArray := []TestQueryWhereS{}
	q = dbUtils.Struct.QuerySlice(&gotArray)
	q.Where().Column("fieldName|"+tblName+".id", "=", 0).IfNotZero()
	if err := q.Run(); err != nil {
		t.Errorf("Failed to select values: %s", err)
	}

	expectedArray := []TestQueryWhereS{
		*expected,
		{
			ID:   20,
			Col2: "20. Part",
		},
	}

	// Compare structs
	if diff := cmp.Diff(expectedArray, gotArray); diff != "" {
		t.Errorf("Mismatch of simple where statement (-want +got):\n%s", diff)
	}
}

// TestQueryCount tests the counting of data
func TestQueryCount(t *testing.T) {
	dbUtils := NewDatabaseUtils(tests.GetDb())
	tblName, err := CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY, col_2 VARCHAR(100)`,
		"DDL_FIXED_TABLE_NAME_WHERE",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	defer DropTable(dbUtils.Db, tblName)

	// Insert test data
	if _, err := dbUtils.Db.Exec(
		fmt.Sprintf("INSERT INTO %s VALUES (?, ?)", tblName), 10, "10. Part",
	); err != nil {
		t.Errorf("Failed to insert data 2: %s", err)
	}
	if _, err := dbUtils.Db.Exec(
		fmt.Sprintf("INSERT INTO %s VALUES (?, ?)", tblName), 20, "20. Part",
	); err != nil {
		t.Errorf("Failed to insert data 2: %s", err)
	}

	got := &TestQueryWhereS{}

	// Test basic count
	count, err := dbUtils.Struct.Query(got).Count()
	if err != nil {
		t.Errorf("Failed to count rows: %s", err)
	}
	if count != 2 {
		t.Errorf("Missmatch of row count. Expected 2. Got %d", count)
	}

	// Test with where
	q := dbUtils.Struct.Query(got)
	q.Where().Column(tblName+".id", "=", 10).Add()
	count, err = q.Count()
	if err != nil {
		t.Errorf("Failed to count rows: %s", err)
	}
	if count != 1 {
		t.Errorf("Missmatch of row count. Expected 1. Got %d", count)
	}
}

type TestParseReference struct {
	ID     int                         `dbColumn:"Column:id,PrimaryKey"`
	ColRef *TestParseReferenceIncluded `dbColumn:"Column:col_ref,ForeignKey:DDL_FIXED_TABLE_NAME_REFERENCE_INCLUDED.id"`

	DbMetadata_ any `dbMetadata:"Table:DDL_FIXED_TABLE_NAME_REFERENCE"`
}
type TestParseReferenceIncluded struct {
	ID   int    `dbColumn:"Column:id,PrimaryKey"`
	Col2 string `dbColumn:"Column:col_2"`

	DbMetadata_ any `dbMetadata:"Table:DDL_FIXED_TABLE_NAME_REFERENCE_INCLUDED"`
}

// Tests a 1:1 relationship
func TestQueryReference(t *testing.T) {
	dbUtils := NewDatabaseUtils(tests.GetDb())
	tblNameIncluded, err := CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY, col_2 VARCHAR(100)`,
		"DDL_FIXED_TABLE_NAME_REFERENCE_INCLUDED",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	defer DropTable(dbUtils.Db, tblNameIncluded)

	tblName, err := CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY, col_ref INT, 
		 CONSTRAINT FK_DDL_TEST_QUERY_REF FOREIGN KEY (col_ref) REFERENCES DDL_FIXED_TABLE_NAME_REFERENCE_INCLUDED(id)`,
		"DDL_FIXED_TABLE_NAME_REFERENCE",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	defer DropTable(dbUtils.Db, tblName)

	exptected := TestParseReference{
		ID: 100,
		ColRef: &TestParseReferenceIncluded{
			ID:   500,
			Col2: "My value",
		},
	}

	// Insert data
	if _, err := dbUtils.Db.Exec(
		fmt.Sprintf("INSERT INTO %s VALUES (?, ?)", tblNameIncluded), exptected.ColRef.ID, exptected.ColRef.Col2,
	); err != nil {
		t.Errorf("Failed to insert data 2: %s", err)
	}
	if _, err := dbUtils.Db.Exec(
		fmt.Sprintf("INSERT INTO %s VALUES (?, ?)", tblName), exptected.ID, exptected.ColRef.ID,
	); err != nil {
		t.Errorf("Failed to insert data 2: %s", err)
	}

	// Resolve reference
	got := &TestParseReference{}
	if err := dbUtils.Struct.Query(got).Run(); err != nil {
		t.Errorf("Failed to select values: %s", err)
	}

	// Compare structs
	if diff := cmp.Diff(&exptected, got); diff != "" {
		t.Errorf("Mismatch of Query with reference (-want +got):\n%s", diff)
	}

	// Without resolving reference we should only receive the ID
	got = &TestParseReference{}
	if err := dbUtils.Struct.Query(got).Selector(ColumnSelector{ForeignKeyReference: false}).Run(); err != nil {
		t.Errorf("Failed to select values: %s", err)
	}

	// Compare structs
	if diff := cmp.Diff(&TestParseReferenceIncluded{ID: exptected.ColRef.ID}, got.ColRef); diff != "" {
		t.Errorf("Mismatch of Query without reference (-want +got):\n%s", diff)
	}
}

type TestParseReferenceOneToN struct {
	Id     int    `dbColumn:"Column:id,PrimaryKey,AutoIncrement"`
	ColRef int    `dbColumn:"Column:col_ref,ForeignKey:DDL_FIXED_TABLE_NAME_REFERENCE_INCLUDED.id"`
	Val    string `dbColumn:"Column:val"`

	DbMetadata_ any `dbMetadata:"Table:DDL_FIXED_TABLE_NAME_REFERENCE"`
}

type TestParseReferenceIncludedOneToN struct {
	Id   int    `dbColumn:"Column:id,PrimaryKey,AutoIncrement"`
	Col2 string `dbColumn:"Column:col_2"`

	Included []TestParseReferenceOneToN `dbColumn:"PointedForeignKey:DDL_FIXED_TABLE_NAME_REFERENCE.col_ref"`

	DbMetadata_ any `dbMetadata:"Table:DDL_FIXED_TABLE_NAME_REFERENCE_INCLUDED"`
}

// Tests a n:1 relationship
func TestQueryOneToN(t *testing.T) {
	dbUtils := NewDatabaseUtils(tests.GetDb())
	tblNameIncluded, err := CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY, col_2 VARCHAR(100)`,
		"DDL_FIXED_TABLE_NAME_REFERENCE_INCLUDED",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	defer DropTable(dbUtils.Db, tblNameIncluded)

	tblName, err := CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY, col_ref INT, val VARCHAR(100),
		 CONSTRAINT FK_DDL_TEST_QUERY_REF FOREIGN KEY (col_ref) REFERENCES DDL_FIXED_TABLE_NAME_REFERENCE_INCLUDED(id)`,
		"DDL_FIXED_TABLE_NAME_REFERENCE",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	defer DropTable(dbUtils.Db, tblName)

	// Expect Data
	exptected := &TestParseReferenceIncludedOneToN{
		Id:   100,
		Col2: "Servus",
		Included: []TestParseReferenceOneToN{
			{
				Id:     1,
				ColRef: 100,
				Val:    "It's me, Mario",
			},
			{
				Id:     2,
				ColRef: 100,
				Val:    "And im Peach",
			},
		},
	}

	// Insert data
	if err := Insert1ToNReference(dbUtils, tblName, tblNameIncluded, *exptected); err != nil {
		t.Errorf(err.Error())
	}

	// Without resolving pointed reference we should not get any array element
	got := &TestParseReferenceIncludedOneToN{}
	if err := dbUtils.Struct.Query(got).Run(); err != nil {
		t.Errorf("Failed to select values: %s", err)
	}

	exptected2 := &TestParseReferenceIncludedOneToN{
		Id:   exptected.Id,
		Col2: exptected.Col2,
	}

	// Compare structs
	if diff := cmp.Diff(exptected2, got); diff != "" {
		t.Errorf("Mismatch of Query without pointed reference (-want +got):\n%s", diff)
	}

	// With resolving pointed reference we should not get any array element
	got = &TestParseReferenceIncludedOneToN{}
	if err := dbUtils.Struct.Query(got).Selector(ColumnSelector{PointedKeyReference: true}).Run(); err != nil {
		t.Errorf("Failed to select values: %s", err)
	}

	// Compare structs
	if diff := cmp.Diff(exptected, got); diff != "" {
		t.Errorf("Mismatch of Query with pointed reference (-want +got):\n%s", diff)
	}

	// Test array
	exptected3 := TestParseReferenceIncludedOneToN{
		Id:   101,
		Col2: "Moin",
		Included: []TestParseReferenceOneToN{
			{
				Id:     3,
				ColRef: 101,
				Val:    "Bowser",
			},
			{
				Id:     4,
				ColRef: 101,
				Val:    "Is here!",
			},
		},
	}
	// Insert data
	if err := Insert1ToNReference(dbUtils, tblName, tblNameIncluded, exptected3); err != nil {
		t.Errorf(err.Error())
	}

	gotArray := []TestParseReferenceIncludedOneToN{}
	expectedArray := []TestParseReferenceIncludedOneToN{*exptected, exptected3}
	if err := dbUtils.Struct.QuerySlice(&gotArray).Selector(ColumnSelector{PointedKeyReference: true}).Run(); err != nil {
		t.Errorf("Failed to select values: %s", err)
	} else {
		// Compare structs
		if diff := cmp.Diff(expectedArray, gotArray); diff != "" {
			t.Errorf("Mismatch of QuerySlice with pointed reference (-want +got):\n%s", diff)
		}
	}

	gotArray = []TestParseReferenceIncludedOneToN{}
	if err := dbUtils.Struct.QuerySlice(&gotArray).Selector(ColumnSelector{PointedKeyReference: true, PointedKeyReferenceAsync: true}).Run(); err != nil {
		t.Errorf("Failed to select ASYNC values: %s", err)
	} else {
		// Compare structs
		if diff := cmp.Diff(expectedArray, gotArray); diff != "" {
			t.Errorf("Mismatch of QuerySlice with ASYNC pointed reference (-want +got):\n%s", diff)
		}
	}

}

func Insert1ToNReference(dbUtils *DatabaseUtils, tblName, tblNameIncluded string, exptected TestParseReferenceIncludedOneToN) error {
	if _, err := dbUtils.Db.Exec(
		fmt.Sprintf("INSERT INTO %s VALUES (?, ?)", tblNameIncluded), exptected.Id, exptected.Col2,
	); err != nil {
		return fmt.Errorf("Failed to insert data 2: %s", err)
	}

	for _, inc := range exptected.Included {
		if _, err := dbUtils.Db.Exec(
			fmt.Sprintf("INSERT INTO %s VALUES (?, ?, ?)", tblName), inc.Id, inc.ColRef, inc.Val,
		); err != nil {
			return fmt.Errorf("Failed to insert data 2: %s", err)
		}
	}

	return nil
}

type TestCustomColumnType struct {
	ID              int `dbColumn:"Column:id,PrimaryKey"`
	AdditionalField int

	DbMetadata_ any `dbMetadata:"Table:DDL_FIXED_TABLE_NAME_CUSTOM_CULUMN"`
}
type TestCustomColumnTypeEmbedded struct {
	TestCustomColumnType
	AdditionalFieldEmbedded int
}

// TestCustomColumn tests a query with an additional custom column
func TestCustomColumn(t *testing.T) {
	dbUtils := NewDatabaseUtils(tests.GetDb())
	tblNameIncluded, err := CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY`,
		"DDL_FIXED_TABLE_NAME_CUSTOM_CULUMN",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	defer DropTable(dbUtils.Db, tblNameIncluded)

	// Insert data
	for i := 0; i < 2; i++ {
		if _, err := dbUtils.Db.Exec(
			fmt.Sprintf("INSERT INTO %s VALUES (?)", tblNameIncluded), i+1,
		); err != nil {
			t.Errorf("Failed to insert data: %s", err)
		}
	}

	expected := []TestCustomColumnType{
		{
			ID:              1,
			AdditionalField: 2,
		},
		{
			ID:              2,
			AdditionalField: 4,
		},
	}
	got := []TestCustomColumnType{}

	// Select with custom value
	sel := dbUtils.Struct.QuerySlice(&got).CustomColumn("", "AdditionalField", `
		id * 2
	`)
	if err := sel.Run(); err != nil {
		t.Errorf("Failed to select custom column: %s", err)
	}

	// Compare structs
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Mismatch of result (-want +got):\n%s", diff)
	}
}

// TestCustomColumnEmbedded tests a query with an additional custom column.
// Because the main use case of a custom column would be implemented in this
// case (the ddl generated struct mustn't be modified), we should test this
// also. The internal handling slightly different with embedded structs
func TestCustomColumnEmbedded(t *testing.T) {
	dbUtils := NewDatabaseUtils(tests.GetDb())
	tblNameIncluded, err := CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY`,
		"DDL_FIXED_TABLE_NAME_CUSTOM_CULUMN",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	defer DropTable(dbUtils.Db, tblNameIncluded)

	// Insert data
	for i := 0; i < 2; i++ {
		if _, err := dbUtils.Db.Exec(
			fmt.Sprintf("INSERT INTO %s VALUES (?)", tblNameIncluded), i+1,
		); err != nil {
			t.Errorf("Failed to insert data: %s", err)
		}
	}

	expected := []TestCustomColumnTypeEmbedded{
		{
			TestCustomColumnType: TestCustomColumnType{
				ID: 1,
			},
			AdditionalFieldEmbedded: 2,
		},
		{
			TestCustomColumnType: TestCustomColumnType{
				ID: 2,
			},
			AdditionalFieldEmbedded: 4,
		},
	}
	got := []TestCustomColumnTypeEmbedded{}

	// Select with custom value
	sel := dbUtils.Struct.QuerySlice(&got).CustomColumn("", "AdditionalFieldEmbedded", `
		id * 2
	`)
	if err := sel.Run(); err != nil {
		t.Errorf("Failed to select custom column: %s", err)
	}

	// Compare structs
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Mismatch of result (-want +got):\n%s", diff)
	}
}

// TestQueryIn tests the building of an IN statement
func TestQueryIn(t *testing.T) {
	dbUtils := NewDatabaseUtils(tests.GetDb())
	tblNameIncluded, err := CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY`,
		"DDL_FIXED_TABLE_NAME_CUSTOM_CULUMN",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	defer DropTable(dbUtils.Db, tblNameIncluded)

	// Insert data
	for i := 0; i < 10; i++ {
		if _, err := dbUtils.Db.Exec(
			fmt.Sprintf("INSERT INTO %s VALUES (?)", tblNameIncluded), i+1,
		); err != nil {
			t.Errorf("Failed to insert data: %s", err)
		}
	}

	expected := []TestCustomColumnType{
		{ID: 2},
		{ID: 5},
	}
	got := []TestCustomColumnType{}

	// Select with custom value
	sel := dbUtils.Struct.QuerySlice(&got).Where().Column("ID", "IN", []int{2, 5}).Add()
	if err := sel.Run(); err != nil {
		t.Errorf("Failed to use IN statement: %s", err)
	}

	// Compare structs
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Mismatch of result (-want +got):\n%s", diff)
	}

}

func TestInsert(t *testing.T) {
	dbUtils := NewDatabaseUtils(tests.GetDb())

	tblName, err := CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY AUTO_INCREMENT, col_2 VARCHAR(100)`,
		"DDL_FIXED_TABLE_NAME_REFERENCE_INCLUDED",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	defer DropTable(dbUtils.Db, tblName)

	expected := &TestParseReferenceIncludedOneToN{
		Id:   1,
		Col2: "MyValue",
	}

	// Insert a single value
	if id, err := dbUtils.Struct.Insert(expected).Run(); err != nil {
		t.Errorf("Failed to insert a single value: %s", err)
	} else {
		// Mariadb begins with 1 for auto_increment
		if id != 1 {
			t.Errorf("Got incorret ID for auto_increment: %d. Expected 1", id)
		}
	}

	// Validate with select statement
	got := &TestParseReferenceIncludedOneToN{}
	if err := dbUtils.Struct.Query(got).Run(); err != nil {
		t.Errorf("Failed to select a single value: %s", err)
	}

	// Compare
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Mismatch of single insert (-want +got):\n%s", diff)
	}

	// Insert multiple values
	expectedArray := &[]TestParseReferenceIncludedOneToN{
		{
			Id:   2,
			Col2: "200 is the looser!",
		},
		{
			Id:   3,
			Col2: "Four wins",
		},
	}

	// Insert multiple value
	if _, err := dbUtils.Struct.InsertSlice(expectedArray).Run(); err != nil {
		t.Errorf("Failed to insert multiple values: %s", err)
	}

	// Validate with select statement
	gotArray := &[]TestParseReferenceIncludedOneToN{}
	if err := dbUtils.Struct.QuerySlice(gotArray).Where().Column("id", "<>", expected.Id).Add().Run(); err != nil {
		t.Errorf("Failed to select multiple values: %s", err)
	}

	// Compare
	if diff := cmp.Diff(expectedArray, gotArray); diff != "" {
		t.Errorf("Mismatch of multiple insert (-want +got):\n%s", diff)
	}
}

// Tests an n:1 relationship
func TestInsertOneToN(t *testing.T) {
	testInsertOneToN(t, true)
}

func testInsertOneToN(t *testing.T, dropTable bool) (tableName []string) {
	dbUtils := NewDatabaseUtils(tests.GetDb())
	tblNameIncluded, err := CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY AUTO_INCREMENT, col_2 VARCHAR(100)`,
		"DDL_FIXED_TABLE_NAME_REFERENCE_INCLUDED",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	if dropTable {
		defer DropTable(dbUtils.Db, tblNameIncluded)
	}

	tblName, err := CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY AUTO_INCREMENT, col_ref INT, val VARCHAR(100),
		 CONSTRAINT FK_DDL_TEST_QUERY_REF FOREIGN KEY (col_ref) REFERENCES DDL_FIXED_TABLE_NAME_REFERENCE_INCLUDED(id)`,
		"DDL_FIXED_TABLE_NAME_REFERENCE",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	if dropTable {
		defer DropTable(dbUtils.Db, tblName)
	}

	// Data to insert
	insert := &[]TestParseReferenceIncludedOneToN{
		{
			Col2: "Servus",
			Included: []TestParseReferenceOneToN{
				{
					Val: "Hello",
				},
				{
					Val: "It's over",
				},
			},
		},
		{
			Col2: "Moin",
			Included: []TestParseReferenceOneToN{
				{
					Val: "For you!",
				},
				{
					Val: "And for me",
				},
			},
		},
	}

	// Insert multiple value
	if _, err := dbUtils.Struct.InsertSlice(insert).Selector(ColumnSelector{PointedKeyReference: true}).Run(); err != nil {
		t.Errorf("Failed to insert multiple values: %s", err)
	}

	// Validate with select statement
	gotArray := &[]TestParseReferenceIncludedOneToN{}
	if err := dbUtils.Struct.QuerySlice(gotArray).Selector(ColumnSelector{PointedKeyReference: true}).Run(); err != nil {
		t.Errorf("Failed to select multiple values: %s", err)
	}

	expected := &[]TestParseReferenceIncludedOneToN{
		{
			Id:   1,
			Col2: "Servus",
			Included: []TestParseReferenceOneToN{
				{
					Id:     1,
					Val:    "Hello",
					ColRef: 1,
				},
				{
					Id:     2,
					Val:    "It's over",
					ColRef: 1,
				},
			},
		},
		{
			Id:   2,
			Col2: "Moin",
			Included: []TestParseReferenceOneToN{
				{
					Id:     3,
					Val:    "For you!",
					ColRef: 2,
				},
				{
					Id:     4,
					Val:    "And for me",
					ColRef: 2,
				},
			},
		},
	}

	// Compare
	if diff := cmp.Diff(expected, gotArray); diff != "" {
		t.Errorf("Mismatch of multiple insert with reference (-want +got):\n%s", diff)
	}

	return []string{tblName, tblNameIncluded}
}

func TestUpdate(t *testing.T) {
	dbUtils := NewDatabaseUtils(tests.GetDb())

	tblName, err := CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY AUTO_INCREMENT, col_2 VARCHAR(100)`,
		"DDL_FIXED_TABLE_NAME_REFERENCE_INCLUDED",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	defer DropTable(dbUtils.Db, tblName)

	expected := &TestParseReferenceIncludedOneToN{
		Id:   1,
		Col2: "MyValue",
	}

	// Insert a single value
	if _, err := dbUtils.Struct.Insert(expected).Run(); err != nil {
		t.Fatalf("Failed to insert a single value: %s", err)
	}

	// Update the value
	expected.Col2 = "Updated value!"
	if err := dbUtils.Struct.Update(expected).Run(); err != nil {
		t.Errorf("Failed to update value: %s", err)
	}

	got := &TestParseReferenceIncludedOneToN{}
	if err := dbUtils.Struct.Query(got).Run(); err != nil {
		t.Errorf("Failed to select value: %s", err)
	}

	// Compare
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Mismatch of single update (-want +got):\n%s", diff)
	}

	// Update multiple values
	expectedArray := &[]TestParseReferenceIncludedOneToN{
		{
			Id:   2,
			Col2: "200 is the looser!",
		},
		{
			Id:   3,
			Col2: "Four wins",
		},
	}

	// Insert multiple value
	if _, err := dbUtils.Struct.InsertSlice(expectedArray).Run(); err != nil {
		t.Fatalf("Failed to insert multiple values: %s", err)
	}

	// Update thema again
	(*expectedArray)[0].Col2 = "And i'm the winner"
	(*expectedArray)[1].Col2 = "And you not"
	if err := dbUtils.Struct.UpdateSlice(expectedArray).Run(); err != nil {
		t.Errorf("Failed to update multiple values: %s", err)
	}

	// Validate with select statement
	gotArray := &[]TestParseReferenceIncludedOneToN{}
	if err := dbUtils.Struct.QuerySlice(gotArray).Where().Column("id", "<>", expected.Id).Add().Run(); err != nil {
		t.Errorf("Failed to select multiple values: %s", err)
	}

	// Compare
	if diff := cmp.Diff(expectedArray, gotArray); diff != "" {
		t.Errorf("Mismatch of multiple insert (-want +got):\n%s", diff)
	}
}

// Tests an n:1 relationship
func TestUpdateOneToN(t *testing.T) {
	dbUtils := NewDatabaseUtils(tests.GetDb())

	// Create basic data
	tblNames := testInsertOneToN(t, false)
	defer func() {
		for _, t := range tblNames {
			DropTable(dbUtils.Db, t)
		}
	}()

	// Basic data that were inserted in called test
	newData := []TestParseReferenceIncludedOneToN{
		{
			Id:   1,
			Col2: "Servus",
			Included: []TestParseReferenceOneToN{
				{
					Id:     1,
					Val:    "Hello",
					ColRef: 1,
				},
				{
					Id:     2,
					Val:    "It's over",
					ColRef: 1,
				},
			},
		},
		{
			Id:   2,
			Col2: "Moin",
			Included: []TestParseReferenceOneToN{
				{
					Id:     3,
					Val:    "For you!",
					ColRef: 2,
				},
				{
					Id:     4,
					Val:    "And for me",
					ColRef: 2,
				},
			},
		},
	}

	// Modify test
	newData[0].Included = []TestParseReferenceOneToN{}
	newData[1].Included = []TestParseReferenceOneToN{
		{
			Id:     5,
			Val:    "It's over",
			ColRef: 2,
		},
		{
			Id:     6,
			Val:    "And for me",
			ColRef: 2,
		},
	}

	// Update
	gotArray := &[]TestParseReferenceIncludedOneToN{}
	if err := dbUtils.Struct.UpdateSlice(&newData).Selector(ColumnSelector{PointedKeyReference: true}).Run(); err != nil {
		t.Errorf("Failed to update 1:n relationship: %s", err)
	}
	if err := dbUtils.Struct.QuerySlice(gotArray).Selector(ColumnSelector{PointedKeyReference: true}).Run(); err != nil {
		t.Errorf("Failed to select data for test: %s", err)
	}

	// Compare
	if diff := cmp.Diff(&newData, gotArray); diff != "" {
		t.Errorf("Mismatch of multiple insert with reference (-want +got):\n%s", diff)
	}
}
