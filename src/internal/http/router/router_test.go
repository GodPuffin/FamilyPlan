package router

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
)

func TestSetupRegistersExpectedRoutes(t *testing.T) {
	t.Parallel()

	e := echo.New()
	Setup(&pocketbase.PocketBase{}, e)

	expected := map[string]string{
		http.MethodGet + " /":                            "/",
		http.MethodGet + " /login":                       "/login",
		http.MethodPost + " /login":                      "/login",
		http.MethodGet + " /register":                    "/register",
		http.MethodPost + " /register":                   "/register",
		http.MethodGet + " /logout":                      "/logout",
		http.MethodGet + " /profile":                     "/profile",
		http.MethodPost + " /profile":                    "/profile",
		http.MethodGet + " /family-plans":                "/family-plans",
		http.MethodPost + " /family-plans/create":        "/family-plans/create",
		http.MethodPost + " /family-plans/join":          "/family-plans/join",
		http.MethodGet + " /:join_code":                  "/:join_code",
		http.MethodPost + " /:join_code/delete":          "/:join_code/delete",
		http.MethodPost + " /:join_code/update":          "/:join_code/update",
		http.MethodPost + " /:join_code/approve-request": "/:join_code/approve-request",
		http.MethodPost + " /:join_code/deny-request":    "/:join_code/deny-request",
		http.MethodPost + " /:join_code/remove-member":   "/:join_code/remove-member",
		http.MethodPost + " /:join_code/leave":           "/:join_code/leave",
		http.MethodPost + " /:join_code/claim-payment":   "/:join_code/claim-payment",
		http.MethodPost + " /:join_code/add-payment":     "/:join_code/add-payment",
	}

	registered := map[string]string{}
	for _, route := range e.Router().Routes() {
		method := route.Method()
		path := route.Path()
		registered[method+" "+path] = path
	}

	for key, wantPath := range expected {
		if gotPath, ok := registered[key]; !ok || gotPath != wantPath {
			t.Fatalf("missing expected route %q", key)
		}
	}
}
