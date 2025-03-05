package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
)

// Helper function to create a pointer to an int
func pointerTo(i int) *int {
	return &i
}

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db)

		// Create the users collection if it doesn't exist
		_, err := dao.FindCollectionByNameOrId("users")
		if err != nil {
			usersCollection := &models.Collection{
				Name:       "users",
				Type:       models.CollectionTypeAuth,
				ListRule:   nil,
				ViewRule:   nil,
				CreateRule: nil,
				UpdateRule: nil,
				DeleteRule: nil,
				Schema: schema.NewSchema(
					&schema.SchemaField{
						Name:     "username",
						Type:     schema.FieldTypeText,
						Required: true,
						Options: &schema.TextOptions{
							Min:     pointerTo(3),
							Max:     pointerTo(150),
							Pattern: "",
						},
					},
				),
			}

			if err := dao.SaveCollection(usersCollection); err != nil {
				return err
			}
		}

		// Create the family_plans collection if it doesn't exist
		_, err = dao.FindCollectionByNameOrId("family_plans")
		if err != nil {
			plansCollection := &models.Collection{
				Name:       "family_plans",
				Type:       models.CollectionTypeBase,
				ListRule:   nil,
				ViewRule:   nil,
				CreateRule: nil,
				UpdateRule: nil,
				DeleteRule: nil,
				Schema: schema.NewSchema(
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
						Options: &schema.TextOptions{
							Min:     pointerTo(6),
							Max:     pointerTo(10),
							Pattern: "",
						},
					},
				),
			}

			if err := dao.SaveCollection(plansCollection); err != nil {
				return err
			}
		}

		// Create the memberships collection if it doesn't exist
		_, err = dao.FindCollectionByNameOrId("memberships")
		if err != nil {
			// Set rules as pointers to strings
			listRule := "@request.auth.id != ''"
			viewRule := "@request.auth.id != ''"
			createRule := "@request.auth.id != ''"
			updateRule := "@request.auth.id = user_id"
			deleteRule := "@request.auth.id = user_id"

			membershipsCollection := &models.Collection{
				Name:       "memberships",
				Type:       models.CollectionTypeBase,
				ListRule:   &listRule,
				ViewRule:   &viewRule,
				CreateRule: &createRule,
				UpdateRule: &updateRule,
				DeleteRule: &deleteRule,
				Schema: schema.NewSchema(
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
				),
			}

			if err := dao.SaveCollection(membershipsCollection); err != nil {
				return err
			}
		}

		// Create the join_requests collection if it doesn't exist
		_, err = dao.FindCollectionByNameOrId("join_requests")
		if err != nil {
			// Set rules as pointers to strings
			listRule := "@request.auth.id != ''"
			viewRule := "@request.auth.id != ''"
			createRule := "@request.auth.id != ''"
			updateRule := "@request.auth.id = user_id"
			deleteRule := "@request.auth.id = user_id"

			joinRequestsCollection := &models.Collection{
				Name:       "join_requests",
				Type:       models.CollectionTypeBase,
				ListRule:   &listRule,
				ViewRule:   &viewRule,
				CreateRule: &createRule,
				UpdateRule: &updateRule,
				DeleteRule: &deleteRule,
				Schema: schema.NewSchema(
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
				),
			}

			if err := dao.SaveCollection(joinRequestsCollection); err != nil {
				return err
			}
		}

		return nil
	}, func(db dbx.Builder) error {
		// Revert operation - would drop the collections, but this could be dangerous
		// so we're not implementing the revert for this migration
		return nil
	})
}
