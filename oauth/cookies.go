package oauth

import (
	"net/http"
)

const (
	cookieName = "caddyoauth"
)

func setOauthCookies(code string, w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:  cookieName,
		Value: code,
		Path:  "/",
	}
	http.SetCookie(w, &cookie)
}

func getOauthCookies(r *http.Request) string {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}
	return c.Value
}
