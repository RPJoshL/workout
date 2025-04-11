package database

import (
	"fmt"
	"testing"

	tests "git.rpjosh.de/RPJosh/workout/internal/tests/extra"
	"git.rpjosh.de/RPJosh/workout/pkg/assert"
)

// Tests the isolation between two transactions
func TestIsolation(t *testing.T) {
	// Create table to test isolation on
	db0 := &DB{tests.GetDb()}
	table, err := CreateTable(db0, "id INT(10)")
	if err != nil {
		t.Fatalf("failed to create test table: %s", err)
	}
	defer DropTable(db0, table)

	// Create two separate database connections
	db1, err := NewTestDB(tests.GetDb())
	if err != nil {
		t.Fatalf("failed to create database: %s", err)
	}
	defer db1.Close()
	db2, err := NewTestDB(tests.GetDb())
	if err != nil {
		t.Fatalf("failed to crate database: %s", err)
	}
	defer db2.Close()

	// Insert data (table1)
	val := 11
	_, err = db1.Exec(fmt.Sprintf("INSERT INTO %s VALUES (?)", table), val)
	if err != nil {
		t.Fatalf("failed to insert value: %s", err)
	}

	// Try to select it
	res, err := db2.Query(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE id = ?", table), val)
	if err != nil {
		t.Fatalf("failed to select value: %s", err)
	}
	res.Next()
	result := 0
	res.Scan(&result)
	res.Close()
	if result != 0 {
		t.Errorf("No isolation. Expected 0. Got %d", result)
	}
	assert.NoError(t, res.Err())

	// For db1 it has to be present
	res1, err := db1.Query(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE id = ?", table), val)
	if err != nil {
		t.Fatalf("failed to select value: %s", err)
	}
	res1.Next()
	res1.Scan(&result)
	res1.Close()
	if result != 1 {
		t.Errorf("Insert doesn't work. Expected 1. Got %d", result)
	}
	assert.NoError(t, res1.Err())
}

// Tests the rollback and commit of transactions inside transactions
func TestTransaction(t *testing.T) {
	// Create table to test isolation on
	db0 := &DB{tests.GetDb()}
	table, err := CreateTable(db0, "id INT(10)")
	if err != nil {
		t.Fatalf("failed to create test table: %s", err)
	}
	defer DropTable(db0, table)

	// Create database connection
	db1, err := NewTestDB(tests.GetDb())
	if err != nil {
		t.Fatalf("failed to create database: %s", err)
	}
	defer db1.Close()

	// Insert data
	val := 11
	_, err = db1.Exec(fmt.Sprintf("INSERT INTO %s VALUES (?)", table), val)
	if err != nil {
		t.Fatalf("failed to insert value: %s", err)
	}

	// Try to select it
	result := getValueFromTable(db1, table, t)
	if result != val {
		t.Errorf("Insert failed. Expected %d. Got %d", val, result)
	}

	// Update it within transaction
	trans, err := db1.BeginTransaction()
	if err != nil {
		t.Errorf("Failed to create transaction: %s", err)
	}
	valNew := 15
	_, err = trans.Exec(fmt.Sprintf("UPDATE %s SET id = ?", table), valNew)
	if err != nil {
		t.Fatalf("failed to insert value: %s", err)
	}

	// We expect the new value inside transaction
	result = getValueFromTable(db1, table, t)
	if result != valNew {
		t.Errorf("Insert failed. Expected %d. Got %d", valNew, result)
	}

	// After rollback we should get the old value
	if err := trans.Rollback(); err != nil {
		t.Errorf("Failed to rollback transaction: %s", err)
	}
	result = getValueFromTable(db1, table, t)
	if result != val {
		t.Errorf("Insert failed. Expected %d. Got %d", val, result)
	}

	// Update value again and commit it this time
	trans, err = db1.BeginTransaction()
	if err != nil {
		t.Errorf("Failed to create transaction: %s", err)
	}
	_, err = trans.Exec(fmt.Sprintf("UPDATE %s SET id = ?", table), valNew)
	if err != nil {
		t.Fatalf("failed to insert value: %s", err)
	}

	if err := trans.Commit(); err != nil {
		t.Errorf("Failed to commit transaction: %s", err)
	}

	// We expect the new value inside transaction
	result = getValueFromTable(db1, table, t)
	if result != valNew {
		t.Errorf("Insert failed. Expected %d. Got %d", valNew, result)
	}
}

func getValueFromTable(db SqlConnection, tbl string, t *testing.T) (result int) {
	res, err := db.Query("SELECT id FROM " + tbl)
	if err != nil {
		t.Fatalf("failed to select value: %s", err)
	}
	res.Next()
	res.Scan(&result)
	res.Close()
	assert.NoError(t, res.Err())

	return
}
