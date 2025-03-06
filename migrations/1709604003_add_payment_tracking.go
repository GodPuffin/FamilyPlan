package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
)

// The pointerTo function is available from the main package but we need one here
func int_pointer(i int) *int {
	return &i
}

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db)

		// Check if payments collection already exists
		_, err := dao.FindCollectionByNameOrId("payments")
		if err != nil {
			// Collection doesn't exist, create it
			paymentsCollection := &models.Collection{
				Name:       "payments",
				Type:       models.CollectionTypeBase,
				ListRule:   nil,
				ViewRule:   nil,
				CreateRule: nil,
				UpdateRule: nil,
				DeleteRule: nil,
				Schema: schema.NewSchema(
					&schema.SchemaField{
						Name:     "plan_id",
						Type:     schema.FieldTypeRelation,
						Required: true,
						Options: &schema.RelationOptions{
							CollectionId:  "family_plans",
							MaxSelect:     int_pointer(1),
							CascadeDelete: true,
						},
					},
					&schema.SchemaField{
						Name:     "user_id",
						Type:     schema.FieldTypeRelation,
						Required: true,
						Options: &schema.RelationOptions{
							CollectionId:  "_pb_users_auth_",
							MaxSelect:     int_pointer(1),
							CascadeDelete: false,
						},
					},
					&schema.SchemaField{
						Name:     "amount",
						Type:     schema.FieldTypeNumber,
						Required: true,
					},
					&schema.SchemaField{
						Name:     "date",
						Type:     schema.FieldTypeDate,
						Required: true,
					},
					&schema.SchemaField{
						Name:     "status",
						Type:     schema.FieldTypeSelect,
						Required: true,
						Options: &schema.SelectOptions{
							Values:    []string{"pending", "approved", "rejected"},
							MaxSelect: 1,
						},
					},
					&schema.SchemaField{
						Name:     "notes",
						Type:     schema.FieldTypeText,
						Required: false,
					},
					&schema.SchemaField{
						Name:     "for_month",
						Type:     schema.FieldTypeDate,
						Required: false,
					},
				),
			}

			// Save the new payments collection
			if err := dao.SaveCollection(paymentsCollection); err != nil {
				return err
			}
		}

		// Update the memberships collection to add leave_requested
		membershipsCollection, err := dao.FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		// Add leave_requested field if it doesn't exist
		hasLeaveField := false
		for _, field := range membershipsCollection.Schema.Fields() {
			if field.Name == "leave_requested" {
				hasLeaveField = true
				break
			}
		}

		if !hasLeaveField {
			membershipsCollection.Schema.AddField(&schema.SchemaField{
				Name:     "leave_requested",
				Type:     schema.FieldTypeBool,
				Required: false,
			})

			// Save the updated memberships collection
			if err := dao.SaveCollection(membershipsCollection); err != nil {
				return err
			}
		}

		return nil
	}, func(db dbx.Builder) error {
		// Revert changes if needed
		dao := daos.New(db)

		// Remove leave_requested field from memberships
		memberships, err := dao.FindCollectionByNameOrId("memberships")
		if err == nil {
			for _, field := range memberships.Schema.Fields() {
				if field.Name == "leave_requested" {
					memberships.Schema.RemoveField(field.Id)
					break
				}
			}
			if err := dao.SaveCollection(memberships); err != nil {
				return err
			}
		}

		return nil
	})
}
