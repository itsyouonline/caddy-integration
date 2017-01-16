# caddy oauth plugin

This plugin protects resource paths using itsyou.online oauth2.

## Features

Plugin features:

- protects paths based on organization membership
- protects paths based on username
- use JWT to make it stateless and reduce API calls to Oauth2 server
- log following infos to stdout : host, time, http verb, path, http method, username

## Usage
Add oauth block to Caddyfile

example
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
}

```

## Limitations

Limitations:

- each path can only specify one organization because of IYO limitation

## Build and Run in Development

Install caddydev
```
go get github.com/caddyserver/caddydev
```

Create `Caddyfile` based on `Caddyfile.example` file

Run it
```
caddydev
```

It will serve this directory

## Build and Run in Production

Add below import line in Caddy's [run.go](https://github.com/mholt/caddy/blob/master/caddy/caddymain/run.go)
```
_ "github.com/itsyouonline/caddy-integration/oauth"
```

Then build caddy as usual
