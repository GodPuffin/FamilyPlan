package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

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
		membershipMap := make(map[string]*models.Record) // Map to store user's membership in each plan

		for _, record := range ownedPlansRecords {
			planMap[record.Id] = record
		}

		for _, membership := range memberships {
			planId := membership.GetString("plan_id")
			membershipMap[planId] = membership

			if _, exists := planMap[planId]; !exists {
				plan, err := app.Dao().FindRecordById(plansCollection.Id, planId)
				if err != nil {
					continue
				}
				planMap[plan.Id] = plan
			}
		}

		// Convert to a list of FamilyPlan structs
		plans := []FamilyPlan{}
		for _, planRecord := range planMap {
			ownerSlice := planRecord.GetStringSlice("owner")
			if len(ownerSlice) == 0 {
				continue
			}

			// Count members
			memberCount, err := app.Dao().FindRecordsByFilter(membershipsCollection.Id,
				fmt.Sprintf("plan_id = '%s'", planRecord.Id), "", 100, 0)
			membersCount := 0
			if err == nil {
				// Only count active memberships that haven't left (owner is already counted separately)
				for _, member := range memberCount {
					dateEnded := member.GetDateTime("date_ended")
					if !member.GetBool("leave_requested") && dateEnded.IsZero() {
						membersCount++
					}
				}
			}

			// Calculate balance for this plan if user is a member (not owner)
			var balance float64 = 0.0
			isOwner := ownerSlice[0] == session.UserId

			if !isOwner && membershipMap[planRecord.Id] != nil {
				// Only calculate balance for plans where user is a member (not owner)
				balanceAmount, err := calculateMemberBalance(app, planRecord.Id, session.UserId)
				if err == nil {
					balance = balanceAmount
				}
			}

			plans = append(plans, FamilyPlan{
				Id:             planRecord.Id,
				Name:           planRecord.GetString("name"),
				Description:    planRecord.GetString("description"),
				Cost:           planRecord.GetFloat("cost"),
				IndividualCost: planRecord.GetFloat("individual_cost"),
				Owner:          ownerSlice[0],
				JoinCode:       planRecord.GetString("join_code"),
				CreatedAt:      planRecord.GetDateTime("created").String(),
				MembersCount:   membersCount,
				Balance:        balance,
			})
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
			"formatMoney": func(amount float64) string {
				return fmt.Sprintf("$%.2f", amount)
			},
			"div": func(a, b float64) float64 {
				if b == 0 {
					return 0
				}
				return a / b
			},
			"mul": func(a, b float64) float64 {
				return a * b
			},
			"sub": func(a, b float64) float64 {
				return a - b
			},
			"float64": func(i int) float64 {
				return float64(i)
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
		individualCostStr := c.FormValue("individual_cost")

		// Validate required fields
		if name == "" || costStr == "" || individualCostStr == "" {
			// Handle validation error
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Parse cost to float
		cost, err := strconv.ParseFloat(costStr, 64)
		if err != nil {
			// Handle parsing error
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Parse individual cost to float
		individualCost, err := strconv.ParseFloat(individualCostStr, 64)
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
		newPlan.Set("individual_cost", individualCost)
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
		newMembership.Set("is_artificial", false)

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
				"formatMoney": func(amount float64) string {
					return fmt.Sprintf("$%.2f", amount)
				},
				"div": func(a, b float64) float64 {
					if b == 0 {
						return 0
					}
					return a / b
				},
				"mul": func(a, b float64) float64 {
					return a * b
				},
				"sub": func(a, b float64) float64 {
					return a - b
				},
				"float64": func(i int) float64 {
					return float64(i)
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
			Id:             plan.Id,
			Name:           plan.GetString("name"),
			Description:    plan.GetString("description"),
			Cost:           plan.GetFloat("cost"),
			IndividualCost: plan.GetFloat("individual_cost"),
			Owner:          ownerSlice[0],
			JoinCode:       plan.GetString("join_code"),
			CreatedAt:      plan.GetDateTime("created").String(),
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
				// Owner should not have a balance displayed, so we don't calculate it

				members = append(members, Member{
					Id:       ownerRecord.Id,
					Username: ownerRecord.GetString("username"),
					Name:     ownerRecord.GetString("name"),
					Balance:  0.0,
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

				// Skip members who have left the plan
				dateEnded := membership.GetDateTime("date_ended")
				if !dateEnded.IsZero() {
					continue
				}

				// Check if this is an artificial member
				isArtificial := membership.GetBool("is_artificial")
				memberName := membership.GetString("name")

				if isArtificial {
					// For artificial members, use the name stored in the membership
					balance, _ := calculateMemberBalance(app, plan.Id, userId)

					members = append(members, Member{
						Id:             userId,
						Username:       "",
						Name:           memberName,
						Balance:        balance,
						LeaveRequested: membership.GetBool("leave_requested"),
						DateEnded:      membership.GetDateTime("date_ended").String(),
						IsArtificial:   true,
					})
					uniqueMembers[userId] = true
				} else {
					// For real members, get the user record
					userRecord, err := app.Dao().FindRecordById(usersCollection.Id, userId)
					if err == nil && userRecord != nil {
						// Calculate member's balance
						balance := 0.0
						// Always calculate balance for all members regardless of who is viewing
						balance, _ = calculateMemberBalance(app, plan.Id, userId)

						members = append(members, Member{
							Id:             userRecord.Id,
							Username:       userRecord.GetString("username"),
							Name:           userRecord.GetString("name"),
							Balance:        balance,
							LeaveRequested: membership.GetBool("leave_requested"),
							DateEnded:      membership.GetDateTime("date_ended").String(),
							IsArtificial:   false,
						})
						uniqueMembers[userRecord.Id] = true
					}
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

		// Get payment data
		pendingPayments := []Payment{}
		userPayments := []Payment{}
		allPayments := []Payment{}

		if isMember {
			paymentsCollection, err := app.Dao().FindCollectionByNameOrId("payments")
			if err == nil {
				// If owner, get all pending payments and all approved/rejected payments
				if isOwner {
					pendingPaymentRecords, err := app.Dao().FindRecordsByFilter(paymentsCollection.Id,
						fmt.Sprintf("plan_id = '%s' && status = 'pending'", plan.Id), "-created", 100, 0)
					if err == nil {
						usersCollection, err := app.Dao().FindCollectionByNameOrId("users")
						if err == nil {
							// Get user details for each payment
							for _, paymentRecord := range pendingPaymentRecords {
								userId := paymentRecord.GetString("user_id")

								// Check if this is a payment from an artificial member
								isArtificialPayment := false
								artificialUserName := ""

								// Try to find a membership record to check if this is an artificial member
								artificialMembership, _ := app.Dao().FindFirstRecordByFilter(membershipsCollection.Id,
									fmt.Sprintf("plan_id = '%s' && user_id = '%s' && is_artificial = true", plan.Id, userId))

								if artificialMembership != nil {
									isArtificialPayment = true
									artificialUserName = artificialMembership.GetString("name")
								}

								if isArtificialPayment {
									// For artificial members, use the name from the membership
									forMonth := ""
									forMonthDate := paymentRecord.GetDateTime("for_month")
									if !forMonthDate.IsZero() {
										forMonth = forMonthDate.String()[:7] // Get YYYY-MM part
									}

									pendingPayments = append(pendingPayments, Payment{
										Id:       paymentRecord.Id,
										PlanId:   paymentRecord.GetString("plan_id"),
										UserId:   userId,
										Amount:   paymentRecord.GetFloat("amount"),
										Date:     paymentRecord.GetDateTime("date").String()[:10], // Get YYYY-MM-DD part
										Status:   paymentRecord.GetString("status"),
										Notes:    paymentRecord.GetString("notes"),
										ForMonth: forMonth,
										Username: "",
										Name:     artificialUserName,
									})
								} else {
									// For real members, get the user record
									userRecord, err := app.Dao().FindRecordById(usersCollection.Id, userId)
									if err == nil && userRecord != nil {
										// Add to pending payments
										forMonth := ""
										forMonthDate := paymentRecord.GetDateTime("for_month")
										if !forMonthDate.IsZero() {
											forMonth = forMonthDate.String()[:7] // Get YYYY-MM part
										}

										pendingPayments = append(pendingPayments, Payment{
											Id:       paymentRecord.Id,
											PlanId:   paymentRecord.GetString("plan_id"),
											UserId:   userId,
											Amount:   paymentRecord.GetFloat("amount"),
											Date:     paymentRecord.GetDateTime("date").String()[:10], // Get YYYY-MM-DD part
											Status:   paymentRecord.GetString("status"),
											Notes:    paymentRecord.GetString("notes"),
											ForMonth: forMonth,
											Username: userRecord.GetString("username"),
											Name:     userRecord.GetString("name"),
										})
									}
								}
							}
						}
					}

					// Get all payments for all members (for owner view)
					allPaymentRecords, err := app.Dao().FindRecordsByFilter(paymentsCollection.Id,
						fmt.Sprintf("plan_id = '%s'", plan.Id), "-created", 100, 0)
					if err == nil {
						usersCollection, err := app.Dao().FindCollectionByNameOrId("users")
						if err == nil {
							for _, paymentRecord := range allPaymentRecords {
								userId := paymentRecord.GetString("user_id")

								// Check if this is a payment from an artificial member
								isArtificialPayment := false
								artificialUserName := ""

								// Try to find a membership record to check if this is an artificial member
								artificialMembership, _ := app.Dao().FindFirstRecordByFilter(membershipsCollection.Id,
									fmt.Sprintf("plan_id = '%s' && user_id = '%s' && is_artificial = true", plan.Id, userId))

								if artificialMembership != nil {
									isArtificialPayment = true
									artificialUserName = artificialMembership.GetString("name")
								}

								if isArtificialPayment {
									// For artificial members, use the name from the membership
									forMonth := ""
									forMonthDate := paymentRecord.GetDateTime("for_month")
									if !forMonthDate.IsZero() {
										forMonth = forMonthDate.String()[:7] // Get YYYY-MM part
									}

									allPayments = append(allPayments, Payment{
										Id:       paymentRecord.Id,
										PlanId:   paymentRecord.GetString("plan_id"),
										UserId:   userId,
										Amount:   paymentRecord.GetFloat("amount"),
										Date:     paymentRecord.GetDateTime("date").String()[:10], // Get YYYY-MM-DD part
										Status:   paymentRecord.GetString("status"),
										Notes:    paymentRecord.GetString("notes"),
										ForMonth: forMonth,
										Username: "",
										Name:     artificialUserName,
									})
								} else {
									// For real members, look up the user record
									userRecord, err := app.Dao().FindRecordById(usersCollection.Id, userId)
									if err == nil && userRecord != nil {
										forMonth := ""
										forMonthDate := paymentRecord.GetDateTime("for_month")
										if !forMonthDate.IsZero() {
											forMonth = forMonthDate.String()[:7] // Get YYYY-MM part
										}

										allPayments = append(allPayments, Payment{
											Id:       paymentRecord.Id,
											PlanId:   paymentRecord.GetString("plan_id"),
											UserId:   userId,
											Amount:   paymentRecord.GetFloat("amount"),
											Date:     paymentRecord.GetDateTime("date").String()[:10], // Get YYYY-MM-DD part
											Status:   paymentRecord.GetString("status"),
											Notes:    paymentRecord.GetString("notes"),
											ForMonth: forMonth,
											Username: userRecord.GetString("username"),
											Name:     userRecord.GetString("name"),
										})
									}
								}
							}
						}
					}
				}

				// Get user's own payment history
				userPaymentRecords, err := app.Dao().FindRecordsByFilter(paymentsCollection.Id,
					fmt.Sprintf("plan_id = '%s' && user_id = '%s'", plan.Id, session.UserId), "-created", 20, 0)
				if err == nil {
					for _, paymentRecord := range userPaymentRecords {
						forMonth := ""
						forMonthDate := paymentRecord.GetDateTime("for_month")
						if !forMonthDate.IsZero() {
							forMonth = forMonthDate.String()[:7]
						}

						userPayments = append(userPayments, Payment{
							Id:       paymentRecord.Id,
							PlanId:   paymentRecord.GetString("plan_id"),
							UserId:   paymentRecord.GetString("user_id"),
							Amount:   paymentRecord.GetFloat("amount"),
							Date:     paymentRecord.GetDateTime("date").String()[:10],
							Status:   paymentRecord.GetString("status"),
							Notes:    paymentRecord.GetString("notes"),
							ForMonth: forMonth,
						})
					}
				}
			}
		}

		// Calculate total payments made for this plan
		totalPayments := 0.0
		paymentsCollection, err := app.Dao().FindCollectionByNameOrId("payments")
		if err != nil {
			return fmt.Errorf("error getting payments collection: %w", err)
		}

		approvedPayments, err := app.Dao().FindRecordsByFilter(paymentsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && status = 'approved'", plan.Id), "", -1, 0)
		if err != nil {
			return fmt.Errorf("error getting approved payments: %w", err)
		}

		for _, payment := range approvedPayments {
			amount := payment.GetFloat("amount")
			totalPayments += amount
		}

		// Calculate total savings (individual cost * members - family plan cost)
		totalSavings := 0.0
		individualCost := plan.GetFloat("individual_cost")
		familyPlanCost := plan.GetFloat("cost")

		// Get all months the plan has been active
		planCreated := plan.GetDateTime("created")
		planCreationTime := planCreated.Time()
		currentTime := time.Now()

		// Start from the first day of the month the plan was created
		startDate := time.Date(planCreationTime.Year(), planCreationTime.Month(), 1, 0, 0, 0, 0, planCreationTime.Location())

		// Go month by month until current month
		for currentDate := startDate; currentDate.Before(currentTime); currentDate = currentDate.AddDate(0, 1, 0) {
			// Get active memberships for this month
			activeMemberships, err := getActiveMembershipsForMonth(app, plan.Id, currentDate)
			if err != nil {
				// Continue with next month even if there's an error
				continue
			}

			// Calculate savings for this month
			memberCount := len(activeMemberships)
			if memberCount > 0 {
				monthlySavings := (individualCost * float64(memberCount)) - familyPlanCost
				if monthlySavings > 0 {
					totalSavings += monthlySavings
				}
			}
		}

		// Calculate plan age in days
		planAgeDays := int(time.Since(planCreationTime).Hours() / 24)

		// Get user's current balance
		userBalance := 0.0
		if isMember && !isOwner {
			userBalance, _ = calculateMemberBalance(app, plan.Id, session.UserId)
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
			"formatMoney": func(amount float64) string {
				return fmt.Sprintf("$%.2f", amount)
			},
			"div": func(a, b float64) float64 {
				if b == 0 {
					return 0
				}
				return a / b
			},
			"mul": func(a, b float64) float64 {
				return a * b
			},
			"sub": func(a, b float64) float64 {
				return a - b
			},
			"float64": func(i int) float64 {
				return float64(i)
			},
		}

		// Create a data map for the template
		data := map[string]interface{}{
			"title":              familyPlan.Name,
			"isAuthenticated":    session.IsAuthenticated,
			"username":           session.Username,
			"name":               session.Name,
			"userId":             session.UserId,
			"plan":               familyPlan,
			"is_owner":           isOwner,
			"is_member":          isMember,
			"members":            members,
			"total_members":      totalMembers,
			"join_requests":      joinRequests,
			"pending_request":    pendingRequest,
			"pending_payments":   pendingPayments,
			"user_payments":      userPayments,
			"user_balance":       userBalance,
			"existingMembership": existingMembership,
			"all_payments":       allPayments,
			"total_savings":      totalSavings,
			"plan_age_days":      planAgeDays,
		}

		tmpl, err := template.New("layout").Funcs(funcMap).ParseFS(templatesFS, "templates/layout.html", "templates/plan_details.html")
		if err != nil {
			return err
		}

		return tmpl.ExecuteTemplate(c.Response().Writer, "layout", data)
	}
}
