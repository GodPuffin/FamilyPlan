package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models/schema"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db)

		// Update the memberships collection to add is_artificial and name fields
		collection, err := dao.FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		// Add is_artificial field if it doesn't exist
		hasArtificialField := false
		for _, field := range collection.Schema.Fields() {
			if field.Name == "is_artificial" {
				hasArtificialField = true
				break
			}
		}

		if !hasArtificialField {
			collection.Schema.AddField(&schema.SchemaField{
				Name:     "is_artificial",
				Type:     schema.FieldTypeBool,
				Required: false, // Default to false
			})
		}

		// Add name field for artificial members if it doesn't exist
		hasNameField := false
		for _, field := range collection.Schema.Fields() {
			if field.Name == "name" {
				hasNameField = true
				break
			}
		}

		if !hasNameField {
			collection.Schema.AddField(&schema.SchemaField{
				Name:     "name",
				Type:     schema.FieldTypeText,
				Required: false,
			})
		}

		// Save the updated collection
		if err := dao.SaveCollection(collection); err != nil {
			return err
		}

		return nil
	}, func(db dbx.Builder) error {
		// Revert changes if needed
		dao := daos.New(db)

		// Remove fields from memberships
		collection, err := dao.FindCollectionByNameOrId("memberships")
		if err == nil {
			// Remove is_artificial field
			for _, field := range collection.Schema.Fields() {
				if field.Name == "is_artificial" {
					collection.Schema.RemoveField(field.Id)
					break
				}
			}

			// Remove name field
			for _, field := range collection.Schema.Fields() {
				if field.Name == "name" {
					collection.Schema.RemoveField(field.Id)
					break
				}
			}

			if err := dao.SaveCollection(collection); err != nil {
				return err
			}
		}

		return nil
	})
}
