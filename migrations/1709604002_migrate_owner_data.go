package migrations

import (
	"fmt"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db)

		// Check if the family_plans collection exists
		collection, err := dao.FindCollectionByNameOrId("family_plans")
		if err != nil {
			// Collection doesn't exist, nothing to migrate
			return nil
		}

		// Direct SQL query to update the data format for the relation field
		// This is a workaround for the FindRecordsByFilter issue
		_, err = db.NewQuery(fmt.Sprintf(`
			UPDATE %s
			SET owner = json_array(owner)
			WHERE json_type(owner) = 'text'
		`, collection.Name)).Execute()

		return err
	}, func(db dbx.Builder) error {
		// No revert needed
		return nil
	})
}
