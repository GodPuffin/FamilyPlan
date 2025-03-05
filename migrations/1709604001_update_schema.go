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

		// Update family_plans collection schema
		collection, err := dao.FindCollectionByNameOrId("family_plans")
		if err != nil {
			return err
		}

		// Remove service field
		for _, field := range collection.Schema.Fields() {
			if field.Name == "service" {
				collection.Schema.RemoveField(field.Id)
				break
			}
		}

		// Update owner field to be a relation
		for _, field := range collection.Schema.Fields() {
			if field.Name == "owner" {
				// Create a new relation field
				newField := collection.Schema.GetFieldByName("owner")
				newField.Type = schema.FieldTypeRelation
				newField.Options = &schema.RelationOptions{
					CollectionId:  "_pb_users_auth_",
					MaxSelect:     pointerTo(1),
					CascadeDelete: false,
				}
				break
			}
		}

		// Save the updated collection
		if err := dao.SaveCollection(collection); err != nil {
			return err
		}

		// We're not updating existing records here because
		// changing field types requires a different approach.
		// PocketBase will attempt to migrate the data automatically
		// when the field type changes.

		return nil
	}, func(db dbx.Builder) error {
		// Revert changes if needed
		return nil
	})
}
