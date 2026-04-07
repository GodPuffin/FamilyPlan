package memberships

import (
	"errors"
	"net/http"

	"familyplan/src/internal/memberclaim"
	"familyplan/src/internal/planutil"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
)

// HandleTransferMembership converts an artificial member into a real user membership.
func HandleTransferMembership(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := sessionOrRedirect(c)
		if err != nil {
			return err
		}
		joinCode := c.PathParam("join_code")
		realUserID := c.FormValue("user_id")
		artificialMemberID := c.FormValue("artificial_member_id")

		if realUserID == "" || artificialMemberID == "" {
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

		missingTransferPrerequisite := errors.New("membership transfer prerequisite not found")
		err = app.Dao().RunInTransaction(func(txDao *daos.Dao) error {
			joinRequestsCollection, err := txDao.FindCollectionByNameOrId("join_requests")
			if err != nil {
				return err
			}

			requestFilter, err := planutil.BuildEqualsFilter(
				planutil.FilterTerm{Field: "plan_id", Value: planRecord.Id},
				planutil.FilterTerm{Field: "user_id", Value: realUserID},
			)
			if err != nil {
				return err
			}

			request, err := txDao.FindFirstRecordByFilter(
				joinRequestsCollection.Id,
				requestFilter.Expression,
				requestFilter.Params,
			)
			if err != nil || request == nil {
				return missingTransferPrerequisite
			}

			usersCollection, err := txDao.FindCollectionByNameOrId("users")
			if err != nil {
				return err
			}

			realUserRecord, err := txDao.FindRecordById(usersCollection.Id, realUserID)
			if err != nil || realUserRecord == nil {
				return missingTransferPrerequisite
			}

			if err := memberclaim.TransferArtificialMembership(txDao, planRecord, artificialMemberID, realUserID); err != nil {
				if errors.Is(err, memberclaim.ErrArtificialMemberUnavailable) || errors.Is(err, memberclaim.ErrAlreadyMember) {
					return missingTransferPrerequisite
				}
				return err
			}

			return nil
		})
		if err != nil {
			if errors.Is(err, missingTransferPrerequisite) {
				return c.Redirect(http.StatusSeeOther, "/"+joinCode)
			}
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/"+joinCode)
	}
}
