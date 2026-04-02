package router

import (
	"net/http"
	"time"

	authhandlers "familyplan/src/internal/http/handlers/auth"
	"familyplan/src/internal/http/handlers/memberships"
	"familyplan/src/internal/http/handlers/payments"
	"familyplan/src/internal/http/handlers/plans"
	profilehandlers "familyplan/src/internal/http/handlers/profile"
	authmw "familyplan/src/internal/http/middleware"

	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
	"github.com/pocketbase/pocketbase"
)

// Setup configures the application routes.
func Setup(app *pocketbase.PocketBase, e *echo.Echo) {
	e.Use(authmw.SetupAuth(app))

	authLimiter := echomw.RateLimiterWithConfig(echomw.RateLimiterConfig{
		Store: echomw.NewRateLimiterMemoryStoreWithConfig(echomw.RateLimiterMemoryStoreConfig{
			Rate:      0.25,
			Burst:     5,
			ExpiresIn: 10 * time.Minute,
		}),
		DenyHandler: func(c echo.Context, identifier string, err error) error {
			path := "/login"
			if c.Path() == "/register" {
				path = "/register"
			}

			return c.Redirect(http.StatusSeeOther, path+"?error=Too+many+attempts.+Please+wait+and+try+again.")
		},
	})

	e.GET("/", authhandlers.HandleHome())
	e.GET("/login", authhandlers.HandleLoginPage())
	e.POST("/login", authhandlers.HandleLoginSubmit(app), authLimiter)
	e.GET("/register", authhandlers.HandleRegisterPage())
	e.POST("/register", authhandlers.HandleRegisterSubmit(app), authLimiter)
	e.GET("/logout", authhandlers.HandleLogout())

	authenticated := e.Group("", authmw.RequireAuth)

	authenticated.GET("/profile", profilehandlers.HandleProfilePage(app))
	authenticated.POST("/profile", profilehandlers.HandleProfileUpdate(app))

	authenticated.GET("/family-plans", plans.HandleFamilyPlansList(app))
	authenticated.POST("/family-plans/create", plans.HandleCreateFamilyPlan(app))
	authenticated.POST("/family-plans/join", plans.HandleJoinPlan(app))
	authenticated.GET("/:join_code", plans.HandlePlanDetails(app))
	authenticated.POST("/:join_code/delete", plans.HandleDeletePlan(app))
	authenticated.POST("/:join_code/update", plans.HandleUpdatePlan(app))

	authenticated.GET("/:join_code/request-join", memberships.HandleRequestJoin(app))
	authenticated.POST("/:join_code/request-join", memberships.HandleRequestJoin(app))
	authenticated.POST("/:join_code/approve-request", memberships.HandleApproveRequest(app))
	authenticated.POST("/:join_code/deny-request", memberships.HandleDenyRequest(app))
	authenticated.POST("/:join_code/remove-member", memberships.HandleRemoveMember(app))
	authenticated.POST("/:join_code/leave", memberships.HandleLeavePlan(app))
	authenticated.POST("/:join_code/add-artificial-member", memberships.HandleAddArtificialMember(app))
	authenticated.POST("/:join_code/transfer-membership", memberships.HandleTransferMembership(app))

	authenticated.POST("/:join_code/claim-payment", payments.HandleClaimPayment(app))
	authenticated.POST("/:join_code/approve-payment", payments.HandleApprovePayment(app))
	authenticated.POST("/:join_code/reject-payment", payments.HandleRejectPayment(app))
	authenticated.POST("/:join_code/add-payment", payments.HandleAddManualPayment(app))
}
