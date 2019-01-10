package database

import (
	"encoding/json"
	"log"
	"os"

	"github.com/HouzuoGuo/tiedot/db"
)

// https://github.com/HouzuoGuo/tiedot/blob/master/examples/example.go

// DatabasePath is the directory where the database files are stored
const DatabasePath = ".db"

// Database is an abstraction for the backing database
type Database struct {
	Client *db.DB
	Path   string
}

// NewDatabase creates and returns a new instance of Database
func NewDatabase() (*Database, error) {
	database := &Database{}

	// Get the database path
	dbPath, err := getDatabasePath()
	if err != nil {
		return nil, err
	}
	database.Path = dbPath

	return database, nil
}

// Open the database
func (database *Database) Open() error {
	log.Println("Opening database:", database.Path)

	// Open the database (creates a new one if it doesn't yet exist)
	newDB, err := db.OpenDB(database.Path)
	if err != nil {
		return err
	}
	database.Client = newDB

	return nil
}

// Close the database
func (database *Database) Close() error {
	log.Println("Closing the database..")
	return database.Client.Close()
}

// Get fetches an object from the database
func (database *Database) Get(collection string, objectID int) (map[string]interface{}, error) {
	// Switch to the collection
	objects, collectionErr := database.GetCollection(collection)
	if collectionErr != nil {
		return nil, collectionErr
	}

	// Get the object from the collection
	object, err := objects.Read(objectID)
	if err != nil {
		return nil, err
	}

	// Return the resulting object
	return object, nil
}

// TODO: Implement automatic "scrubbing", which both repairs and compresses a collection?
// NOTE: Scrubbing invalidates all collection references!

// Set stores an object in the database
func (database *Database) Set(collection string, objectID int, value map[string]interface{}) (int, error) {
	// Switch to the collection
	objects, collectionErr := database.GetCollection(collection)
	if collectionErr != nil {
		return 0, collectionErr
	}

	// Check if the object already exists
	existingObject, existingObjectErr := objects.Read(objectID)
	if existingObjectErr != nil {
		// Object doesn't yet exist, so we can safely create it
		//log.Println("Creating a new object")
		id, err := objects.Insert(value)
		if err != nil {
			return 0, err
		}
		objectID = id
	} else {
		// Object already exists, so we should only update changes
		//log.Println("Updating an existing object")
		for k, v := range existingObject {
			if _, exists := value[k]; !exists {
				value[k] = v
			}
		}
		err := objects.Update(objectID, value)
		if err != nil {
			return 0, err
		}
	}

	// Create an index array
	indexes := make([]string, 0)
	for k := range value {
		indexes = append(indexes, k)
	}

	// Update database indexes
	if err := database.createIndexes(objects, indexes); err != nil {
		return 0, err
	}

	// Return the object ID on success
	return objectID, nil
}

// Query the database directly
func (database *Database) Query(collection string, query string) (map[int]map[string]interface{}, error) {
	//log.Println("Executing query:", query)

	// Switch to the collection
	objects, collectionErr := database.GetCollection(collection)
	if collectionErr != nil {
		return nil, collectionErr
	}

	// Convert the query string to a query object
	var queryObject interface{}
	if err := json.Unmarshal([]byte(query), &queryObject); err != nil {
		return nil, err
	}

	// Prepare the query results
	queryResults := make(map[int]struct{}) // query result (document IDs) goes into map keys

	// Run the query
	if err := db.EvalQuery(queryObject, objects, &queryResults); err != nil {
		return nil, err
	}

	// Construct a pre-populated map of the result objects
	results := make(map[int]map[string]interface{})
	for id := range queryResults {
		o, err := objects.Read(id)
		if err != nil {
			return nil, err
		}
		results[id] = o
	}

	// Return the query results
	return results, nil
}

// GetCollection returns a reference to the collection object
func (database *Database) GetCollection(collection string) (*db.Col, error) {
	// Make sure the collection exists first
	if err := database.createCollection(collection); err != nil {
		return nil, err
	}

	// Switch to the collection
	objects := database.Client.Use(collection)

	// Return the collection
	return objects, nil
}

func (database *Database) createCollection(collection string) error {
	// Check if the collection already exists
	exists := false
	for _, name := range database.Client.AllCols() {
		if name == collection {
			exists = true
		}
	}

	// Create the collection if it doesn't yet exist
	if !exists {
		if err := database.Client.Create(collection); err != nil {
			return err
		}
	}

	// Return nil on success
	return nil
}

func (database *Database) createIndexes(collection *db.Col, indexes []string) error {
	// Add indexes individually, skipping on error
	for _, index := range indexes {
		if err := collection.Index([]string{index}); err != nil {
			continue
		}
		//log.Println("Added new index for key:", index)
	}

	// Return nil on success
	return nil
}

func getDatabasePath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if cwd == "/" {
		return cwd + DatabasePath, nil
	}
	//return cwd + "/../" + DatabasePath, nil
	return cwd + "/" + DatabasePath, nil
}
