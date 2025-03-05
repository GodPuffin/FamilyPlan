package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

// Handler for requesting to join a plan
func handleRequestJoin(app *pocketbase.PocketBase) echo.HandlerFunc {
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
			return c.Redirect(http.StatusSeeOther, "/family-plans")
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

		// Redirect back to plan page
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
	}
}

// Handler for approving a join request
func handleApproveRequest(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)
		joinCode := c.PathParam("join_code")
		userId := c.FormValue("user_id")

		// Find the plan by join code
		plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
		if err != nil {
			return err
		}

		plan, err := app.Dao().FindFirstRecordByData(plansCollection.Id, "join_code", joinCode)
		if err != nil || plan == nil {
			// Plan not found
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Verify user is the plan owner
		ownerSlice := plan.GetStringSlice("owner")
		if len(ownerSlice) == 0 || ownerSlice[0] != session.UserId {
			// Not authorized
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Find the join request
		joinRequestsCollection, err := app.Dao().FindCollectionByNameOrId("join_requests")
		if err != nil {
			return err
		}

		request, err := app.Dao().FindFirstRecordByFilter(joinRequestsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && user_id = '%s'", plan.Id, userId))
		if err != nil || request == nil {
			// Request not found
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Add user to memberships
		membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		// Check if membership already exists
		existingMembership, _ := app.Dao().FindFirstRecordByFilter(membershipsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && user_id = '%s'", plan.Id, userId))

		if existingMembership == nil {
			// Create new membership
			newMembership := models.NewRecord(membershipsCollection)
			newMembership.Set("plan_id", plan.Id)
			newMembership.Set("user_id", userId)

			if err := app.Dao().SaveRecord(newMembership); err != nil {
				return err
			}
		}

		// Delete the join request
		if err := app.Dao().DeleteRecord(request); err != nil {
			return err
		}

		// Redirect back to plan page
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
	}
}

// Handler for denying a join request
func handleDenyRequest(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)
		joinCode := c.PathParam("join_code")
		userId := c.FormValue("user_id")

		// Find the plan by join code
		plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
		if err != nil {
			return err
		}

		plan, err := app.Dao().FindFirstRecordByData(plansCollection.Id, "join_code", joinCode)
		if err != nil || plan == nil {
			// Plan not found
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Verify user is the plan owner
		ownerSlice := plan.GetStringSlice("owner")
		if len(ownerSlice) == 0 || ownerSlice[0] != session.UserId {
			// Not authorized
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Find the join request
		joinRequestsCollection, err := app.Dao().FindCollectionByNameOrId("join_requests")
		if err != nil {
			return err
		}

		request, err := app.Dao().FindFirstRecordByFilter(joinRequestsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && user_id = '%s'", plan.Id, userId))
		if err != nil || request == nil {
			// Request not found
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Delete the join request
		if err := app.Dao().DeleteRecord(request); err != nil {
			return err
		}

		// Redirect back to plan page
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
	}
}

// Handler for removing a member from a plan
func handleRemoveMember(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)
		joinCode := c.PathParam("join_code")
		userId := c.FormValue("user_id")

		// Find the plan by join code
		plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
		if err != nil {
			return err
		}

		plan, err := app.Dao().FindFirstRecordByData(plansCollection.Id, "join_code", joinCode)
		if err != nil || plan == nil {
			// Plan not found
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Verify user is the plan owner
		ownerSlice := plan.GetStringSlice("owner")
		if len(ownerSlice) == 0 || ownerSlice[0] != session.UserId {
			// Not authorized
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Find the membership
		membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		membership, err := app.Dao().FindFirstRecordByFilter(membershipsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && user_id = '%s'", plan.Id, userId))
		if err == nil && membership != nil {
			// Delete the membership
			if err := app.Dao().DeleteRecord(membership); err != nil {
				return err
			}
		}

		// Redirect back to plan page
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
	}
}

// Handler for leaving a plan (for members)
func handleLeavePlan(app *pocketbase.PocketBase) echo.HandlerFunc {
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
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Can't leave if you're the owner
		ownerSlice := plan.GetStringSlice("owner")
		if len(ownerSlice) > 0 && ownerSlice[0] == session.UserId {
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Find the membership
		membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		membership, err := app.Dao().FindFirstRecordByFilter(membershipsCollection.Id,
			fmt.Sprintf("plan_id = '%s' && user_id = '%s'", plan.Id, session.UserId))
		if err == nil && membership != nil {
			// Delete the membership
			if err := app.Dao().DeleteRecord(membership); err != nil {
				return err
			}
		}

		// Redirect to family plans page
		return c.Redirect(http.StatusSeeOther, "/family-plans")
	}
}

// Handler for deleting a plan (for owner)
func handleDeletePlan(app *pocketbase.PocketBase) echo.HandlerFunc {
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
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Verify user is the plan owner
		ownerSlice := plan.GetStringSlice("owner")
		if len(ownerSlice) == 0 || ownerSlice[0] != session.UserId {
			// Not authorized
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s", joinCode))
		}

		// Use a transaction to ensure all deletions succeed or fail together
		err = app.Dao().RunInTransaction(func(txDao *daos.Dao) error {
			// Delete all related memberships first
			membershipsCollection, err := txDao.FindCollectionByNameOrId("memberships")
			if err != nil {
				return err
			}

			// Use -1 for no limit to ensure we get all records
			memberships, err := txDao.FindRecordsByFilter(membershipsCollection.Id,
				fmt.Sprintf("plan_id = '%s'", plan.Id), "", -1, 0)
			if err == nil {
				for _, membership := range memberships {
					if err := txDao.DeleteRecord(membership); err != nil {
						return err
					}
				}
			}

			// Delete all related join requests
			joinRequestsCollection, err := txDao.FindCollectionByNameOrId("join_requests")
			if err != nil {
				return err
			}

			// Use -1 for no limit to ensure we get all records
			joinRequests, err := txDao.FindRecordsByFilter(joinRequestsCollection.Id,
				fmt.Sprintf("plan_id = '%s'", plan.Id), "", -1, 0)
			if err == nil {
				for _, request := range joinRequests {
					if err := txDao.DeleteRecord(request); err != nil {
						return err
					}
				}
			}

			// Finally delete the plan itself
			return txDao.DeleteRecord(plan)
		})

		if err != nil {
			// If there was an error, redirect with error message
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%s?error=Failed+to+delete+plan", joinCode))
		}

		// Redirect to family plans page
		return c.Redirect(http.StatusSeeOther, "/family-plans")
	}
}

// Handler for updating plan details
func handleUpdatePlan(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session := c.Get("session").(SessionData)
		joinCode := c.PathParam("join_code")

		// Get the family plan
		plansCollection, err := app.Dao().FindCollectionByNameOrId("family_plans")
		if err != nil || plansCollection == nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		plan, err := app.Dao().FindFirstRecordByData(plansCollection.Id, "join_code", joinCode)
		if err != nil || plan == nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Check if user is the owner
		ownerSlice := plan.GetStringSlice("owner")
		if len(ownerSlice) == 0 || ownerSlice[0] != session.UserId {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		// Get form values
		name := c.FormValue("name")
		description := c.FormValue("description")
		costStr := c.FormValue("cost")

		// Parse cost
		cost, err := strconv.ParseFloat(costStr, 64)
		if err != nil {
			// Use current cost if error
			cost = plan.GetFloat("cost")
		}

		// Update plan
		plan.Set("name", name)
		plan.Set("description", description)
		plan.Set("cost", cost)

		// Save changes
		if err := app.Dao().SaveRecord(plan); err != nil {
			return err
		}

		// Redirect back to plan details
		return c.Redirect(http.StatusSeeOther, "/"+joinCode)
	}
}
