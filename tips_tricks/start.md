
## use https & make sure that you auto agree letsencrypt

```
./caddy --log stdout --agree=true -conf Caddyfile -alsologtostderr --email=admin@example.com
```

## specify manually the letsecnrypt directory (DO THIS FOR TESTING, to avoid rate limiting) !!!

```
./caddy --log stdout --agree=true -conf Caddyfile -alsologtostderr -ca https://acme-staging.api.letsencrypt.org/directory --email=admin@example.com
```
