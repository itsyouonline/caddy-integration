package oauth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"bytes"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"golang.org/x/oauth2"
	"time"
)

type token struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	ExpiresIn   int64  `json:"expires_in"`
	Info        struct {
		Username string `json:"username"`
	} `json:"info"`
}

type handler struct {
	ExtraScopes            string
	LoginPage              string
	LoginURL               string
	LogoutURL              string
	JwtURL                 string
	CallbackPath           string
	APIPath                string
	OauthConfs             map[string]*oauth2.Config
	Usernames              map[string][]string
	Organizations          map[string][]string
	AuthenticationRequired []string
	AllowedExtensions      []string
	ForwardPayload         bool
	Next                   httpserver.Handler
	hc                     http.Client
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	switch {
	case httpserver.Path(r.URL.Path).Matches(h.CallbackPath):
		return h.serveCallback(w, r)
	case h.LogoutURL != "" && httpserver.Path(r.URL.Path).Matches(h.LogoutURL):
		return h.serveLogout(w, r)
	default:
		return h.serveHTTP(w, r)
	}
}

func (h handler) getPathConf(r *http.Request) (string, *oauth2.Config) {
	for path, conf := range h.OauthConfs {
		if httpserver.Path(r.URL.Path).Matches(path) {
			return path, conf
		}
	}
	return "", nil
}

// server oauth2 login page
func (h handler) serveLogin(w http.ResponseWriter, r *http.Request) (int, error) {
	path, conf := h.getPathConf(r)
	if conf == nil {
		return 500, fmt.Errorf("null oauth conf when serving login page for path `%v`", r.URL.Path)
	}

	// request all scopes permissions while login
	var scopes bytes.Buffer
	for _, conf := range h.OauthConfs {
		if len(conf.Scopes) > 0 {
			scopes.WriteString(strings.Join(conf.Scopes, ","))
			scopes.WriteString(",")
		}
	}
	scopes.WriteString(h.ExtraScopes)
	var url string
	if scopes.String() != "" {
		scopeOption := oauth2.SetAuthURLParam("scope", scopes.String())
		url = conf.AuthCodeURL(path, scopeOption)
	} else {
		url = conf.AuthCodeURL(path)
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	return http.StatusTemporaryRedirect, nil
}

// server oauth2 logout page
func (h handler) serveLogout(w http.ResponseWriter, r *http.Request) (int, error) {
	h.delCookies(w)
	r.Header.Del("Authorization")
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	return http.StatusTemporaryRedirect, nil
}

// server oauth2 callback page
func (h handler) serveCallback(w http.ResponseWriter, r *http.Request) (int, error) {

	// get authorization code
	code := r.FormValue("code")
	state := r.FormValue("state")
	if code == "" || state == "" {
		return http.StatusUnauthorized, nil
	}

	conf, ok := h.OauthConfs[state]
	if !ok {
		return http.StatusUnauthorized, fmt.Errorf("oauth2 config not found")
	}

	// get JWT token from IYO server
	expire, jwtToken, err := h.getJWTToken(conf, code, state)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized)
		return http.StatusUnauthorized, err
	}

	// save JWT token in cookies
	h.setCookies(jwtToken, expire, w)

	//Redirect back to the origin path saved in the cookies
	var origin string
	c, err := r.Cookie("origin")
	if err != nil {
		origin = "/"
	} else {
		origin = c.Value
	}
	cookie := http.Cookie{
		Name:    "origin",
		Value:   "",
		Path:    "/",
		Expires: time.Now(),
	}
	http.SetCookie(w, &cookie)

	http.Redirect(w, r, origin, http.StatusTemporaryRedirect)
	return http.StatusTemporaryRedirect, nil
}

// serve other dirs
func (h handler) serveHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	if h.LoginURL != "" && httpserver.Path(r.URL.Path).Matches(h.LoginURL) {
		if _, ok := r.URL.Query()["redirect_back"]; ok {
			origin := r.Referer()
			cookie := &http.Cookie{
				Name:  "origin",
				Value: origin,
				Path:  "/",
			}
			http.SetCookie(w, cookie)
		}
		return h.serveLogin(w, r)
	}
	//Check if a valid jwt is present in the `Authorization` header
	authorizationHeader := r.Header.Get("Authorization")
	token := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(authorizationHeader), "bearer"), "Bearer"))
	if token == "" {
		//Check if a jwt is present in a cookie (normally an interactive session through a browser)
		token = h.getJWTTokenFromCookies(r)
	}

	// if the user isn't logged in and he requested the login page, serve it normally
	// else if the user is already logged in and he requested the login page, redirect to the root path "/"
	if token == "" && h.LoginPage != "" && httpserver.Path(r.URL.Path).Matches(h.LoginPage) {
		return h.Next.ServeHTTP(w, r)
	} else if h.LoginPage != "" && httpserver.Path(r.URL.Path).Matches(h.LoginPage) {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return http.StatusTemporaryRedirect, nil
	}

	// serve allowed extensions
	for _, extension := range h.AllowedExtensions {
		if strings.HasSuffix(r.URL.Path, extension) {
			return h.Next.ServeHTTP(w, r)
		}
	}
	for p, conf := range h.OauthConfs {
		if !httpserver.Path(r.URL.Path).Matches(p) {
			continue
		}

		if token == "" {
			// Return Unauthorized if the url relates to any API endpoint and don't continue with login flow
			if httpserver.Path(r.URL.Path).Matches(h.APIPath) {
				return http.StatusUnauthorized, nil
			}

			//Save the origin path into cookies to redirect back to it after login
			cookie := &http.Cookie{
				Name:  "origin",
				Value: r.URL.Path,
				Path:  "/",
			}
			http.SetCookie(w, cookie)
			if h.LoginPage != "" {
				http.Redirect(w, r, h.LoginPage, http.StatusTemporaryRedirect)
				return http.StatusTemporaryRedirect, nil
			}
			return h.serveLogin(w, r)
		}

		// verify jwt token
		info, err := h.verifyJWTToken(conf, p, token)
		if err != nil {
			// Raise forbidden if the user is correctly logged in but has no access to the resource
			if err.Error() == string(http.StatusForbidden){
				return http.StatusForbidden, err
			}
			// Delete cookies and raise unauthorized if the user has invalid JWT
			h.delCookies(w)
			return http.StatusUnauthorized, err
		}

		r.Header.Set("X-Iyo-Username", info.Username)

		if h.ForwardPayload {
			json, err := json.Marshal(info.Payload)
			if err == nil {
				r.Header.Set("X-Iyo-Token", string(json))
			}
		}

		logRequest(w, r, info)
	}
	return h.Next.ServeHTTP(w, r)
}

func (h handler) getToken(conf *oauth2.Config, code, state string) (*token, error) {
	// build request
	req, err := http.NewRequest("POST", conf.Endpoint.TokenURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("client_id", conf.ClientID)
	q.Add("client_secret", conf.ClientSecret)
	q.Add("code", code)
	q.Add("redirect_uri", conf.RedirectURL)
	q.Add("state", state)
	req.URL.RawQuery = q.Encode()

	// do request
	resp, err := h.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// decode response
	var t token

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body:%v", err)
	}

	err = json.Unmarshal(body, &t)
	return &t, err
}

func (h handler) getJWTTokenFromCookies(r *http.Request) string {
	return h.getCookies(r)
}

func (h handler) writeError(w http.ResponseWriter, code int) (int, error) {
	w.WriteHeader(code)
	return code, nil
}
