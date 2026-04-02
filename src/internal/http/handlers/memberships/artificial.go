package memberships

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"familyplan/src/internal/planutil"
	"familyplan/src/internal/support/random"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

const maxArtificialMemberNameLength = 80

// HandleAddArtificialMember creates an artificial member record.
func HandleAddArtificialMember(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := sessionOrRedirect(c)
		if err != nil {
			return err
		}
		joinCode := c.PathParam("join_code")
		memberName := strings.TrimSpace(c.FormValue("name"))

		if memberName == "" || utf8.RuneCountInString(memberName) > maxArtificialMemberNameLength {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		planRecord, err := planutil.FindPlanByJoinCode(app, joinCode)
		if err != nil {
			return err
		}
		if planRecord == nil {
			return c.Redirect(http.StatusSeeOther, "/family-plans")
		}

		if !planutil.IsOwner(planRecord, session.UserID) {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		membershipsCollection, err := app.Dao().FindCollectionByNameOrId("memberships")
		if err != nil {
			return err
		}

		existingArtificialFilter, err := planutil.BuildEqualsFilter(
			planutil.FilterTerm{Field: "plan_id", Value: planRecord.Id},
			planutil.FilterTerm{Field: "is_artificial", Value: true},
		)
		if err != nil {
			return err
		}

		existingArtificialMembers, err := app.Dao().FindRecordsByFilter(
			membershipsCollection.Id,
			existingArtificialFilter.Expression,
			"",
			-1,
			0,
			existingArtificialFilter.Params,
		)
		if err != nil {
			return err
		}

		for _, membership := range existingArtificialMembers {
			if !membership.GetDateTime("date_ended").IsZero() {
				continue
			}
			if strings.EqualFold(strings.TrimSpace(membership.GetString("name")), memberName) {
				return c.Redirect(http.StatusSeeOther, "/"+joinCode)
			}
		}

		artificialUserID, err := random.GenerateUUID()
		if err != nil {
			return err
		}

		newMembership := pbmodels.NewRecord(membershipsCollection)
		newMembership.Set("plan_id", planRecord.Id)
		newMembership.Set("user_id", artificialUserID)
		newMembership.Set("is_artificial", true)
		newMembership.Set("name", memberName)
		if err := app.Dao().SaveRecord(newMembership); err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/"+joinCode)
	}
}
