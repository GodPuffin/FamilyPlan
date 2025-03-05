package main

import (
	"embed"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
)

// setupRoutes configures all application routes
func setupRoutes(app *pocketbase.PocketBase, e *echo.Echo, templatesFS embed.FS) {
	// Static files
	e.Static("/", "static")

	// Middleware for all routes
	e.Use(setupAuthMiddleware(app, templatesFS))

	// Public routes
	e.GET("/", handleHome(templatesFS))
	e.GET("/login", handleLoginPage(templatesFS))
	e.POST("/login", handleLoginSubmit(app))
	e.GET("/register", handleRegisterPage(templatesFS))
	e.POST("/register", handleRegisterSubmit(app))
	e.GET("/logout", handleLogout())

	// Protected routes
	authenticated := e.Group("", requireAuth)

	// Profile routes
	authenticated.POST("/profile", handleProfileUpdate(app))

	// Family plans routes
	authenticated.GET("/family-plans", handleFamilyPlansList(app, templatesFS))
	authenticated.POST("/family-plans/create", handleCreateFamilyPlan(app))
	authenticated.POST("/family-plans/join", handleJoinPlan(app))
	authenticated.GET("/:join_code", handlePlanDetails(app, templatesFS))

	// Plan actions
	authenticated.GET("/:join_code/request-join", handleRequestJoin(app))
	authenticated.POST("/:join_code/request-join", handleRequestJoin(app))
	authenticated.POST("/:join_code/approve-request", handleApproveRequest(app))
	authenticated.POST("/:join_code/deny-request", handleDenyRequest(app))
	authenticated.POST("/:join_code/remove-member", handleRemoveMember(app))
	authenticated.POST("/:join_code/leave", handleLeavePlan(app))
	authenticated.POST("/:join_code/delete", handleDeletePlan(app))
	authenticated.POST("/:join_code/update", handleUpdatePlan(app))
}
