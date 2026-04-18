package handler

import (
	"net/http"
	"time"
)

const (
	cookieAccessToken  = "access_token"
	cookieRefreshToken = "refresh_token"
)

const refreshCookiePath = "/api/v1/auth"

type cookieIssuer struct {
	secure bool
	domain string
}

func newCookieIssuer(secure bool, domain string) *cookieIssuer {
	return &cookieIssuer{secure: secure, domain: domain}
}

func (c *cookieIssuer) setAccess(w http.ResponseWriter, token string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieAccessToken,
		Value:    token,
		Path:     "/",
		Domain:   c.domain,
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   c.secure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (c *cookieIssuer) setRefresh(w http.ResponseWriter, token string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieRefreshToken,
		Value:    token,
		Path:     refreshCookiePath,
		Domain:   c.domain,
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   c.secure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (c *cookieIssuer) clearAccess(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieAccessToken,
		Value:    "",
		Path:     "/",
		Domain:   c.domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   c.secure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (c *cookieIssuer) clearRefresh(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieRefreshToken,
		Value:    "",
		Path:     refreshCookiePath,
		Domain:   c.domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   c.secure,
		SameSite: http.SameSiteStrictMode,
	})
}
