# caddy oauth plugin

This plugin protects resource paths using itsyou.online oauth2.

## why

- [why](why.md)


## to play with it on your local machine

### Pre-requisites

- install bash tools:

https://github.com/Jumpscale/bash

Then create a docker which has our docgenerator inside as well as caddy and all other required tools. This will take some time.

```bash
#install & run the doc_generator (make sure there are no other dockers running on port 2222)
ZInstall_docgenerator
```

### get this repository
```
js9_code get --url git@github.com:itsyouonline/caddy-integration.git
```

### play

go to the examples directories & play

## Run it

```
cd ..directorywhere Caddyfile is
caddy
```

It will serve this directory

## Tech Features

Plugin features:

- protects paths based on organization membership
- protects paths based on username
- use JWT to make it stateless and reduce API calls to Oauth2 server
- log following infos to stdout : host, time, http verb, path, http method, username
- sets a `X-Iyo-Username` header with the username of the logged in user