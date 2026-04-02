package planutil

import (
	"github.com/pocketbase/pocketbase"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

const (
	collectionFamilyPlans  = "family_plans"
	collectionMemberships  = "memberships"
	collectionJoinRequests = "join_requests"
)

// FindPlanByJoinCode returns the plan record matching the join code.
func FindPlanByJoinCode(app *pocketbase.PocketBase, joinCode string) (*pbmodels.Record, error) {
	plansCollection, err := app.Dao().FindCollectionByNameOrId(collectionFamilyPlans)
	if err != nil {
		return nil, err
	}

	return app.Dao().FindFirstRecordByData(plansCollection.Id, "join_code", joinCode)
}

// OwnerID returns the plan owner's user ID.
func OwnerID(plan *pbmodels.Record) string {
	ownerIDs := plan.GetStringSlice("owner")
	if len(ownerIDs) == 0 {
		return ""
	}

	return ownerIDs[0]
}

// IsOwner reports whether the user owns the plan.
func IsOwner(plan *pbmodels.Record, userID string) bool {
	return OwnerID(plan) == userID
}

// FindMembership returns the membership record for a plan/user pair.
func FindMembership(app *pocketbase.PocketBase, planID, userID string) (*pbmodels.Record, error) {
	membershipsCollection, err := app.Dao().FindCollectionByNameOrId(collectionMemberships)
	if err != nil {
		return nil, err
	}

	filter, err := BuildEqualsFilter(
		FilterTerm{Field: "plan_id", Value: planID},
		FilterTerm{Field: "user_id", Value: userID},
	)
	if err != nil {
		return nil, err
	}

	return app.Dao().FindFirstRecordByFilter(
		membershipsCollection.Id,
		filter,
	)
}

// FindJoinRequest returns the join request record for a plan/user pair.
func FindJoinRequest(app *pocketbase.PocketBase, planID, userID string) (*pbmodels.Record, error) {
	joinRequestsCollection, err := app.Dao().FindCollectionByNameOrId(collectionJoinRequests)
	if err != nil {
		return nil, err
	}

	filter, err := BuildEqualsFilter(
		FilterTerm{Field: "plan_id", Value: planID},
		FilterTerm{Field: "user_id", Value: userID},
	)
	if err != nil {
		return nil, err
	}

	return app.Dao().FindFirstRecordByFilter(
		joinRequestsCollection.Id,
		filter,
	)
}
