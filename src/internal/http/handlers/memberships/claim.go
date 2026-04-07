package memberships

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"familyplan/src/internal/http/sessionutil"
	"familyplan/src/internal/memberclaim"
	"familyplan/src/internal/planutil"
	"familyplan/src/internal/view"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
)

var errClaimLinkUnavailable = errors.New("member claim link unavailable")

// HandleCreateMemberClaimLink creates or reuses a public claim link for an artificial member.
func HandleCreateMemberClaimLink(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := sessionOrRedirect(c)
		if err != nil {
			return err
		}

		joinCode := c.PathParam("join_code")
		artificialMemberID := strings.TrimSpace(c.FormValue("artificial_member_id"))
		if artificialMemberID == "" {
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

		err = app.Dao().RunInTransaction(func(txDao *daos.Dao) error {
			membership, err := planutil.FindMembershipWithDao(txDao, planRecord.Id, artificialMemberID)
			if err != nil {
				return err
			}
			if membership == nil || !membership.GetBool("is_artificial") || !membership.GetDateTime("date_ended").IsZero() {
				return errClaimLinkUnavailable
			}

			_, err = memberclaim.EnsureWithDao(txDao, planRecord.Id, artificialMemberID)
			return err
		})
		if err != nil {
			if errors.Is(err, errClaimLinkUnavailable) {
				return c.Redirect(http.StatusSeeOther, "/"+joinCode)
			}
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/"+joinCode)
	}
}

// HandleClaimMemberPage renders the public claim page for a member-takeover link.
func HandleClaimMemberPage(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := strings.TrimSpace(c.PathParam("token"))
		errorMessage := strings.TrimSpace(c.QueryParam("error"))

		info, err := memberclaim.Lookup(app, token)
		if err != nil {
			if !errors.Is(err, memberclaim.ErrClaimLinkNotFound) && !errors.Is(err, memberclaim.ErrArtificialMemberUnavailable) {
				return err
			}

			if errorMessage == "" {
				errorMessage = memberclaim.ErrorMessage(err)
			}

			return view.RenderPage(c, "claim_member.html", map[string]interface{}{
				"title":         "Claim Link Unavailable - Family Plan Manager",
				"claim_invalid": true,
				"error":         errorMessage,
			})
		}

		canClaim := false
		session, ok := sessionutil.Current(c)
		if ok && session.IsAuthenticated {
			if errorMessage == "" {
				if planutil.IsOwner(info.PlanRecord, session.UserID) {
					errorMessage = memberclaim.ErrorMessage(memberclaim.ErrAlreadyMember)
				} else {
					existingMembership, err := planutil.FindMembership(app, info.PlanID, session.UserID)
					if err != nil {
						return err
					}
					if existingMembership != nil {
						errorMessage = memberclaim.ErrorMessage(memberclaim.ErrAlreadyMember)
					} else {
						canClaim = true
					}
				}
			}
		}

		return view.RenderPage(c, "claim_member.html", map[string]interface{}{
			"title":         "Claim Membership - Family Plan Manager",
			"claim_invalid": false,
			"claim_token":   token,
			"plan_name":     info.PlanName,
			"join_code":     info.JoinCode,
			"member_name":   info.ArtificialName,
			"can_claim":     canClaim,
			"error":         errorMessage,
		})
	}
}

// HandleClaimMember completes a public artificial-member claim for the signed-in user.
func HandleClaimMember(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := strings.TrimSpace(c.PathParam("token"))
		session, ok := sessionutil.Current(c)
		if !ok || !session.IsAuthenticated {
			return c.Redirect(http.StatusSeeOther, loginPathWithClaim(token))
		}

		result, err := memberclaim.Claim(app, token, session.UserID)
		if err != nil {
			if message := memberclaim.ErrorMessage(err); message != "" {
				return c.Redirect(http.StatusSeeOther, claimPathWithError(token, message))
			}
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/"+result.JoinCode)
	}
}

func loginPathWithClaim(token string) string {
	values := url.Values{}
	if strings.TrimSpace(token) != "" {
		values.Set("claim", token)
	}

	return pathWithQuery("/login", values)
}

func claimPathWithError(token, errorMessage string) string {
	values := url.Values{}
	if strings.TrimSpace(errorMessage) != "" {
		values.Set("error", errorMessage)
	}

	return pathWithQuery(memberclaim.Path(token), values)
}

func pathWithQuery(path string, values url.Values) string {
	if values == nil {
		return path
	}

	encoded := values.Encode()
	if encoded == "" {
		return path
	}

	return path + "?" + encoded
}
