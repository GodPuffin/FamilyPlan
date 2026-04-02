package auth

import (
	"net/http"
	"os"
	"strconv"
	"time"
)

func newAuthCookie(token string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secureCookiesEnabled(),
	}
}

func secureCookiesEnabled() bool {
	value := os.Getenv("FAMILYPLAN_COOKIE_SECURE")
	if value == "" {
		return true
	}

	secure, err := strconv.ParseBool(value)
	if err != nil {
		return true
	}

	return secure
}
