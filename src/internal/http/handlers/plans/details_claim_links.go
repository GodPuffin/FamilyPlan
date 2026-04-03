package plans

import (
	"fmt"

	"familyplan/src/internal/memberclaim"

	"github.com/pocketbase/pocketbase"
)

func loadMemberClaimLinks(app *pocketbase.PocketBase, planID, scheme, host string) (map[string]string, error) {
	records, err := memberclaim.FindAllForPlan(app, planID)
	if err != nil {
		return nil, err
	}

	links := make(map[string]string, len(records))
	for _, record := range records {
		token := record.GetString("token")
		artificialMemberID := record.GetString("artificial_member_id")
		if token == "" || artificialMemberID == "" {
			continue
		}

		links[artificialMemberID] = absoluteClaimLink(scheme, host, token)
	}

	return links, nil
}

func absoluteClaimLink(scheme, host, token string) string {
	if scheme == "" {
		scheme = "http"
	}
	if host == "" {
		return memberclaim.Path(token)
	}

	return fmt.Sprintf("%s://%s%s", scheme, host, memberclaim.Path(token))
}
