package database

import (
	"os"
	"testing"
)

func TestDatabase(t *testing.T) {
	// Create the database handler
	database, err := NewDatabase()
	if err != nil {
		t.Fatal(err)
	}

	// Remove any existing data now and when done
	os.RemoveAll(database.Path)
	defer os.RemoveAll(database.Path)

	// Open the database
	if err := database.Open(); err != nil {
		t.Fatal(err)
	}

	// Store data in the database
	id, err := database.Set("test_data", 0, map[string]interface{}{
		"someKey":    "someValue",
		"anotherKey": "anotherValue"})
	if err != nil {
		t.Fatal(id, err)
	}

	// Modify the existing object to verify that partial changes work as expected
	if _, err := database.Set("test_data", id, map[string]interface{}{
		"someKey":  "someValueModified",
		"thirdKey": "thirdValue"}); err != nil {
		t.Fatal(id, err)
	}

	// Read data from the database
	object, err := database.Get("test_data", id)
	if err != nil {
		t.Fatal(object, err)
	}

	// Attempt to query for the inserted data
	results, err := database.Query("test_data", `[{"eq": "someValueModified", "in": ["someKey"]}, {"eq": "anotherValue", "in": ["anotherKey"]}, {"eq": "thirdValue", "in": ["thirdKey"]}]`)
	if err != nil {
		t.Fatal(results, err)
	}

	// Verify that we have results
	if len(results) <= 0 {
		t.Fatal(results, "Query returned empty results")
	}

	// Close the database
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}
}
