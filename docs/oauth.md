# OAuth

expose directory over oauth integrated with IYO.

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

    # jwt from oauth URL
    # leave it blank for default value
    # default value : https://itsyou.online/v1/oauth/jwt
    jwt_url         https://itsyou.online/v1/oauth/jwt

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

    # set X-Iyo-Token header with the payload contents as json
    # by default, this directive is not set
    forward_payload
}

```

## Limitations

Limitations:

- each path can only specify one organization because of IYO limitation
