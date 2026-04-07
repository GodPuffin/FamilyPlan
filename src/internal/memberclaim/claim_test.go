package memberclaim

import (
	"testing"
	"time"

	"familyplan/src/internal/planutil"

	_ "familyplan/migrations"

	"github.com/pocketbase/pocketbase"
	pbmigrations "github.com/pocketbase/pocketbase/migrations"
	pbmodels "github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/migrate"
	"github.com/pocketbase/pocketbase/tools/types"
)

func TestTransferArtificialMembershipPreservesCreated(t *testing.T) {
	app := newMigratedTestApp(t)

	owner := saveTestUser(t, app, "owner")
	realUser := saveTestUser(t, app, "real")
	plan := saveTestPlan(t, app, owner.Id)

	artificialMemberID := "placeholder-member"
	artificialCreated := mustDateTime(t, time.Date(2026, time.January, 15, 10, 30, 0, 0, time.UTC))
	saveTestMembership(t, app, plan.Id, artificialMemberID, true, artificialCreated)

	if err := TransferArtificialMembership(app.Dao(), plan, artificialMemberID, realUser.Id); err != nil {
		t.Fatalf("TransferArtificialMembership returned error: %v", err)
	}

	membership, err := planutil.FindMembershipWithDao(app.Dao(), plan.Id, realUser.Id)
	if err != nil {
		t.Fatalf("FindMembershipWithDao returned error: %v", err)
	}
	if membership == nil {
		t.Fatal("expected transferred membership to exist")
	}

	if got := membership.GetDateTime("created").String(); got != artificialCreated.String() {
		t.Fatalf("transferred membership created = %q, want %q", got, artificialCreated.String())
	}
}

func newMigratedTestApp(t *testing.T) *pocketbase.PocketBase {
	t.Helper()

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir: t.TempDir(),
	})

	if err := app.Bootstrap(); err != nil {
		t.Fatalf("failed to bootstrap app: %v", err)
	}

	runner, err := migrate.NewRunner(app.DB(), pbmigrations.AppMigrations)
	if err != nil {
		t.Fatalf("failed to create migrations runner: %v", err)
	}
	if _, err := runner.Up(); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	if err := app.Bootstrap(); err != nil {
		t.Fatalf("failed to refresh app after migrations: %v", err)
	}

	t.Cleanup(func() {
		if err := app.ResetBootstrapState(); err != nil {
			t.Fatalf("failed to reset app bootstrap state: %v", err)
		}
	})

	return app
}

func saveTestUser(t *testing.T, app *pocketbase.PocketBase, username string) *pbmodels.Record {
	t.Helper()

	collection, err := app.Dao().FindCollectionByNameOrId("users")
	if err != nil {
		t.Fatalf("failed to find users collection: %v", err)
	}

	record := pbmodels.NewRecord(collection)
	record.Set("username", username)
	if err := record.SetPassword("password123"); err != nil {
		t.Fatalf("failed to set password for user %q: %v", username, err)
	}
	if err := record.SetVerified(true); err != nil {
		t.Fatalf("failed to verify user %q: %v", username, err)
	}
	if err := app.Dao().SaveRecord(record); err != nil {
		t.Fatalf("failed to save user %q: %v", username, err)
	}

	return record
}

func saveTestPlan(t *testing.T, app *pocketbase.PocketBase, ownerID string) *pbmodels.Record {
	t.Helper()

	collection, err := app.Dao().FindCollectionByNameOrId("family_plans")
	if err != nil {
		t.Fatalf("failed to find family_plans collection: %v", err)
	}

	record := pbmodels.NewRecord(collection)
	record.Set("name", "Test Family")
	record.Set("description", "")
	record.Set("cost", 1000)
	record.Set("individual_cost", 0)
	record.Set("owner", ownerID)
	record.Set("join_code", "ABC123")
	if err := app.Dao().SaveRecord(record); err != nil {
		t.Fatalf("failed to save family plan: %v", err)
	}

	return record
}

func saveTestMembership(t *testing.T, app *pocketbase.PocketBase, planID, userID string, isArtificial bool, created types.DateTime) *pbmodels.Record {
	t.Helper()

	collection, err := app.Dao().FindCollectionByNameOrId("memberships")
	if err != nil {
		t.Fatalf("failed to find memberships collection: %v", err)
	}

	record := pbmodels.NewRecord(collection)
	record.Set("plan_id", planID)
	record.Set("user_id", userID)
	record.Set("is_artificial", isArtificial)
	record.Set("name", "Placeholder")
	record.Set("created", created)
	if err := app.Dao().SaveRecord(record); err != nil {
		t.Fatalf("failed to save membership: %v", err)
	}

	return record
}

func mustDateTime(t *testing.T, value time.Time) types.DateTime {
	t.Helper()

	result, err := types.ParseDateTime(value)
	if err != nil {
		t.Fatalf("failed to parse datetime: %v", err)
	}

	return result
}
