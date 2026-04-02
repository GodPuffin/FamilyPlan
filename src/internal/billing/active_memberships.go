package billing

import (
	"fmt"
	"time"

	"github.com/pocketbase/pocketbase"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

// GetActiveMembershipsForMonth returns memberships that were active for the target month.
func GetActiveMembershipsForMonth(app *pocketbase.PocketBase, planID string, targetMonth time.Time) ([]*pbmodels.Record, error) {
	membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
	if err != nil {
		return nil, err
	}

	allMemberships, err := app.Dao().FindRecordsByFilter(
		membershipsCollection.Id,
		fmt.Sprintf("plan_id = '%s'", planID),
		"",
		100,
		0,
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

		dateEndedTime := membership.GetDateTime("date_ended").Time()
		hasEnded := !membership.GetDateTime("date_ended").IsZero()

		if hasEnded {
			endedMonth := time.Date(dateEndedTime.Year(), dateEndedTime.Month(), 1, 0, 0, 0, 0, dateEndedTime.Location())
			if endedMonth.Before(monthStart) {
				continue
			}
		}

		activeMemberships = append(activeMemberships, membership)
	}

	return activeMemberships, nil
}
