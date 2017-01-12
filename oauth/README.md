# caddy oauth middleware

## using it in development

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

## using it in production

Add below import line in Caddy's [run.go](https://github.com/mholt/caddy/blob/master/caddy/caddymain/run.go)
```
_ "github.com/itsyouonline/caddy-integration/oauth"
```

Then build caddy as usual
