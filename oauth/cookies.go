package oauth

import (
	"net/http"
)

const (
	cookieName = "caddyoauth"
)

func delCookies(w http.ResponseWriter) {
	setCookies("", w)
}

func setCookies(code string, w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:  cookieName,
		Value: code,
		Path:  "/",
	}
	http.SetCookie(w, &cookie)
}

func getCookies(r *http.Request) string {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}
	return c.Value
}
