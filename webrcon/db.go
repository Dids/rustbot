package webrcon

import (
	"errors"

	"github.com/Dids/rustbot/database"
)

func incrementKillCount(database *database.Database, killerID string) error {
	return incrementFieldForSteamID(database, "Kills", killerID)
}

func incrementDeathCount(database *database.Database, victimID string) error {
	return incrementFieldForSteamID(database, "Deaths", victimID)
}

func incrementFieldForSteamID(database *database.Database, field string, steamID string) error {
	if database == nil || database.Client == nil {
		return errors.New("Database is nil")
	}
	if len(field) <= 0 {
		return errors.New("field is nil or invalid")
	}
	if len(steamID) <= 0 {
		return errors.New("steamID is nil or invalid")
	}

	// Find the matching user
	user := make(map[string]interface{})
	matches, err := database.Query("users", `[{"eq": "`+steamID+`", "in": ["SteamID"]}]`)
	if err != nil {
		return err
	}

	// Create the object id (or use the existing one, if available)
	objectID := 0
	for id := range matches {
		objectID = id
		break
	}

	// Get the existing user object or create a new one
	if len(matches) > 0 {
		user = matches[objectID]
	} else {
		user = map[string]interface{}{
			"SteamID": steamID,
			field:     0,
		}
	}

	// Verify that the user object is valid
	if user == nil || len(user) <= 0 {
		return errors.New("User is nil or invalid, cannot increment field: " + field)
	}

	// Increment the field (with a hack that accounts for JSON unmarshaling converting ints to floats)
	floatValue, ok := user[field].(float64)
	if ok {
		intValue := int(floatValue)
		user[field] = intValue
	}

	// Update the user in the database
	if _, err := database.Set("users", objectID, user); err != nil {
		return err
	}

	return nil
}
