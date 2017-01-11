package oauth

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"golang.org/x/oauth2"
)

type config struct {
	Paths        []string
	RedirectURL  string
	LoginPath    string
	CallbackPath string
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	Scopes       []string
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
			OauthConf:    oauthConfig,
			LoginPath:    conf.LoginPath,
			CallbackPath: conf.CallbackPath,
			Paths:        conf.Paths,
			Next:         next,
			hc:           http.Client{},
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
	conf := config{}
	for c.Next() {
		args := c.RemainingArgs()
		switch len(args) {
		case 0:
			// no argument passed, check the config block
			for c.NextBlock() {
				switch c.Val() {
				case "path":
					p, err := parseOne(c)
					if err != nil {
						return conf, err
					}
					conf.Paths = append(conf.Paths, p)
				case "redirect_url":
					conf.RedirectURL, err = parseOne(c)
				case "callback_path":
					conf.CallbackPath, err = parseOne(c)
				case "login_path":
					conf.LoginPath, err = parseOne(c)
				case "client_id":
					conf.ClientID, err = parseOne(c)
				case "client_secret":
					conf.ClientSecret, err = parseOne(c)
				case "auth_url":
					conf.AuthURL, err = parseOne(c)
				case "token_url":
					conf.TokenURL, err = parseOne(c)
				case "scopes":
					var str string
					str, err = parseOne(c)
					if err != nil {
						return conf, err
					}
					conf.Scopes = append(conf.Scopes, strings.Split(str, ",")...)
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
