
create set of examples with
- https://github.com/tarent/loginsrv

You will need to use [caddyman](https://github.com/itsyouonline/caddyman/) to build caddy with the following plugins:
```text
github.com/BTBurke/caddy-jwt
github.com/itsyouonline/loginsrv/caddy
```
## example 1: [basic authentication with httpasswd file](./htpasswd_example)

- Create htpasswd file with format:
```text
<USER>:<ENCRYPTED_PASSWORD>
```
demo credentials: (user: `demo`, password: `demo`)

_You can encrypt password using MD5, SHA1 and Bcrypt are supported. But using bcrypt is recommended for security reasons:_
```bash
htpasswd -n -b -B -C 15 USER PASSWORD
```

- Add password file path to Caddyfile as in [example](./htpasswd_example/Caddyfile)

- Run Caddy

You can use `template` directive in the `Caddyfile` if you need to use customized login page instead of the default one

## example 2: [auth against github](./github_example)

- Create an oauth application on [github](http://github.com) and set the `Authorization callback URL` to `YOUR_HOST/login/github`

- set client_id and client_secret from bash:
```bash
export github_client_id="<CLIENT_ID>"
export github_client_secret="<CLIENT_SECRET>"
```

- run Caddy
You can use `template` directive in the `Caddyfile` if you need to use customized login page instead of the default one

## example 3: [auth against IYO](./iyo_example)

- set client_id and client_secret from bash:
```bash
export iyo_client_id="<CLIENT_ID>"
export iyo_client_secret="<CLIENT_SECRET>"
```

- run Caddy
You can use `template` directive in the `Caddyfile` if you need to use customized login page instead of the default one

use the browse directive to show content of example dir (1 file is enough)
do user of certain group all (all is all users in IYO as long as they are authenticated)


