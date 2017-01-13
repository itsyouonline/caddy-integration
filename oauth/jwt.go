package oauth

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
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

func (h handler) verifyJWTToken(tokenStr string) (*jwtInfo, error) {
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

	// check scopes
	ok = func() bool {
		// if no scopes specified, ignore it
		if len(h.OauthConf.Scopes) == 0 {
			return true
		}

		for _, v := range claims["scope"].([]interface{}) {
			scope := v.(string)
			if inArray(scope, h.OauthConf.Scopes) {
				return true
			}
		}
		return false
	}()
	if !ok {
		return nil, fmt.Errorf("user doesn't have one of  `%v` scope", h.OauthConf.Scopes)
	}

	// check usernames
	username := claims["username"].(string)
	ok = func() bool {
		if len(h.Usernames) == 0 {
			return true
		}
		_, exists := h.Usernames[username]
		return exists
	}()
	if !ok {
		return nil, fmt.Errorf("username `%v` not allowed to access this resource", username)
	}

	return &jwtInfo{
		Username: username,
	}, nil
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

func (h handler) getJWTToken(code string) (int64, string, error) {
	// get oauth2 token
	token, err := h.getToken(code)
	if err != nil {
		return 0, "", err
	}

	// build request
	req, err := http.NewRequest("GET", "https://itsyou.online/v1/oauth/jwt", nil)
	if err != nil {
		return 0, "", err
	}

	req.Header.Set("Authorization", "token "+token.AccessToken)

	if len(h.OauthConf.Scopes) > 0 {
		q := req.URL.Query()
		q.Add("scope", strings.Join(h.OauthConf.Scopes, ","))
		req.URL.RawQuery = q.Encode()
	}

	// do request
	resp, err := h.hc.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, "", fmt.Errorf("code=%v", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	return token.ExpiresIn, string(body), err
}
