package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
)

// Helper function to create a pointer to an int
func pointerTo(i int) *int {
	return &i
}

// InitDatabase initializes the PocketBase database
func InitDatabase(app *pocketbase.PocketBase) error {
	// Set up the data directory
	dataDir := filepath.Join(".", "pb_data")
	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		return err
	}

	// Bootstrap the app
	if err := app.Bootstrap(); err != nil {
		return err
	}

	// Check if migrations are enabled
	migrationsEnabled := false
	for _, arg := range os.Args {
		if arg == "migrate" {
			migrationsEnabled = true
			break
		}
	}

	// Only initialize collections if migrations are not enabled
	if !migrationsEnabled {
		// Initialize collections
		if err := InitCollections(app); err != nil {
			return err
		}
		log.Println("Database initialized successfully!")
	} else {
		log.Println("Skipping manual collection initialization as migrations are enabled")
	}

	return nil
}

// InitCollections ensures all necessary collections exist in the database
func InitCollections(app *pocketbase.PocketBase) error {
	// Initialize the family_plans collection
	if err := ensureFamilyPlansCollection(app); err != nil {
		return err
	}

	// Initialize the memberships collection
	if err := ensureMembershipsCollection(app); err != nil {
		return err
	}

	// Initialize the join_requests collection
	if err := ensureJoinRequestsCollection(app); err != nil {
		return err
	}

	log.Println("All collections initialized successfully!")
	return nil
}

// ensureFamilyPlansCollection creates the family_plans collection if it doesn't exist
func ensureFamilyPlansCollection(app *pocketbase.PocketBase) error {
	collection, err := app.Dao().FindCollectionByNameOrId("family_plans")
	if err == nil {
		// Collection already exists
		return nil
	}

	// Create collection
	collection = &models.Collection{
		Name: "family_plans",
		Type: models.CollectionTypeBase,
	}

	// Set rules as pointers to strings
	createRule := "@request.auth.id != ''"
	updateRule := "@request.auth.id = owner"
	deleteRule := "@request.auth.id = owner"

	collection.CreateRule = &createRule
	collection.UpdateRule = &updateRule
	collection.DeleteRule = &deleteRule

	// Define schema fields
	collection.Schema = schema.NewSchema(
		&schema.SchemaField{
			Name:     "name",
			Type:     schema.FieldTypeText,
			Required: true,
		},
		&schema.SchemaField{
			Name:     "description",
			Type:     schema.FieldTypeText,
			Required: false,
		},
		&schema.SchemaField{
			Name:     "cost",
			Type:     schema.FieldTypeNumber,
			Required: true,
		},
		&schema.SchemaField{
			Name:     "owner",
			Type:     schema.FieldTypeRelation,
			Required: true,
			Options: &schema.RelationOptions{
				CollectionId:  "_pb_users_auth_",
				MaxSelect:     pointerTo(1),
				CascadeDelete: false,
			},
		},
		&schema.SchemaField{
			Name:     "join_code",
			Type:     schema.FieldTypeText,
			Required: true,
			Unique:   true,
		},
	)

	// Save the collection
	if err := app.Dao().SaveCollection(collection); err != nil {
		return err
	}

	log.Println("Created family_plans collection")
	return nil
}

// ensureMembershipsCollection creates the memberships collection if it doesn't exist
func ensureMembershipsCollection(app *pocketbase.PocketBase) error {
	collection, err := app.Dao().FindCollectionByNameOrId("memberships")
	if err == nil {
		// Collection already exists
		return nil
	}

	// Create collection
	collection = &models.Collection{
		Name: "memberships",
		Type: models.CollectionTypeBase,
	}

	// Set rules as pointers to strings
	listRule := "@request.auth.id != ''"
	viewRule := "@request.auth.id != ''"
	createRule := "@request.auth.id != ''"
	updateRule := "@request.auth.id = user_id"
	deleteRule := "@request.auth.id = user_id"

	collection.ListRule = &listRule
	collection.ViewRule = &viewRule
	collection.CreateRule = &createRule
	collection.UpdateRule = &updateRule
	collection.DeleteRule = &deleteRule

	// Define schema fields
	collection.Schema = schema.NewSchema(
		&schema.SchemaField{
			Name:     "plan_id",
			Type:     schema.FieldTypeText,
			Required: true,
		},
		&schema.SchemaField{
			Name:     "user_id",
			Type:     schema.FieldTypeText,
			Required: true,
		},
	)

	// Save the collection
	if err := app.Dao().SaveCollection(collection); err != nil {
		return err
	}

	log.Println("Created memberships collection")
	return nil
}

// ensureJoinRequestsCollection creates the join_requests collection if it doesn't exist
func ensureJoinRequestsCollection(app *pocketbase.PocketBase) error {
	collection, err := app.Dao().FindCollectionByNameOrId("join_requests")
	if err == nil {
		// Collection already exists
		return nil
	}

	// Create collection
	collection = &models.Collection{
		Name: "join_requests",
		Type: models.CollectionTypeBase,
	}

	// Set rules as pointers to strings
	listRule := "@request.auth.id != ''"
	viewRule := "@request.auth.id != ''"
	createRule := "@request.auth.id != ''"
	updateRule := "@request.auth.id = user_id"
	deleteRule := "@request.auth.id = user_id"

	collection.ListRule = &listRule
	collection.ViewRule = &viewRule
	collection.CreateRule = &createRule
	collection.UpdateRule = &updateRule
	collection.DeleteRule = &deleteRule

	// Define schema fields
	collection.Schema = schema.NewSchema(
		&schema.SchemaField{
			Name:     "plan_id",
			Type:     schema.FieldTypeText,
			Required: true,
		},
		&schema.SchemaField{
			Name:     "user_id",
			Type:     schema.FieldTypeText,
			Required: true,
		},
	)

	// Save the collection
	if err := app.Dao().SaveCollection(collection); err != nil {
		return err
	}

	log.Println("Created join_requests collection")
	return nil
}
