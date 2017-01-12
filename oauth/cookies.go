package oauth

import (
	"net/http"
	"time"
)

const (
	cookieName = "caddyoauth"
)

func delCookies(w http.ResponseWriter) {
	setCookies("", 0, w)
}

func setCookies(code string, expire int64, w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:    cookieName,
		Value:   code,
		Path:    "/",
		Expires: time.Now().Add(time.Second * time.Duration(expire)),
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
