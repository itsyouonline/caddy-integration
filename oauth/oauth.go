package oauth

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/itsyouonline/caddy"
	"github.com/itsyouonline/caddy/caddyhttp/httpserver"
	"golang.org/x/oauth2"
)

type config struct {
	RedirectURL            string
	CallbackPath           string
	ClientID               string
	ClientSecret           string
	AuthURL                string
	TokenURL               string
	Organizations          map[string][]string
	Usernames              map[string][]string
	AuthenticationRequired []string
}

func newConfig() config {
	return config{
		Organizations:          map[string][]string{},
		Usernames:              map[string][]string{},
		AuthenticationRequired: []string{},
	}
}

func init() {
	caddy.RegisterPlugin("oauth", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	conf, err := parse(c)
	if err != nil {
		return err
	}

	// Runs on Caddy startup, useful for services or other setups.
	c.OnStartup(func() error {
		fmt.Printf("caddy_oauth plugin is initiated with conf=%#v\n", conf)
		return nil
	})

	// Runs on Caddy shutdown, useful for cleanups.
	c.OnShutdown(func() error {
		fmt.Println("caddy_oauth plugin is cleaning up")
		return nil
	})

	oauthConfs := map[string]*oauth2.Config{}

	// create oauth conf for organizations
	for path, orgs := range conf.Organizations {
		var scopes []string
		for _, org := range orgs {
			scopes = append(scopes, "user:memberof:"+org)
		}
		oauthConfs[path] = newOauthConf(conf, scopes)
	}

	// create oauth conf for usernames
	for path := range conf.Usernames {
		if _, ok := oauthConfs[path]; ok {
			continue
		}
		oauthConfs[path] = newOauthConf(conf, []string{})
	}

	for _, path := range conf.AuthenticationRequired {
		if _, ok := oauthConfs[path]; ok {
			continue
		}
		oauthConfs[path] = newOauthConf(conf, []string{})
	}

	httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return &handler{
			CallbackPath:           conf.CallbackPath,
			Next:                   next,
			hc:                     http.Client{},
			OauthConfs:             oauthConfs,
			Usernames:              conf.Usernames,
			Organizations:          conf.Organizations,
			AuthenticationRequired: conf.AuthenticationRequired,
		}
	})
	return nil
}

func newOauthConf(conf config, scopes []string) *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  conf.RedirectURL,
		ClientID:     conf.ClientID,
		ClientSecret: conf.ClientSecret,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  conf.AuthURL,
			TokenURL: conf.TokenURL,
		},
	}
}

func parse(c *caddy.Controller) (config, error) {
	// This parses the following config blocks
	var err error
	conf := newConfig()

	for c.Next() {
		args := c.RemainingArgs()
		switch len(args) {
		case 0:
			// no argument passed, check the config block
			for c.NextBlock() {
				switch c.Val() {
				case "redirect_url":
					conf.RedirectURL, err = parseOne(c)
				case "client_id":
					conf.ClientID, err = parseOne(c)
				case "client_secret":
					conf.ClientSecret, err = parseOne(c)
				case "auth_url":
					conf.AuthURL, err = parseOne(c)
				case "token_url":
					conf.TokenURL, err = parseOne(c)
				case "organizations":
					path, orgs, e := parseTwo(c)
					if e != nil {
						return conf, e
					}
					conf.Organizations[path] = strings.Split(orgs, ",")
				case "usernames":
					path, usernames, e := parseTwo(c)
					if e != nil {
						return conf, e
					}
					conf.Usernames[path] = strings.Split(usernames, ",")
				case "authentication_required":
					path, e := parseOne(c)
					if e != nil {
						return conf, e
					}
					conf.AuthenticationRequired = append(conf.AuthenticationRequired, path)
				}
				if err != nil {
					return conf, err
				}
			}
		default:
			// we want only one argument max
			return conf, c.ArgErr()
		}
	}
	if conf.RedirectURL == "" || conf.ClientID == "" || conf.ClientSecret == "" {
		return conf, fmt.Errorf("redirect_url, client_id, and client_secret can't be empty")
	}
	if conf.AuthURL == "" {
		conf.AuthURL = "https://itsyou.online/v1/oauth/authorize"
	}
	if conf.TokenURL == "" {
		conf.TokenURL = "https://itsyou.online/v1/oauth/access_token"
	}

	// callback path
	redirURL, err := url.Parse(conf.RedirectURL)
	if err != nil {
		return conf, err
	}
	conf.CallbackPath = redirURL.Path

	return conf, nil
}

// parse exactly one arguments
func parseOne(c *caddy.Controller) (string, error) {
	if !c.NextArg() {
		// we are expecting a value
		return "", c.ArgErr()
	}
	val := c.Val()
	if c.NextArg() {
		// we are expecting only one value.
		return "", c.ArgErr()
	}
	return val, nil
}

func parseTwo(c *caddy.Controller) (string, string, error) {
	args := c.RemainingArgs()
	if len(args) != 2 {
		return "", "", fmt.Errorf("expected 2 args, get %v args", len(args))
	}
	return args[0], args[1], nil
}

func init() {
	if os.Getenv("CADDY_DEV_MODE") == "1" {
		httpserver.RegisterDevDirective("oauth", "browse")
	}
}
