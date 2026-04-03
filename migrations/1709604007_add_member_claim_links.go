package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db)

		if _, err := dao.FindCollectionByNameOrId("member_claim_links"); err == nil {
			return nil
		}

		collection := &models.Collection{
			Name: "member_claim_links",
			Type: models.CollectionTypeBase,
			Schema: schema.NewSchema(
				&schema.SchemaField{
					Name:     "plan_id",
					Type:     schema.FieldTypeText,
					Required: true,
				},
				&schema.SchemaField{
					Name:     "artificial_member_id",
					Type:     schema.FieldTypeText,
					Required: true,
				},
				&schema.SchemaField{
					Name:     "token",
					Type:     schema.FieldTypeText,
					Required: true,
					Unique:   true,
					Options: &schema.TextOptions{
						Min: pointerTo(32),
						Max: pointerTo(64),
					},
				},
			),
		}

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db)

		collection, err := dao.FindCollectionByNameOrId("member_claim_links")
		if err != nil {
			return nil
		}

		return dao.DeleteCollection(collection)
	})
}
