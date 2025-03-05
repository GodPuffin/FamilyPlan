package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
)

// Handler for listing family plans
func handleFamilyPlansList(app *pocketbase.PocketBase, templatesFS embed.FS) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)

		// Get family plans collection
		plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
		if err != nil {
			return err
		}

		// Get all plans where user is the owner
		ownedPlansRecords, err := app.Dao().FindRecordsByFilter(plansCollection.Id,
			fmt.Sprintf("owner ~ '%s'", session.UserId), "", 100, 0)
		if err != nil {
			return err
		}

		// Get memberships collection
		membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		// Get user's memberships
		memberships, err := app.Dao().FindRecordsByFilter(membershipsCollection.Id,
			fmt.Sprintf("user_id = '%s'", session.UserId), "", 100, 0)
		if err != nil {
			return err
		}

		// Get plans for each membership
		planMap := make(map[string]*models.Record)
		for _, record := range ownedPlansRecords {
			planMap[record.Id] = record
		}

		for _, membership := range memberships {
			planId := membership.GetString("plan_id")

			if _, exists := planMap[planId]; !exists {
				plan, err := app.Dao().FindRecordById(plansCollection.Id, planId)
				if err != nil {
					continue
				}
				planMap[planId] = plan
			}
		}

		// Convert to slice for template
		uniquePlanMap := make(map[string]*models.Record)
		for id, plan := range planMap {
			uniquePlanMap[id] = plan
		}

		plans := []FamilyPlan{}
		for _, record := range uniquePlanMap {
			// Count members
			membersCount, err := app.Dao().FindRecordsByFilter(membershipsCollection.Id,
				fmt.Sprintf("plan_id = '%s'", record.Id), "", 100, 0)
			if err != nil {
				continue
			}

			// Add plan to list
			plan := FamilyPlan{
				Id:           record.Id,
				Name:         record.GetString("name"),
				Description:  record.GetString("description"),
				Cost:         record.GetFloat("cost"),
				Owner:        record.GetStringSlice("owner")[0],
				JoinCode:     record.GetString("join_code"),
				CreatedAt:    record.GetDateTime("created").String(),
				MembersCount: len(membersCount),
			}

			plans = append(plans, plan)
		}

		// Parse and render the template
		funcMap := template.FuncMap{
			"upper": strings.ToUpper,
			"slice": func(s string, i, j int) string {
				if i >= len(s) {
					return ""
				}
				if j > len(s) {
					j = len(s)
				}
				return s[i:j]
			},
		}

		tmpl, err := template.New("layout").Funcs(funcMap).ParseFS(templatesFS, "templates/layout.html", "templates/family_plans.html")
		if err != nil {
			return err
		}

		return tmpl.ExecuteTemplate(c.Response().Writer, "layout", map[string]interface{}{
			"title":           "My Family Plans",
			"isAuthenticated": session.IsAuthenticated,
			"username":        session.Username,
			"name":            session.Name,
			"userId":          session.UserId,
			"plans":           plans,
		})
	}
}

// Handler for creating a new family plan
func handleCreateFamilyPlan(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)

		// Get form data
		name := c.FormValue("name")
		description := c.FormValue("description")
		costStr := c.FormValue("cost")

		// Validate required fields
		if name == "" || costStr == "" {
			// Handle validation error
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Parse cost to float
		cost, err := strconv.ParseFloat(costStr, 64)
		if err != nil {
			// Handle parsing error
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Generate a unique 6-character join code
		joinCode := generateJoinCode(6)

		// Get family plans collection
		plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
		if err != nil {
			return err
		}

		// Create new plan record
		newPlan := models.NewRecord(plansCollection)
		newPlan.Set("name", name)
		newPlan.Set("description", description)
		newPlan.Set("cost", cost)
		newPlan.Set("owner", []string{session.UserId})
		newPlan.Set("join_code", joinCode)

		// Save the new plan
		if err := app.Dao().SaveRecord(newPlan); err != nil {
			return err
		}

		// Get memberships collection
		membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		// Create new membership record for the owner
		newMembership := models.NewRecord(membershipsCollection)
		newMembership.Set("plan_id", newPlan.Id)
		newMembership.Set("user_id", session.UserId)

		// Save the membership
		_ = app.Dao().SaveRecord(newMembership)

		// Redirect to family plans page
		return c.Redirect(http.StatusSeeOther, "/family-plans")
	}
}

// Handler for joining an existing plan
func handleJoinPlan(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)

		// Get form data
		joinCode := c.FormValue("join_code")

		// Validate join code
		if joinCode == "" {
			// Handle validation error
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Find the plan by join code
		plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
		if err != nil {
			return err
		}

		plan, err := app.Dao().FindFirstRecordByData(plansCollection.Id, "join_code", joinCode)
		if err != nil || plan == nil {
			// Plan not found, redirect back
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Check if user is already a member or owner
		ownerSlice := plan.GetStringSlice("owner")
		if len(ownerSlice) > 0 && ownerSlice[0] == session.UserId {
			// User is already the owner
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Check if user is already a member
		membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		existingMembership, _ := app.Dao().FindFirstRecordByFilter(membershipsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && user_id = '%s'", plan.Id, session.UserId))

		if existingMembership != nil {
			// User is already a member
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Create a join request
		joinRequestsCollection, err := app.Dao().FindCollectionByNameOrId("join_requests")
		if err != nil {
			return err
		}

		// Check if a request already exists
		existingRequest, _ := app.Dao().FindFirstRecordByFilter(joinRequestsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && user_id = '%s'", plan.Id, session.UserId))

		if existingRequest == nil {
			// Create new request
			newRequest := models.NewRecord(joinRequestsCollection)
			newRequest.Set("plan_id", plan.Id)
			newRequest.Set("user_id", session.UserId)

			if err := app.Dao().SaveRecord(newRequest); err != nil {
				return err
			}
		}

		// Redirect to plan details page
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
	}
}

// Handler for viewing plan details
func handlePlanDetails(app *pocketbase.PocketBase, templatesFS embed.FS) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)
		joinCode := c.PathParam("join_code")

		// Find the plan by join code
		plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
		if err != nil {
			return err
		}

		plan, err := app.Dao().FindFirstRecordByData(plansCollection.Id, "join_code", joinCode)
		if err != nil || plan == nil {
			// Plan not found
			funcMap := template.FuncMap{
				"upper": strings.ToUpper,
				"slice": func(s string, i, j int) string {
					if i >= len(s) {
						return ""
					}
					if j > len(s) {
						j = len(s)
					}
					return s[i:j]
				},
			}

			tmpl, err := template.New("layout").Funcs(funcMap).ParseFS(templatesFS, "templates/layout.html", "templates/plan_details.html")
			if err != nil {
				return err
			}

			return tmpl.ExecuteTemplate(c.Response().Writer, "layout", map[string]interface{}{
				"title":           "Plan Not Found",
				"isAuthenticated": session.IsAuthenticated,
				"username":        session.Username,
				"name":            session.Name,
				"userId":          session.UserId,
				"not_found":       true,
				"plan":            nil,
			})
		}

		// Check if user is the owner
		isOwner := false
		ownerSlice := plan.GetStringSlice("owner")
		if len(ownerSlice) > 0 && ownerSlice[0] == session.UserId {
			isOwner = true
		}

		// Check if user is a member
		membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		existingMembership, _ := app.Dao().FindFirstRecordByFilter(membershipsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && user_id = '%s'", plan.Id, session.UserId))

		isMember := existingMembership != nil || isOwner

		// If not a member, check if there's a pending request
		pendingRequest := false
		if !isMember {
			joinRequestsCollection, err := app.Dao().FindCollectionByNameOrId("join_requests")
			if err != nil {
				return err
			}

			existingRequest, _ := app.Dao().FindFirstRecordByFilter(joinRequestsCollection.Id,
				fmt.Sprintf("plan_id = '%s' && user_id = '%s'", plan.Id, session.UserId))

			pendingRequest = existingRequest != nil
		}

		// Convert plan record to struct
		familyPlan := FamilyPlan{
			Id:          plan.Id,
			Name:        plan.GetString("name"),
			Description: plan.GetString("description"),
			Cost:        plan.GetFloat("cost"),
			Owner:       ownerSlice[0],
			JoinCode:    plan.GetString("join_code"),
			CreatedAt:   plan.GetDateTime("created").String(),
		}

		// If user is a member or owner, get all members
		members := []Member{}
		var totalMembers int

		if isMember {
			// Get owner info
			usersCollection, err := app.Dao().FindCollectionByNameOrId("users")
			if err != nil {
				return err
			}

			// Create a map to track unique members by ID
			uniqueMembers := make(map[string]bool)

			// Add owner to members list
			ownerRecord, err := app.Dao().FindRecordById(usersCollection.Id, familyPlan.Owner)
			if err == nil && ownerRecord != nil {
				members = append(members, Member{
					Id:       ownerRecord.Id,
					Username: ownerRecord.GetString("username"),
					Name:     ownerRecord.GetString("name"),
				})
				uniqueMembers[ownerRecord.Id] = true
			}

			// Get all memberships for this plan
			memberships, err := app.Dao().FindRecordsByFilter(membershipsCollection.Id,
				fmt.Sprintf("plan_id = '%s'", plan.Id), "", 100, 0)
			if err != nil {
				return err
			}

			// Get user details for each member
			for _, membership := range memberships {
				userId := membership.GetString("user_id")

				// Skip if we already added this user (avoid duplicates)
				if uniqueMembers[userId] {
					continue
				}

				userRecord, err := app.Dao().FindRecordById(usersCollection.Id, userId)
				if err == nil && userRecord != nil {
					members = append(members, Member{
						Id:       userRecord.Id,
						Username: userRecord.GetString("username"),
						Name:     userRecord.GetString("name"),
					})
					uniqueMembers[userRecord.Id] = true
				}
			}

			totalMembers = len(members)
		}

		// If user is the owner, get join requests
		joinRequests := []JoinRequest{}

		if isOwner {
			joinRequestsCollection, err := app.Dao().FindCollectionByNameOrId("join_requests")
			if err != nil {
				return err
			}

			requestRecords, err := app.Dao().FindRecordsByFilter(joinRequestsCollection.Id,
				fmt.Sprintf("plan_id = '%s'", plan.Id), "", 100, 0)
			if err != nil {
				return err
			}

			// Get user details for each request
			usersCollection, err := app.Dao().FindCollectionByNameOrId("users")
			if err != nil {
				return err
			}

			for _, request := range requestRecords {
				userId := request.GetString("user_id")
				userRecord, err := app.Dao().FindRecordById(usersCollection.Id, userId)
				if err == nil && userRecord != nil {
					joinRequests = append(joinRequests, JoinRequest{
						UserId:      userId,
						Username:    userRecord.GetString("username"),
						Name:        userRecord.GetString("name"),
						RequestedAt: request.GetDateTime("created").String(),
					})
				}
			}
		}

		// Parse and render the template
		funcMap := template.FuncMap{
			"upper": strings.ToUpper,
			"slice": func(s string, i, j int) string {
				if i >= len(s) {
					return ""
				}
				if j > len(s) {
					j = len(s)
				}
				return s[i:j]
			},
		}

		// Create a data map for the template
		data := map[string]interface{}{
			"title":           familyPlan.Name,
			"isAuthenticated": session.IsAuthenticated,
			"username":        session.Username,
			"name":            session.Name,
			"userId":          session.UserId,
			"plan":            familyPlan,
			"is_owner":        isOwner,
			"is_member":       isMember,
			"members":         members,
			"total_members":   totalMembers,
			"join_requests":   joinRequests,
			"pending_request": pendingRequest,
		}

		tmpl, err := template.New("layout").Funcs(funcMap).ParseFS(templatesFS, "templates/layout.html", "templates/plan_details.html")
		if err != nil {
			return err
		}

		return tmpl.ExecuteTemplate(c.Response().Writer, "layout", data)
	}
}
