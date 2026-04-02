package planutil

import (
	"database/sql"
	"errors"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

const (
	collectionFamilyPlans  = "family_plans"
	collectionMemberships  = "memberships"
	collectionJoinRequests = "join_requests"
)

// FindPlanByJoinCode returns the plan record matching the join code.
func FindPlanByJoinCode(app *pocketbase.PocketBase, joinCode string) (*pbmodels.Record, error) {
	return FindPlanByJoinCodeWithDao(app.Dao(), joinCode)
}

// FindPlanByJoinCodeWithDao returns the plan record matching the join code using the provided dao.
func FindPlanByJoinCodeWithDao(dao *daos.Dao, joinCode string) (*pbmodels.Record, error) {
	plansCollection, err := dao.FindCollectionByNameOrId(collectionFamilyPlans)
	if err != nil {
		return nil, err
	}

	record, err := dao.FindFirstRecordByData(plansCollection.Id, "join_code", joinCode)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return record, err
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
	return FindMembershipWithDao(app.Dao(), planID, userID)
}

// FindMembershipWithDao returns the membership record for a plan/user pair using the provided dao.
func FindMembershipWithDao(dao *daos.Dao, planID, userID string) (*pbmodels.Record, error) {
	membershipsCollection, err := dao.FindCollectionByNameOrId(collectionMemberships)
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

	record, err := dao.FindFirstRecordByFilter(
		membershipsCollection.Id,
		filter.Expression,
		filter.Params,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return record, err
}

// FindJoinRequest returns the join request record for a plan/user pair.
func FindJoinRequest(app *pocketbase.PocketBase, planID, userID string) (*pbmodels.Record, error) {
	return FindJoinRequestWithDao(app.Dao(), planID, userID)
}

// FindJoinRequestWithDao returns the join request record for a plan/user pair using the provided dao.
func FindJoinRequestWithDao(dao *daos.Dao, planID, userID string) (*pbmodels.Record, error) {
	joinRequestsCollection, err := dao.FindCollectionByNameOrId(collectionJoinRequests)
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

	record, err := dao.FindFirstRecordByFilter(
		joinRequestsCollection.Id,
		filter.Expression,
		filter.Params,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return record, err
}
