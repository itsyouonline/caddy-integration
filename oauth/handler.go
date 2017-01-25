package oauth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mholt/caddy/caddyhttp/httpserver"
	"golang.org/x/oauth2"
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
	CallbackPath  string
	OauthConfs    map[string]*oauth2.Config
	Usernames     map[string][]string
	Organizations map[string][]string
	Next          httpserver.Handler
	hc            http.Client
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	switch {
	case httpserver.Path(r.URL.Path).Matches(h.CallbackPath):
		return h.serveCallback(w, r)
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

	url := conf.AuthCodeURL(path)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
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
		h.writeError(w, http.StatusUnauthorized, err.Error())
		return http.StatusUnauthorized, err
	}

	// save JWT token in cookies
	h.setCookies(jwtToken, expire, w)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	return http.StatusTemporaryRedirect, nil
}

// serve other dirs
func (h handler) serveHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	for p, conf := range h.OauthConfs {
		if !httpserver.Path(r.URL.Path).Matches(p) {
			continue
		}

		// get JWT token from cookies
		token := h.getJWTTokenFromCookies(r)
		if token == "" {
			return h.serveLogin(w, r)
		}

		// verify jwt token
		info, err := h.verifyJWTToken(conf, p, token)
		if err != nil {
			h.delCookies(w)
			h.writeError(w, http.StatusUnauthorized, err.Error())
			return http.StatusUnauthorized, err
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

func (h handler) writeError(w http.ResponseWriter, code int, msg string) (int, error) {
	w.WriteHeader(code)
	return code, nil
}
