package billing

import (
	"time"

	"familyplan/src/internal/planutil"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

// GetActiveMembershipsForMonth returns memberships that were active for the target month.
func GetActiveMembershipsForMonth(app *pocketbase.PocketBase, planID string, targetMonth time.Time) ([]*pbmodels.Record, error) {
	return getActiveMembershipsForMonth(app.Dao(), planID, targetMonth)
}

func getActiveMembershipsForMonth(dao *daos.Dao, planID string, targetMonth time.Time) ([]*pbmodels.Record, error) {
	membershipsCollection, err := dao.FindCollectionByNameOrId("memberships")
	if err != nil {
		return nil, err
	}

	filter, err := planutil.BuildEqualsFilter(
		planutil.FilterTerm{Field: "plan_id", Value: planID},
	)
	if err != nil {
		return nil, err
	}

	allMemberships, err := dao.FindRecordsByFilter(
		membershipsCollection.Id,
		filter.Expression,
		"",
		-1,
		0,
		filter.Params,
	)
	if err != nil {
		return nil, err
	}

	activeMemberships := make([]*pbmodels.Record, 0, len(allMemberships))
	monthStart := time.Date(targetMonth.Year(), targetMonth.Month(), 1, 0, 0, 0, 0, targetMonth.Location())

	for _, membership := range allMemberships {
		createdAt := membership.GetDateTime("created").Time()
		membershipMonth := time.Date(createdAt.Year(), createdAt.Month(), 1, 0, 0, 0, 0, createdAt.Location())
		if membershipMonth.After(monthStart) {
			continue
		}

		dateEndedField := membership.GetDateTime("date_ended")
		if !dateEndedField.IsZero() {
			dateEndedTime := dateEndedField.Time()
			endedMonth := time.Date(dateEndedTime.Year(), dateEndedTime.Month(), 1, 0, 0, 0, 0, dateEndedTime.Location())
			if endedMonth.Before(monthStart) {
				continue
			}
		}

		activeMemberships = append(activeMemberships, membership)
	}

	return activeMemberships, nil
}
