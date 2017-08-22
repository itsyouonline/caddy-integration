# caddy oauth plugin

This plugin protects resource paths using itsyou.online oauth2.

## why

- [why](why.md)

## Features

Plugin features:

- protects paths based on organization membership
- protects paths based on username
- use JWT to make it stateless and reduce API calls to Oauth2 server
- log following infos to stdout : host, time, http verb, path, http method, username
- sets a `X-Iyo-Username` header with the username of the logged in user


## Usage
Add the following oauth block to your Caddyfile or use the [Caddyfile.example](https://github.com/itsyouonline/caddy-integration/blob/master/oauth/Caddyfile.example)
```
oauth {
    # itsyou.online client ID
    client_id       mylab

    # itsyou.online client secret   
    client_secret   fHfT3yBlZXlNRAbOSVw-PLZI2y9HgqcA0IVzXXXXXXXXXXXXXXX

    # oauth auth url
    # leave it blank for default value
    # default value : https://itsyou.online/v1/oauth/authorize
    auth_url        https://itsyou.online/v1/oauth/authorize

    # oauth2 access token URL
    # leave it blank for default value
    # default value : https://itsyou.online/v1/oauth/access_token
    token_url       https://itsyou.online/v1/oauth/access_token

    # oauth2 redirect URL
    redirect_url    http://localhost:2015/_iyo_callback

    # Organizations allowed to access the protected paths
    # leave it blank if you want to ignore it
    organizations   /developer  mylab.developer
    organizations   /manager    mylab.manager

    # usernames allowed to access this protected paths
    # leave it blank to allow all usernames
    # - each username need to be separated with `,`
    # - you can specify it in multiple lines
    usernames       /manager    iwan

    # Everyone is allowed to access this path but authentication is required.
    # It is possible to specify this multiple times.
    authentication_required /
    
    # login_page to which the user will be redirected when trying to access authentication_required pages
    # leave blank if you need the users to be redirected to IYO page directly
    login_page  /login
    
    # login url is the URL that will redirect the user to itsyou.online login page
    # it can be used if you need to create login button
    login_url   /oauth

    # logout url is the URL that will logout the user and redirect him to "/"
    # it can be used if you need to create logout button
    logout_url   /logout

    # Allow specific files even if they are set in authentication_required
    # It is possible to specify this multiple times
    # typically used with static files (css, js, etc...) for login page
    allow_extension css

    # comma separated extra scopes to be requested from IYO, it can be blank
    extra_scopes user:address,user:email,user:phone
}

```

## Limitations

Limitations:

- each path can only specify one organization because of IYO limitation

## Build and Run in Development

Install from caddy fork on itsyou.online org
These files are patched to have the plugin registered.

in my pc the GOPATH is in ~/go

```bash
go get github.com/itsyouonline/caddy
rm -f $GOPATH/bin/caddy
cd $GOPATH/src/github.com/itsyouonline/caddy/caddy/
go get ./...
bash build.bash
cp caddy $GOPATH/bin/caddy
```

Create `Caddyfile` based on `Caddyfile.example` file

Run it
```
caddy
```

It will serve this directory

## Build and Run in Production

### Order your directive

As described in https://github.com/mholt/caddy/wiki/Writing-a-Plugin:-Directives#3-order-your-directive.

You need to add `oauth` in this [array](https://github.com/mholt/caddy/blob/d3860f95f59b5f18e14ddf3d67b4c44dbbfdb847/caddyhttp/httpserver/plugin.go#L314-L355).

You need to make sure to add it above `proxy` and `browse` if you want these resources to be protected and the `X-Iyo-Username` to be set to proxied sites.

### Plug in your Plugin

as described in https://github.com/mholt/caddy/wiki/Writing-a-Plugin:-Directives#4-plug-in-your-plugin
Add below import line in Caddy's [run.go](https://github.com/mholt/caddy/blob/master/caddy/caddymain/run.go)


```
_ "github.com/itsyouonline/caddy-integration/oauth"
```

### Build caddy

By executing [build.bash](https://github.com/mholt/caddy/blob/d3860f95f59b5f18e14ddf3d67b4c44dbbfdb847/caddy/build.bash)
