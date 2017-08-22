
create set of examples with
- https://github.com/tarent/loginsrv

You will need to use [caddyman](https://github.com/itsyouonline/caddyman/) to build caddy with the following plugins:
```text
github.com/BTBurke/caddy-jwt
github.com/tarent/loginsrv/caddy
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

## example 2: auth against github

use the browse directive to show content of example dir (1 file is enough)

## example 3: auth against IYO

use the browse directive to show content of example dir (1 file is enough)
do user of certain group all (all is all users in IYO as long as they are authenticated)


