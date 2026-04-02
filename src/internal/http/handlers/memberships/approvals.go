package memberships

import (
	"errors"
	"net/http"

	"familyplan/src/internal/planutil"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
	pbmodels "github.com/pocketbase/pocketbase/models"
)

// HandleApproveRequest approves a pending join request.
func HandleApproveRequest(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := sessionOrRedirect(c)
		if err != nil {
			return err
		}
		joinCode := c.PathParam("join_code")
		userID := c.FormValue("user_id")

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

		requestNotFound := errors.New("join request not found")
		err = app.Dao().RunInTransaction(func(txDao *daos.Dao) error {
			request, err := planutil.FindJoinRequestWithDao(txDao, planRecord.Id, userID)
			if err != nil {
				return err
			}
			if request == nil {
				return requestNotFound
			}

			membershipsCollection, err := txDao.FindCollectionByNameOrId("memberships")
			if err != nil {
				return err
			}

			existingMembership, err := planutil.FindMembershipWithDao(txDao, planRecord.Id, userID)
			if err != nil {
				return err
			}
			if existingMembership == nil {
				newMembership := pbmodels.NewRecord(membershipsCollection)
				newMembership.Set("plan_id", planRecord.Id)
				newMembership.Set("user_id", userID)
				newMembership.Set("is_artificial", false)
				if err := txDao.SaveRecord(newMembership); err != nil {
					return err
				}
			}

			return txDao.DeleteRecord(request)
		})
		if err != nil {
			if errors.Is(err, requestNotFound) {
				return c.Redirect(http.StatusSeeOther, "/"+joinCode)
			}
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/"+joinCode)
	}
}

// HandleDenyRequest denies a pending join request.
func HandleDenyRequest(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := sessionOrRedirect(c)
		if err != nil {
			return err
		}
		joinCode := c.PathParam("join_code")
		userID := c.FormValue("user_id")

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

		request, err := planutil.FindJoinRequest(app, planRecord.Id, userID)
		if err != nil {
			return err
		}
		if request == nil {
			return c.Redirect(http.StatusSeeOther, "/"+joinCode)
		}

		if err := app.Dao().DeleteRecord(request); err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/"+joinCode)
	}
}
