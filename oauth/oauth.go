package oauth

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"golang.org/x/oauth2"
)

type config struct {
	Name         string
	Paths        []string
	RedirectURL  string
	CallbackPath string
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	Scopes       []string
	Usernames    map[string]struct{}
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

	oauthConfig := &oauth2.Config{
		RedirectURL:  conf.RedirectURL,
		ClientID:     conf.ClientID,
		ClientSecret: conf.ClientSecret,
		Scopes:       conf.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  conf.AuthURL,
			TokenURL: conf.TokenURL,
		},
	}

	httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return &handler{
			Name:         conf.Name,
			OauthConf:    oauthConfig,
			CallbackPath: conf.CallbackPath,
			Paths:        conf.Paths,
			Next:         next,
			hc:           http.Client{},
			Usernames:    conf.Usernames,
		}
	})
	return nil
}

func parse(c *caddy.Controller) (config, error) {
	// This parses the following config blocks
	/*
		oauth {
			path /hello
		}
	*/
	var err error
	var usernames []string

	conf := config{}

	for c.Next() {
		args := c.RemainingArgs()
		switch len(args) {
		case 0:
			// no argument passed, check the config block
			for c.NextBlock() {
				switch c.Val() {
				case "name":
					conf.Name, err = parseOne(c)
				case "path":
					p, err := parseOne(c)
					if err != nil {
						return conf, err
					}
					conf.Paths = append(conf.Paths, strings.Split(p, ",")...)
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
					str, err := parseOne(c)
					if err != nil {
						return conf, err
					}
					for _, s := range strings.Split(str, ",") {
						scope := "user:memberof:" + strings.TrimSpace(s)
						conf.Scopes = append(conf.Scopes, scope)
					}
				case "usernames":
					str, err := parseOne(c)
					if err != nil {
						return conf, err
					}
					usernames = append(usernames, strings.Split(str, ",")...)

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
	if conf.Name == "" || conf.RedirectURL == "" || conf.ClientID == "" || conf.ClientSecret == "" {
		return conf, fmt.Errorf("name, redirect_url, client_id, and client_secret can't be empty")
	}
	if conf.AuthURL == "" {
		conf.AuthURL = "https://itsyou.online/v1/oauth/authorize"
	}
	if conf.TokenURL == "" {
		conf.TokenURL = "https://itsyou.online/v1/oauth/access_token"
	}

	// usernames
	conf.Usernames = map[string]struct{}{}
	for _, u := range usernames {
		conf.Usernames[u] = struct{}{}
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

func init() {
	if os.Getenv("CADDY_DEV_MODE") == "1" {
		httpserver.RegisterDevDirective("oauth", "browse")
	}
}
