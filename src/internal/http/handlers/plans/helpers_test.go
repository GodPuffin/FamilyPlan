package plans

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	pbmodels "github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
)

func TestOwnerIDReturnsFirstOwner(t *testing.T) {
	t.Parallel()

	record := newTestRecord()
	record.Set("owner", []string{"owner_1", "owner_2"})

	if got := ownerID(record); got != "owner_1" {
		t.Fatalf("ownerID() = %q, want %q", got, "owner_1")
	}
}

func TestOwnerIDReturnsEmptyWhenMissing(t *testing.T) {
	t.Parallel()

	if got := ownerID(newTestRecord()); got != "" {
		t.Fatalf("ownerID() = %q, want empty string", got)
	}
}

func TestActiveMembershipCountCountsOnlyCurrentMembers(t *testing.T) {
	t.Parallel()

	active := newTestRecord()
	active.Set("leave_requested", false)

	leaving := newTestRecord()
	leaving.Set("leave_requested", true)

	ended := newTestRecord()
	ended.Set("leave_requested", false)
	ended.Set("date_ended", mustDateTime(t, time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)))

	got := activeMembershipCount([]*pbmodels.Record{active, leaving, ended})
	if got != 1 {
		t.Fatalf("activeMembershipCount() = %d, want %d", got, 1)
	}
}

func TestBuildFamilyPlanMapsRecordFields(t *testing.T) {
	t.Parallel()

	record := newTestRecord(
		&schema.SchemaField{Name: "name", Type: schema.FieldTypeText},
		&schema.SchemaField{Name: "description", Type: schema.FieldTypeText},
		&schema.SchemaField{Name: "cost", Type: schema.FieldTypeNumber},
		&schema.SchemaField{Name: "individual_cost", Type: schema.FieldTypeNumber},
		&schema.SchemaField{Name: "join_code", Type: schema.FieldTypeText},
	)
	record.Id = "plan_123"
	record.Set("name", "Streaming Bundle")
	record.Set("description", "Shared video plan")
	record.Set("cost", 19.99)
	record.Set("individual_cost", 8.25)
	record.Set("owner", []string{"owner_1"})
	record.Set("join_code", "JOIN42")
	record.Set("created", mustDateTime(t, time.Date(2026, time.March, 15, 10, 30, 0, 0, time.UTC)))

	got := buildFamilyPlan(record, 4, 12.34)

	if got.ID != "plan_123" || got.Name != "Streaming Bundle" || got.Description != "Shared video plan" {
		t.Fatalf("buildFamilyPlan() returned unexpected basic fields: %+v", got)
	}
	if got.Cost != 19.99 || got.IndividualCost != 8.25 {
		t.Fatalf("buildFamilyPlan() returned unexpected costs: %+v", got)
	}
	if got.Owner != "owner_1" || got.JoinCode != "JOIN42" {
		t.Fatalf("buildFamilyPlan() returned unexpected owner/join code: %+v", got)
	}
	if got.MembersCount != 4 || got.Balance != 12.34 {
		t.Fatalf("buildFamilyPlan() returned unexpected counts/balance: %+v", got)
	}
	if got.CreatedAt == "" {
		t.Fatalf("buildFamilyPlan() returned empty CreatedAt: %+v", got)
	}
}

func TestBuildPaymentFormatsDateFields(t *testing.T) {
	t.Parallel()

	record := newTestRecord(
		&schema.SchemaField{Name: "plan_id", Type: schema.FieldTypeText},
		&schema.SchemaField{Name: "user_id", Type: schema.FieldTypeText},
		&schema.SchemaField{Name: "amount", Type: schema.FieldTypeNumber},
		&schema.SchemaField{Name: "status", Type: schema.FieldTypeText},
		&schema.SchemaField{Name: "notes", Type: schema.FieldTypeText},
	)
	record.Id = "payment_123"
	record.Set("plan_id", "plan_123")
	record.Set("user_id", "user_456")
	record.Set("amount", 15.5)
	record.Set("date", mustDateTime(t, time.Date(2026, time.April, 1, 12, 0, 0, 0, time.UTC)))
	record.Set("status", "approved")
	record.Set("notes", "paid")
	record.Set("for_month", mustDateTime(t, time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)))

	got := buildPayment(record, "marcus", "Marcus")

	if got.ID != "payment_123" || got.PlanID != "plan_123" || got.UserID != "user_456" {
		t.Fatalf("buildPayment() returned unexpected ids: %+v", got)
	}
	if got.Amount != 15.5 || got.Date != "2026-04-01" || got.ForMonth != "2026-03" {
		t.Fatalf("buildPayment() returned unexpected date data: %+v", got)
	}
	if got.Status != "approved" || got.Notes != "paid" || got.Username != "marcus" || got.Name != "Marcus" {
		t.Fatalf("buildPayment() returned unexpected metadata: %+v", got)
	}
}

func TestFormatForMonthReturnsEmptyForZeroDate(t *testing.T) {
	t.Parallel()

	if got := formatForMonth(newTestRecord()); got != "" {
		t.Fatalf("formatForMonth() = %q, want empty string", got)
	}
}

func TestRedirectToPlanUsesSeeOther(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/join", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := redirectToPlan(c, "JOIN42"); err != nil {
		t.Fatalf("redirectToPlan returned error: %v", err)
	}

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if location := rec.Header().Get("Location"); location != "/JOIN42" {
		t.Fatalf("Location = %q, want %q", location, "/JOIN42")
	}
}

func newTestRecord(fields ...*schema.SchemaField) *pbmodels.Record {
	collection := &pbmodels.Collection{
		Name:   "test_collection",
		Type:   pbmodels.CollectionTypeBase,
		Schema: schema.NewSchema(fields...),
	}

	return pbmodels.NewRecord(collection)
}

func mustDateTime(t *testing.T, value time.Time) types.DateTime {
	t.Helper()

	dt, err := types.ParseDateTime(value)
	if err != nil {
		t.Fatalf("ParseDateTime(%v) returned error: %v", value, err)
	}

	return dt
}
