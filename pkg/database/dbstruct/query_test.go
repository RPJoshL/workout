package dbstruct

import (
	"fmt"
	"testing"

	tests "git.rpjosh.de/RPJosh/workout/internal/tests/extra"
	"git.rpjosh.de/RPJosh/workout/pkg/database"
	"github.com/google/go-cmp/cmp"
)

type TestQuery struct {
}

// TestQueryTypes tests allowed dst types that can be passed
// to the query function (single + multiple)
func TestQueryTypes(t *testing.T) {
	str := NewOperator(database.NewUtils(tests.GetDb()))

	// Allowed value
	var testQuery TestQuery
	if query := str.Query(&testQuery); query.err != nil {
		t.Errorf("Value should be allowed for Query(): %s", query.err)
	}
	// Empty pointer
	var testQueryNil *TestQuery
	if query := str.Query(testQueryNil); query.err == nil {
		t.Errorf("Empty pointer should be disallowed for Query()")
	}
	// Random type
	var testQueryString string
	if query := str.Query(&testQueryString); query.err == nil {
		t.Errorf("Data type should be disallowed for Query()")
	}

	// Array
	var testQueryArray []TestQuery
	if query := str.QuerySlice(&testQueryArray); query.err != nil {
		t.Errorf("Value should be allowed for QuerySlice(): %s", query.err)
	}
	// Array pointer
	var testQueryArrayPointer []*TestQuery
	if query := str.QuerySlice(&testQueryArrayPointer); query.err == nil {
		t.Errorf("Slice with pointer types should be disallowed for QuerySlice()")
	}
	// StringArray
	var testQueryArrayString []string
	if query := str.QuerySlice(&testQueryArrayString); query.err == nil {
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
	dbUtils := database.NewUtils(tests.GetDb())
	str := NewOperator(dbUtils)

	tblName, err := database.CreateTableWithName(dbUtils.Db, `id INT PRIMARY KEY, col_1 VARCHAR(100)`, "DDL_FIXED_TABLE_NAME_SIMPLE")
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	defer database.DropTable(dbUtils.Db, tblName)

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
	if err := str.Query(got).Run(); err != nil {
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
	if err := str.Query(got).Run(); err == nil {
		t.Errorf("Expected error for two rows in Query()")
	}

	// Query multiple columns
	gotArray := []TestParseSimple{}
	if err := str.QuerySlice(&gotArray).Run(); err != nil {
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
	dbUtils := database.NewUtils(tests.GetDb())
	str := NewOperator(dbUtils)

	tblName, err := database.CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY, col_2 VARCHAR(100)`,
		"DDL_FIXED_TABLE_NAME_WHERE",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	defer database.DropTable(dbUtils.Db, tblName)

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
	q := str.Query(got)
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
	q = str.QuerySlice(&gotArray)
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
	dbUtils := database.NewUtils(tests.GetDb())
	str := NewOperator(dbUtils)

	tblName, err := database.CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY, col_2 VARCHAR(100)`,
		"DDL_FIXED_TABLE_NAME_WHERE",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	defer database.DropTable(dbUtils.Db, tblName)

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
	count, err := str.Query(got).Count()
	if err != nil {
		t.Errorf("Failed to count rows: %s", err)
	}
	if count != 2 {
		t.Errorf("Missmatch of row count. Expected 2. Got %d", count)
	}

	// Test with where
	q := str.Query(got)
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
type TestParseReferenceMultiple struct {
	ID      int                         `dbColumn:"Column:id,PrimaryKey"`
	ColRef  *TestParseReferenceIncluded `dbColumn:"Column:col_ref,ForeignKey:DDL_FIXED_TABLE_NAME_REFERENCE_INCLUDED.id"`
	ColRef2 *TestParseReferenceIncluded `dbColumn:"Column:col_ref2,ForeignKey:DDL_FIXED_TABLE_NAME_REFERENCE_INCLUDED.id"`

	DbMetadata_ any `dbMetadata:"Table:DDL_FIXED_TABLE_NAME_REFERENCE"`
}
type TestParseReferenceIncluded struct {
	ID   int    `dbColumn:"Column:id,PrimaryKey"`
	Col2 string `dbColumn:"Column:col_2"`

	DbMetadata_ any `dbMetadata:"Table:DDL_FIXED_TABLE_NAME_REFERENCE_INCLUDED"`
}

// Tests a 1:1 relationship
func TestQueryReference(t *testing.T) {
	dbUtils := database.NewUtils(tests.GetDb())
	str := NewOperator(dbUtils)

	tblNameIncluded, err := database.CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY, col_2 VARCHAR(100)`,
		"DDL_FIXED_TABLE_NAME_REFERENCE_INCLUDED",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	defer database.DropTable(dbUtils.Db, tblNameIncluded)

	tblName, err := database.CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY, col_ref INT, col_ref2 INT,
		 CONSTRAINT FK_DDL_TEST_QUERY_REF FOREIGN KEY (col_ref) REFERENCES DDL_FIXED_TABLE_NAME_REFERENCE_INCLUDED(id),
		 CONSTRAINT FK_DDL_TEST_QUERY_REF2 FOREIGN KEY (col_ref2) REFERENCES DDL_FIXED_TABLE_NAME_REFERENCE_INCLUDED(id)`,
		"DDL_FIXED_TABLE_NAME_REFERENCE",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	defer database.DropTable(dbUtils.Db, tblName)

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
		fmt.Sprintf("INSERT INTO %s (id, col_ref) VALUES (?, ?)", tblName), exptected.ID, exptected.ColRef.ID,
	); err != nil {
		t.Errorf("Failed to insert data 2: %s", err)
	}

	// Resolve reference
	got := &TestParseReference{}
	if err := str.Query(got).Run(); err != nil {
		t.Errorf("Failed to select values: %s", err)
	}

	// Compare structs
	if diff := cmp.Diff(&exptected, got); diff != "" {
		t.Errorf("Mismatch of Query with reference (-want +got):\n%s", diff)
	}

	// Without resolving reference we should only receive the ID
	got = &TestParseReference{}
	if err := str.Query(got).Selector(ColumnSelector{ForeignKeyReference: false}).Run(); err != nil {
		t.Errorf("Failed to select values: %s", err)
	}

	// Compare structs
	if diff := cmp.Diff(&TestParseReferenceIncluded{ID: exptected.ColRef.ID}, got.ColRef); diff != "" {
		t.Errorf("Mismatch of Query without reference (-want +got):\n%s", diff)
	}

	// Two references (to the same table) should also work
	twoReferences := &TestParseReferenceMultiple{}
	if err := str.Query(twoReferences).Run(); err != nil {
		t.Errorf("Failed to select values: %s", err)
	}
	expectedTwoReferences := &TestParseReferenceMultiple{
		ID:      exptected.ID,
		ColRef:  exptected.ColRef,
		ColRef2: nil,
	}
	if diff := cmp.Diff(expectedTwoReferences, twoReferences); diff != "" {
		t.Errorf("Mismatch of Query with multiple reference (-want +got):\n%s", diff)
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
	dbUtils := database.NewUtils(tests.GetDb())
	str := NewOperator(dbUtils)

	tblNameIncluded, err := database.CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY, col_2 VARCHAR(100)`,
		"DDL_FIXED_TABLE_NAME_REFERENCE_INCLUDED",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	defer database.DropTable(dbUtils.Db, tblNameIncluded)

	tblName, err := database.CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY, col_ref INT, val VARCHAR(100),
		 CONSTRAINT FK_DDL_TEST_QUERY_REF FOREIGN KEY (col_ref) REFERENCES DDL_FIXED_TABLE_NAME_REFERENCE_INCLUDED(id)`,
		"DDL_FIXED_TABLE_NAME_REFERENCE",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	defer database.DropTable(dbUtils.Db, tblName)

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
		t.Errorf("%s", err.Error())
	}

	// Without resolving pointed reference we should not get any array element
	got := &TestParseReferenceIncludedOneToN{}
	if err := str.Query(got).Run(); err != nil {
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
	if err := str.Query(got).Selector(ColumnSelector{PointedKeyReference: true}).Run(); err != nil {
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
		t.Errorf("%s", err.Error())
	}

	gotArray := []TestParseReferenceIncludedOneToN{}
	expectedArray := []TestParseReferenceIncludedOneToN{*exptected, exptected3}
	if err := str.QuerySlice(&gotArray).Selector(ColumnSelector{PointedKeyReference: true}).Run(); err != nil {
		t.Errorf("Failed to select values: %s", err)
	} else {
		// Compare structs
		if diff := cmp.Diff(expectedArray, gotArray); diff != "" {
			t.Errorf("Mismatch of QuerySlice with pointed reference (-want +got):\n%s", diff)
		}
	}

	gotArray = []TestParseReferenceIncludedOneToN{}
	if err := str.QuerySlice(&gotArray).Selector(ColumnSelector{PointedKeyReference: true, PointedKeyReferenceAsync: true}).Run(); err != nil {
		t.Errorf("Failed to select ASYNC values: %s", err)
	} else {
		// Compare structs
		if diff := cmp.Diff(expectedArray, gotArray); diff != "" {
			t.Errorf("Mismatch of QuerySlice with ASYNC pointed reference (-want +got):\n%s", diff)
		}
	}
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
	dbUtils := database.NewUtils(tests.GetDb())
	str := NewOperator(dbUtils)

	tblNameIncluded, err := database.CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY`,
		"DDL_FIXED_TABLE_NAME_CUSTOM_CULUMN",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	defer database.DropTable(dbUtils.Db, tblNameIncluded)

	// Insert data
	for i := range 2 {
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
	sel := str.QuerySlice(&got).CustomColumn("", "AdditionalField", `
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
	dbUtils := database.NewUtils(tests.GetDb())
	str := NewOperator(dbUtils)

	tblNameIncluded, err := database.CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY`,
		"DDL_FIXED_TABLE_NAME_CUSTOM_CULUMN",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	defer database.DropTable(dbUtils.Db, tblNameIncluded)

	// Insert data
	for i := range 2 {
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
	sel := str.QuerySlice(&got).CustomColumn("", "AdditionalFieldEmbedded", `
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
	dbUtils := database.NewUtils(tests.GetDb())
	str := NewOperator(dbUtils)

	tblNameIncluded, err := database.CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY`,
		"DDL_FIXED_TABLE_NAME_CUSTOM_CULUMN",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	defer database.DropTable(dbUtils.Db, tblNameIncluded)

	// Insert data
	for i := range 10 {
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
	sel := str.QuerySlice(&got).Where().Column("ID", "IN", []int{2, 5}).Add()
	if err := sel.Run(); err != nil {
		t.Errorf("Failed to use IN statement: %s", err)
	}

	// Compare structs
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Mismatch of result (-want +got):\n%s", diff)
	}
}
