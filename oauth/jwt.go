package oauth

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/oauth2"
)

var jwtPubKey *ecdsa.PublicKey

const (
	iyoPubKey = `-----BEGIN PUBLIC KEY-----
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAES5X8XrfKdx9gYayFITc89wad4usrk0n2
7MjiGYvqalizeSWTHEpnd7oea9IQ8T5oJjMVH5cc0H5tFSKilFFeh//wngxIyny6
6+Vq5t5B0V0Ehy01+2ceEon2Y0XDkIKv
-----END PUBLIC KEY-----`
)

type jwtInfo struct {
	Username string
	Scopes   []string
}

func init() {
	var err error

	jwtPubKey, err = jwt.ParseECPublicKeyFromPEM([]byte(iyoPubKey))
	if err != nil {
		fmt.Printf("failed to parse pub key:%v\n", err)
		os.Exit(1)
	}
}

func (h handler) verifyJWTToken(conf *oauth2.Config, protectedPath, tokenStr string) (*jwtInfo, error) {
	// verify token
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodES384 {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return jwtPubKey, nil
	})
	if err != nil {
		return nil, err
	}

	// get claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !(ok && token.Valid) {
		return nil, fmt.Errorf("invalid token")
	}

	username, okUsername := h.checkUsername(protectedPath, claims)
	okScope := h.checkScope(conf.Scopes, claims)
	if !okUsername ||  !okScope{
		return nil, fmt.Errorf("not allowed to access this resource")
	}

	return &jwtInfo{
		Username: username,
	}, nil
}

func (h handler) checkUsername(protectedPath string, claims map[string]interface{}) (string, bool) {
	username, ok := claims["username"].(string)
	if !ok {
		if username, ok = claims["globalid"].(string); !ok {
			return username, false
		}
	}

	usernames, exists := h.Usernames[protectedPath]
	if !exists {
		return username, true
	}
	return username, inArray(username, usernames)
}

func (h handler) checkScope(scopes []string, claims map[string]interface{}) bool {
	if len(scopes) == 0 {
		return true
	}

	for _, v := range claims["scope"].([]interface{}) {
		scope, ok := v.(string)
		if !ok {
			continue
		}
		if inArray(scope, scopes) {
			return true
		}
	}
	return false
}

// check if string `str` exist in array `arr`
func inArray(str string, arr []string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}

// get JWT token from oauth2 authorization code
func (h handler) getJWTToken(conf *oauth2.Config, code, state string) (int64, string, error) {
	// get oauth2 token
	token, err := h.getToken(conf, code, state)
	if err != nil {
		return 0, "", err
	}

	// get JWT token with scope of each organization
	for _, scope := range conf.Scopes {
		jwtToken, err := h.getJWTTokenScope(token.AccessToken, scope)
		if err == nil && jwtToken != "" {
			return token.ExpiresIn, jwtToken, nil
		}
	}
	jwtToken, err := h.getJWTTokenScope(token.AccessToken, "")
	return token.ExpiresIn, jwtToken, err
}

func (h handler) getJWTTokenScope(accessToken, scope string) (string, error) {
	// build request
	req, err := http.NewRequest("GET", "https://itsyou.online/v1/oauth/jwt", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "token "+accessToken)

	if len(scope) > 0 {
		q := req.URL.Query()
		q.Add("scope", scope)
		req.URL.RawQuery = q.Encode()
	}

	// do request
	resp, err := h.hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("code=%v", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	return string(body), err
}
