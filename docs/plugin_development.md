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
