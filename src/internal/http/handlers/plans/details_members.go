package plans

import (
	"familyplan/src/internal/billing"
	"familyplan/src/internal/domain"
	"familyplan/src/internal/planutil"

	"github.com/pocketbase/pocketbase"
)

func loadMembers(app *pocketbase.PocketBase, plan domain.FamilyPlan) ([]domain.Member, int, error) {
	usersCollection, err := app.Dao().FindCollectionByNameOrId("users")
	if err != nil {
		return nil, 0, err
	}

	members := make([]domain.Member, 0)
	uniqueMembers := make(map[string]bool)

	ownerRecord, err := app.Dao().FindRecordById(usersCollection.Id, plan.Owner)
	if err == nil && ownerRecord != nil {
		members = append(members, domain.Member{
			ID:       ownerRecord.Id,
			Username: ownerRecord.GetString("username"),
			Name:     ownerRecord.GetString("name"),
			Balance:  0,
		})
		uniqueMembers[ownerRecord.Id] = true
	}

	membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
	if err != nil {
		return nil, 0, err
	}

	membershipFilter, err := planutil.BuildEqualsFilter(
		planutil.FilterTerm{Field: "plan_id", Value: plan.ID},
	)
	if err != nil {
		return nil, 0, err
	}

	membershipRecords, err := app.Dao().FindRecordsByFilter(
		membershipsCollection.Id,
		membershipFilter,
		"",
		-1,
		0,
	)
	if err != nil {
		return nil, 0, err
	}

	for _, membership := range membershipRecords {
		userID := membership.GetString("user_id")
		if uniqueMembers[userID] {
			continue
		}

		dateEnded := membership.GetDateTime("date_ended")
		if !dateEnded.IsZero() {
			continue
		}

		balance, _ := billing.CalculateMemberBalance(app, plan.ID, userID)

		if membership.GetBool("is_artificial") {
			members = append(members, domain.Member{
				ID:             userID,
				Name:           membership.GetString("name"),
				Balance:        balance,
				LeaveRequested: membership.GetBool("leave_requested"),
				DateEnded:      membership.GetDateTime("date_ended").String(),
				IsArtificial:   true,
			})
			uniqueMembers[userID] = true
			continue
		}

		userRecord, err := app.Dao().FindRecordById(usersCollection.Id, userID)
		if err != nil || userRecord == nil {
			continue
		}

		members = append(members, domain.Member{
			ID:             userRecord.Id,
			Username:       userRecord.GetString("username"),
			Name:           userRecord.GetString("name"),
			Balance:        balance,
			LeaveRequested: membership.GetBool("leave_requested"),
			DateEnded:      membership.GetDateTime("date_ended").String(),
			IsArtificial:   false,
		})
		uniqueMembers[userRecord.Id] = true
	}

	return members, len(members), nil
}

func loadJoinRequests(app *pocketbase.PocketBase, planID string) ([]domain.JoinRequest, error) {
	joinRequestsCollection, err := app.Dao().FindCollectionByNameOrId("join_requests")
	if err != nil {
		return nil, err
	}

	requestFilter, err := planutil.BuildEqualsFilter(
		planutil.FilterTerm{Field: "plan_id", Value: planID},
	)
	if err != nil {
		return nil, err
	}

	requestRecords, err := app.Dao().FindRecordsByFilter(
		joinRequestsCollection.Id,
		requestFilter,
		"",
		-1,
		0,
	)
	if err != nil {
		return nil, err
	}

	usersCollection, err := app.Dao().FindCollectionByNameOrId("users")
	if err != nil {
		return nil, err
	}

	joinRequests := make([]domain.JoinRequest, 0, len(requestRecords))
	for _, request := range requestRecords {
		userID := request.GetString("user_id")
		userRecord, err := app.Dao().FindRecordById(usersCollection.Id, userID)
		if err != nil || userRecord == nil {
			continue
		}

		joinRequests = append(joinRequests, domain.JoinRequest{
			UserID:      userID,
			Username:    userRecord.GetString("username"),
			Name:        userRecord.GetString("name"),
			RequestedAt: request.GetDateTime("created").String(),
		})
	}

	return joinRequests, nil
}
