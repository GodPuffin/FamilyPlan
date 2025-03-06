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

		// Add date_ended field to memberships
		collection, err := dao.FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		// Check if the field already exists
		hasDateEndedField := false
		for _, field := range collection.Schema.Fields() {
			if field.Name == "date_ended" {
				hasDateEndedField = true
				break
			}
		}

		// Add the field if it doesn't exist
		if !hasDateEndedField {
			dateEndedField := &schema.SchemaField{
				Name:     "date_ended",
				Type:     schema.FieldTypeDate,
				Required: false,
			}

			collection.Schema.AddField(dateEndedField)

			if err := dao.SaveCollection(collection); err != nil {
				return err
			}
		}

		return nil
	}, func(db dbx.Builder) error {
		dao := daos.New(db)

		// Revert changes (remove date_ended field)
		collection, err := dao.FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		// Check if the field exists
		hasField := false
		for _, field := range collection.Schema.Fields() {
			if field.Name == "date_ended" {
				hasField = true
				collection.Schema.RemoveField(field.Id)
				break
			}
		}

		if hasField {
			if err := dao.SaveCollection(collection); err != nil {
				return err
			}
		}

		return nil
	})
}
