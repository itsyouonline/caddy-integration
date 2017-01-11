package oauth

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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
		log.Fatalf("failed to parse pub key:%v", err)
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
	var scopes []string
	for _, v := range claims["scope"].([]interface{}) {
		scopes = append(scopes, v.(string))
	}

	inScopes := func(scope string) bool {
		for _, s := range scopes {
			if s == scope {
				return true
			}
		}
		return false
	}

	if len(h.OauthConf.Scopes) > 0 {
		for _, s := range h.OauthConf.Scopes {
			if !inScopes(s) {
				return nil, fmt.Errorf("user doesn't have `%v` scope", s)
			}
		}
	}

	return &jwtInfo{
		Username: claims["username"].(string),
		Scopes:   scopes,
	}, nil
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

	body, err := ioutil.ReadAll(resp.Body)
	return token.ExpiresIn, string(body), err
}
