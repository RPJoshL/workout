package dbstruct

import (
	"testing"

	tests "git.rpjosh.de/RPJosh/workout/internal/tests/extra"
	"git.rpjosh.de/RPJosh/workout/pkg/database"
	"github.com/google/go-cmp/cmp"
)

func TestUpdate(t *testing.T) {
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
	if _, err := str.Insert(expected).Run(); err != nil {
		t.Fatalf("Failed to insert a single value: %s", err)
	}

	// Update the value
	expected.Col2 = "Updated value!"
	if err := str.Update(expected).Run(); err != nil {
		t.Errorf("Failed to update value: %s", err)
	}

	got := &TestParseReferenceIncludedOneToN{}
	if err := str.Query(got).Run(); err != nil {
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
	if _, err := str.InsertSlice(expectedArray).Run(); err != nil {
		t.Fatalf("Failed to insert multiple values: %s", err)
	}

	// Update thema again
	(*expectedArray)[0].Col2 = "And i'm the winner"
	(*expectedArray)[1].Col2 = "And you not"
	if err := str.UpdateSlice(expectedArray).Run(); err != nil {
		t.Errorf("Failed to update multiple values: %s", err)
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

// Tests the update of a n:1 relationship
func TestUpdateOneToN(t *testing.T) {
	dbUtils := database.NewUtils(tests.GetDb())
	str := NewOperator(dbUtils)

	// Create basic data
	tblNames := testInsertOneToN(t, false)
	defer func() {
		for _, t := range tblNames {
			database.DropTable(dbUtils.Db, t)
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
	if err := str.UpdateSlice(&newData).Selector(ColumnSelector{PointedKeyReference: true}).Run(); err != nil {
		t.Errorf("Failed to update 1:n relationship: %s", err)
	}
	if err := str.QuerySlice(gotArray).Selector(ColumnSelector{PointedKeyReference: true}).Run(); err != nil {
		t.Errorf("Failed to select data for test: %s", err)
	}

	// Compare
	if diff := cmp.Diff(&newData, gotArray); diff != "" {
		t.Errorf("Mismatch of multiple insert with reference (-want +got):\n%s", diff)
	}
}
