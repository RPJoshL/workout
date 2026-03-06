package args

import (
	"testing"

	"git.rpjosh.de/RPJosh/workout/pkg/assert"
)

func TestStatementExtraction(t *testing.T) {
	t.Parallel()

	input := `
		DELIMITER $$

		CREATE SOMETHING AS
			dddd;
		fine$$

		CREATE ANOTHER
			eeee;
		hello$$

		DELIMITER ;
	`

	expected := []string{
		`CREATE SOMETHING AS
			dddd;
		fine`,
		`CREATE ANOTHER
			eeee;
		hello`,
	}

	got := getStatements(input)

	assert.Require(t, len(expected), len(got))
	for idx, expect := range expected {
		assert.Equal(t, expect, got[idx])
	}
}
