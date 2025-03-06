package migrations

import (
	"fmt"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models/schema"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db)

		// Update the family_plans collection to add individual_cost field
		collection, err := dao.FindCollectionByNameOrId("family_plans")
		if err != nil {
			return err
		}

		// Add individual_cost field if it doesn't exist
		hasIndividualCostField := false
		for _, field := range collection.Schema.Fields() {
			if field.Name == "individual_cost" {
				hasIndividualCostField = true
				break
			}
		}

		if !hasIndividualCostField {
			// Add the individual_cost field to the schema
			collection.Schema.AddField(&schema.SchemaField{
				Name:     "individual_cost",
				Type:     schema.FieldTypeNumber,
				Required: true,
			})

			// Save the updated collection
			if err := dao.SaveCollection(collection); err != nil {
				return err
			}

			// Set default value for existing records using SQL
			_, err = db.NewQuery(fmt.Sprintf(`
				UPDATE %s
				SET individual_cost = 0
				WHERE individual_cost IS NULL OR individual_cost = 0
			`, collection.Name)).Execute()

			if err != nil {
				return err
			}
		}

		return nil
	}, func(db dbx.Builder) error {
		// Revert changes if needed
		dao := daos.New(db)

		// Remove individual_cost field from family_plans
		collection, err := dao.FindCollectionByNameOrId("family_plans")
		if err == nil {
			for _, field := range collection.Schema.Fields() {
				if field.Name == "individual_cost" {
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
