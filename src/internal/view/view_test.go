package view

import (
	"bytes"
	"strings"
	"testing"

	"familyplan/src/internal/domain"
)

func TestLoadTemplatePlanDetails(t *testing.T) {
	t.Parallel()

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
		"all_payments":       []domain.Payment{{ID: "payment-2", UserID: "member-1", Amount: 4.5, Date: "2026-04-02", Status: "approved", Name: "Member"}},
		"total_payments":     4.5,
		"total_savings":      24.0,
		"plan_age_days":      7,
		"isAuthenticated":    true,
		"username":           "owner",
		"name":               "Owner",
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
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("rendered template missing %q", expected)
		}
	}
}

func TestLoadTemplateFamilyPlans(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
