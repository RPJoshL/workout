package dbstruct

import (
	"fmt"
	"testing"

	tests "git.rpjosh.de/RPJosh/workout/internal/tests/extra"
	"git.rpjosh.de/RPJosh/workout/pkg/database"
	"github.com/google/go-cmp/cmp"
)

func Insert1ToNReference(dbUtils *database.Utils, tblName, tblNameIncluded string, exptected TestParseReferenceIncludedOneToN) error {
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

func TestInsert(t *testing.T) {
	dbUtils := database.NewUtils(tests.GetDb())
	str := NewOperator(dbUtils)

	tblName, err := database.CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY AUTO_INCREMENT, col_2 VARCHAR(100)`,
		"DDL_FIXED_TABLE_NAME_REFERENCE_INCLUDED",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	defer database.DropTable(dbUtils.Db, tblName)

	expected := &TestParseReferenceIncludedOneToN{
		Id:   1,
		Col2: "MyValue",
	}

	// Insert a single value
	if id, err := str.Insert(expected).Run(); err != nil {
		t.Errorf("Failed to insert a single value: %s", err)
	} else {
		// Mariadb begins with 1 for auto_increment
		if id != 1 {
			t.Errorf("Got incorret ID for auto_increment: %d. Expected 1", id)
		}
	}

	// Validate with select statement
	got := &TestParseReferenceIncludedOneToN{}
	if err := str.Query(got).Run(); err != nil {
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
	if _, err := str.InsertSlice(expectedArray).Run(); err != nil {
		t.Errorf("Failed to insert multiple values: %s", err)
	}

	// Validate with select statement
	gotArray := &[]TestParseReferenceIncludedOneToN{}
	if err := str.QuerySlice(gotArray).Where().Column("id", "<>", expected.Id).Add().Run(); err != nil {
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
	dbUtils := database.NewUtils(tests.GetDb())
	str := NewOperator(dbUtils)

	tblNameIncluded, err := database.CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY AUTO_INCREMENT, col_2 VARCHAR(100)`,
		"DDL_FIXED_TABLE_NAME_REFERENCE_INCLUDED",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	if dropTable {
		defer database.DropTable(dbUtils.Db, tblNameIncluded)
	}

	tblName, err := database.CreateTableWithName(dbUtils.Db,
		`id INT PRIMARY KEY AUTO_INCREMENT, col_ref INT, val VARCHAR(100),
		 CONSTRAINT FK_DDL_TEST_QUERY_REF FOREIGN KEY (col_ref) REFERENCES DDL_FIXED_TABLE_NAME_REFERENCE_INCLUDED(id)`,
		"DDL_FIXED_TABLE_NAME_REFERENCE",
	)
	if err != nil {
		t.Fatalf("Failed to create table: %s", err)
	}
	if dropTable {
		defer database.DropTable(dbUtils.Db, tblName)
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
	if _, err := str.InsertSlice(insert).Selector(ColumnSelector{PointedKeyReference: true}).Run(); err != nil {
		t.Errorf("Failed to insert multiple values: %s", err)
	}

	// Validate with select statement
	gotArray := &[]TestParseReferenceIncludedOneToN{}
	if err := str.QuerySlice(gotArray).Selector(ColumnSelector{PointedKeyReference: true}).Run(); err != nil {
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
