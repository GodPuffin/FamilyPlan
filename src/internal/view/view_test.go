package view

import (
	"bytes"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"familyplan/src/internal/domain"

	"github.com/labstack/echo/v5"
)

func TestRenderPageUsesSessionDefaults(t *testing.T) {
	resetTemplateCache()
	t.Cleanup(resetTemplateCache)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", domain.SessionData{
		IsAuthenticated: true,
		Username:        "alice",
	})

	data := map[string]interface{}{
		"title": "Home",
	}
	if err := RenderPage(c, "index.html", data); err != nil {
		t.Fatalf("RenderPage returned error: %v", err)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Welcome back, alice!") {
		t.Fatalf("expected rendered body to include session username, got %q", body)
	}
	if !strings.Contains(body, "View My Family Plans") {
		t.Fatalf("expected authenticated body, got %q", body)
	}
	if data["isAuthenticated"] != true || data["username"] != "alice" {
		t.Fatalf("RenderPage did not populate session defaults: %+v", data)
	}
}

func TestRenderPagePreservesProvidedValues(t *testing.T) {
	resetTemplateCache()
	t.Cleanup(resetTemplateCache)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("session", domain.SessionData{
		IsAuthenticated: true,
		Username:        "alice",
		Name:            "Session Name",
	})

	data := map[string]interface{}{
		"title": "Home",
		"name":  "Override Name",
	}
	if err := RenderPage(c, "index.html", data); err != nil {
		t.Fatalf("RenderPage returned error: %v", err)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Welcome back, Override Name!") {
		t.Fatalf("expected provided name to win over session default, got %q", body)
	}
	if data["name"] != "Override Name" {
		t.Fatalf("RenderPage overwrote provided name: %+v", data)
	}
}

func TestRenderPageReturnsTemplateErrorForMissingPage(t *testing.T) {
	resetTemplateCache()
	t.Cleanup(resetTemplateCache)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := RenderPage(c, "missing.html", map[string]interface{}{"title": "Missing"}); err == nil {
		t.Fatal("expected RenderPage to fail for missing template")
	}
}

func TestLoadTemplateCachesTemplates(t *testing.T) {
	resetTemplateCache()
	t.Cleanup(resetTemplateCache)

	first, err := loadTemplate("index.html")
	if err != nil {
		t.Fatalf("first loadTemplate call returned error: %v", err)
	}

	second, err := loadTemplate("index.html")
	if err != nil {
		t.Fatalf("second loadTemplate call returned error: %v", err)
	}

	if first != second {
		t.Fatal("expected loadTemplate to reuse cached template")
	}
}

func TestTemplateFuncsHandleFormattingAndMath(t *testing.T) {
	t.Parallel()

	formatMoney := Funcs["formatMoney"].(func(float64) string)
	slice := Funcs["slice"].(func(string, int, int) string)
	div := Funcs["div"].(func(interface{}, interface{}) float64)
	mul := Funcs["mul"].(func(float64, float64) float64)
	sub := Funcs["sub"].(func(float64, float64) float64)
	toFloat64 := Funcs["float64"].(func(int) float64)

	if got := formatMoney(12.345); got != "$12.35" {
		t.Fatalf("formatMoney() = %q, want %q", got, "$12.35")
	}
	if got := slice("abcdef", 1, 4); got != "bcd" {
		t.Fatalf("slice() = %q, want %q", got, "bcd")
	}
	if got := slice("abc", 5, 8); got != "" {
		t.Fatalf("slice() = %q, want empty string", got)
	}
	if got := div(9, 2); got != 4.5 {
		t.Fatalf("div() = %v, want %v", got, 4.5)
	}
	if got := div(9, 0); got != 0 {
		t.Fatalf("div() = %v, want %v", got, 0.0)
	}
	if got := mul(3, 2.5); got != 7.5 {
		t.Fatalf("mul() = %v, want %v", got, 7.5)
	}
	if got := sub(10, 3.5); got != 6.5 {
		t.Fatalf("sub() = %v, want %v", got, 6.5)
	}
	if got := toFloat64(7); got != 7 {
		t.Fatalf("float64() = %v, want %v", got, 7.0)
	}
}

func TestLoadTemplatePlanDetails(t *testing.T) {
	resetTemplateCache()
	t.Cleanup(resetTemplateCache)

	tmpl, err := loadTemplate("plan_details.html")
	if err != nil {
		t.Fatalf("loadTemplate(plan_details.html) error = %v", err)
	}

	data := map[string]interface{}{
		"title":        "Test Plan",
		"userId":       "owner-1",
		"is_owner":     true,
		"is_member":    true,
		"user_balance": 0.0,
		"plan": domain.FamilyPlan{
			ID:             "plan-1",
			Name:           "Test Plan",
			Description:    "Plan description",
			Cost:           12,
			IndividualCost: 20,
			Owner:          "owner-1",
			JoinCode:       "ABC123",
		},
		"members": []domain.Member{
			{ID: "owner-1", Username: "owner", Name: "Owner"},
			{ID: "member-1", Username: "member", Name: "Member", Balance: -4.5},
			{ID: "artificial-1", Name: "Offline Person", IsArtificial: true},
		},
		"total_members":   3,
		"join_requests":   []domain.JoinRequest{{UserID: "request-1", Username: "joiner", Name: "Joiner", RequestedAt: "2026-04-02 00:00:00Z"}},
		"pending_request": false,
		"pending_payments": []domain.Payment{
			{ID: "payment-1", UserID: "member-1", Amount: 4.5, Date: "2026-04-02", Status: "pending", Name: "Member"},
		},
		"user_payments":      []domain.Payment{},
		"existingMembership": nil,
		"all_payments": []domain.Payment{
			{ID: "payment-2", UserID: "member-1", Amount: 4.5, Date: "2026-04-02", Status: "approved", Name: "Member"},
		},
		"member_payments_pagination": domain.MemberPaymentsPagination{
			CurrentPage: 1,
			HasPrev:     false,
			PrevPage:    1,
			HasNext:     true,
			NextPage:    2,
		},
		"total_payments":  4.5,
		"total_savings":   24.0,
		"plan_age_days":   7,
		"isAuthenticated": true,
		"username":        "owner",
		"name":            "Owner",
	}

	var out bytes.Buffer
	if err := tmpl.ExecuteTemplate(&out, "layout", data); err != nil {
		t.Fatalf("ExecuteTemplate(layout) error = %v", err)
	}

	rendered := out.String()
	for _, expected := range []string{
		"Test Plan",
		"Members",
		"Pending Payment Claims",
		"Join Requests",
		"Transfer Membership",
		"member_payments_page=2#member-payments",
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("rendered template missing %q", expected)
		}
	}
}

func TestLoadTemplateFamilyPlans(t *testing.T) {
	resetTemplateCache()
	t.Cleanup(resetTemplateCache)

	tmpl, err := loadTemplate("family_plans.html")
	if err != nil {
		t.Fatalf("loadTemplate(family_plans.html) error = %v", err)
	}

	data := map[string]interface{}{
		"title":  "My Family Plans",
		"userId": "owner-1",
		"plans": []domain.FamilyPlan{
			{
				ID:           "plan-1",
				Name:         "Test Plan",
				Description:  "Plan description",
				Cost:         12,
				Owner:        "owner-1",
				JoinCode:     "ABC123",
				MembersCount: 2,
			},
		},
		"isAuthenticated": true,
		"username":        "owner",
		"name":            "Owner",
	}

	var out bytes.Buffer
	if err := tmpl.ExecuteTemplate(&out, "layout", data); err != nil {
		t.Fatalf("ExecuteTemplate(layout) error = %v", err)
	}

	rendered := out.String()
	for _, expected := range []string{
		"My Family Plans",
		"Create New Family Plan",
		"Join Existing Plan",
		"go to url '/family-plans'",
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("rendered template missing %q", expected)
		}
	}
}

func TestLoadTemplateProfile(t *testing.T) {
	resetTemplateCache()
	t.Cleanup(resetTemplateCache)

	tmpl, err := loadTemplate("profile.html")
	if err != nil {
		t.Fatalf("loadTemplate(profile.html) error = %v", err)
	}

	data := map[string]interface{}{
		"title":           "Profile",
		"name":            "Owner",
		"isAuthenticated": true,
		"username":        "owner",
	}

	var out bytes.Buffer
	if err := tmpl.ExecuteTemplate(&out, "layout", data); err != nil {
		t.Fatalf("ExecuteTemplate(layout) error = %v", err)
	}

	rendered := out.String()
	for _, expected := range []string{
		"Edit Profile",
		"reload() the location of the window",
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("rendered template missing %q", expected)
		}
	}
}

func resetTemplateCache() {
	templateCacheMu.Lock()
	templateCache = map[string]*template.Template{}
	templateCacheMu.Unlock()
}
