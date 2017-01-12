package oauth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mholt/caddy/caddyhttp/httpserver"
	"golang.org/x/oauth2"
)

var (
	oauthState = "random"
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
	LoginPath    string
	CallbackPath string
	OauthConf    *oauth2.Config
	Usernames    map[string]struct{} // allowed usernames, empty to allow all
	Paths        []string
	Next         httpserver.Handler
	hc           http.Client
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	switch {
	case httpserver.Path(r.URL.Path).Matches(h.CallbackPath):
		return h.serveCallback(w, r)
	default:
		return h.serveHTTP(w, r)
	}
}

// server oauth2 login page
func (h handler) serveLogin(w http.ResponseWriter, r *http.Request) (int, error) {
	url := h.OauthConf.AuthCodeURL(oauthState)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	return http.StatusTemporaryRedirect, nil
}

// server oauth2 callback page
func (h handler) serveCallback(w http.ResponseWriter, r *http.Request) (int, error) {
	// get authorization code
	code := r.FormValue("code")
	if code == "" {
		return 401, nil
	}

	// get JWT token from IYO server
	expire, jwtToken, err := h.getJWTToken(code)
	if err != nil {
		h.writeError(w, 500, err.Error())
		return 500, err
	}

	// save JWT token in cookies
	setCookies(jwtToken, expire, w)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	return http.StatusTemporaryRedirect, nil
}

// serve other dirs
func (h handler) serveHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	for _, p := range h.Paths {
		if !httpserver.Path(r.URL.Path).Matches(p) {
			continue
		}

		// get JWT token from cookies
		token := h.getJWTTokenFromCookies(r)
		if token == "" {
			return h.serveLogin(w, r)
		}

		// verify jwt token
		info, err := h.verifyJWTToken(token)
		if err != nil {
			delCookies(w)
			h.writeError(w, 500, err.Error())
			return 500, err
		}

		logRequest(w, r, info)
	}
	return h.Next.ServeHTTP(w, r)
}

func (h handler) getToken(code string) (*token, error) {
	// build request
	req, err := http.NewRequest("POST", h.OauthConf.Endpoint.TokenURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("client_id", h.OauthConf.ClientID)
	q.Add("client_secret", h.OauthConf.ClientSecret)
	q.Add("code", code)
	q.Add("redirect_uri", h.OauthConf.RedirectURL)
	q.Add("state", oauthState)
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
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall:%v,body:%v", err, string(body))
	}
	return &t, nil
}

func (h handler) getJWTTokenFromCookies(r *http.Request) string {
	return getCookies(r)
}

func (h handler) writeError(w http.ResponseWriter, code int, msg string) (int, error) {
	w.WriteHeader(code)
	w.Write([]byte(msg))
	return code, nil
}
