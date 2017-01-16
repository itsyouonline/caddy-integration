package oauth

import (
	"net/http"
	"time"
)

func (h handler) cookieName() string {
	return "caddyoauth"
}

func (h handler) delCookies(w http.ResponseWriter) {
	h.setCookies("", 0, w)
}

func (h handler) setCookies(code string, expire int64, w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:    h.cookieName(),
		Value:   code,
		Path:    "/",
		Expires: time.Now().Add(time.Second * time.Duration(expire)),
	}
	http.SetCookie(w, &cookie)
}

func (h handler) getCookies(r *http.Request) string {
	c, err := r.Cookie(h.cookieName())
	if err != nil {
		return ""
	}
	return c.Value
}
