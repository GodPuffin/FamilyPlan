package auth

import (
	"net/http"
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
		Secure:   true,
	}
}
