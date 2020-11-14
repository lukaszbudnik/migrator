# Securing migrator with OIDC

In this tutorial I will show you how to use OIDC on top of OAuth2 to implement authentication and authorization.

If you are interested in simple OAuth2 authorization see [Securing migrator with OAuth2](../tutorials/oauth2-proxy).

## OIDC

OpenID Connect (OIDC) is an authentication layer on top of OAuth2 authorization framework.

I will use oauth2-proxy project. It supports multiple OAuth2 providers. To name a few: Google, Facebook, GitHub, LinkedIn, Azure, Keycloak, login.gov, or any OpenID Connect compatible provider.

As an OIDC provider I will re-use oauth2-proxy local-environment which creates and setups a ready-to-use Keycloak server. I extended the Keycloak server with additional configuration (client mappers to include roles in responses), created two migrator roles and two additional test accounts.

I also put haproxy between oauth2-proxy and migrator. haproxy will validate the JWT access token and implement access control based on user's roles to allow or deny access to underlying migrator resources. I re-used a great lua script written by haproxytech folks which I modified to work with Keycloak realm roles.

To learn more about oauth2-proxy visit https://github.com/oauth2-proxy/oauth2-proxy.

To learn more about Keycloak visit https://www.keycloak.org.

To learn more about haproxy jwtverify lua script visit https://github.com/haproxytech/haproxy-lua-jwt.

## Docker setup

The provided `docker-compose.yaml` provision the following services:

* keycloak - the Identity and Access Management service, available at: http://keycloak.localtest.me:9080
* oauth2-proxy - proxy that protects migrator and connects to keycloak for OAuth2/OIDC authentication, available at: http://gateway.localtest.me:4180
* haproxy - proxy that contains JWT access token validation and user access control logic, available at: http://haproxy.localtest.me:8080
* migrator - deployed internally and accessible only from haproxy and only by authorized users

> Note: above setup doesn't have a database as this is to only illustrate how to setup OIDC

To build the test environment execute:

```
docker-compose up -d
```

## Testing OIDC

I created 2 test users in Keycloak:

* `madmin@example.com` - migrator admin, has the following roles: `migrator_admin` and `migrator_user`
* `muser@example.com` - migrator user, has one migrator role: `migrator_user`

In haproxy.cfg I implemented the following sample rules:

* all requests starting `/v1` will return 403 Forbidden
* to access `/v2/service` user must have `migrator_user` role
* to access `/v2/config` user must have `migrator_admin` role

### Test scenarios

There are two ways to access migrator:

1. getting JWT access token via oauth2-proxy - shown in orange
1. getting JWT access token directly from Keycloak - shown in blue

![migrator OIDC setup](migrator-oidc.png?raw=true)

Let's test them.

### oauth2-proxy - muser@example.com

1. Access http://gateway.localtest.me:4180/
1. Authenticate using username: `muser@example.com` and password: `password`.
1. After a successful login you will see `/` response
1. Open http://gateway.localtest.me:4180/v2/config and you will see 403 Forbidden - this user doesn't have `migrator_admin` role
1. Logout from Keycloak http://keycloak.localtest.me:9080/auth/realms/master/protocol/openid-connect/logout
1. Invalidate session on auth2-proxy (oauth2-proxy cookie expires in 15 minutes) http://gateway.localtest.me:4180/oauth2/sign_out

### oauth2-proxy - madmin@example.com

1. Access http://gateway.localtest.me:4180/
1. Authenticate using username: `madmin@example.com` and password: `password`.
1. After a successful login you will see successful `/` response
1. Open http://gateway.localtest.me:4180/v2/config and now you will see migrator config
1. Logout from Keycloak http://keycloak.localtest.me:9080/auth/realms/master/protocol/openid-connect/logout
1. Invalidate session on auth2-proxy (oauth2-proxy cookie expires in 15 minutes) http://gateway.localtest.me:4180/oauth2/sign_out

### Keycloak REST API

1. Get JWT access token for the `madmin@example.com` user:

```
access_token=$(curl -s http://keycloak.localtest.me:9080/auth/realms/master/protocol/openid-connect/token \
    -H 'Content-Type: application/x-www-form-urlencoded' \
    -d 'username=madmin@example.com' \
    -d 'password=password' \
    -d 'grant_type=password' \
    -d 'client_id=oauth2-proxy' \
    -d 'client_secret=72341b6d-7065-4518-a0e4-50ee15025608' | jq -r '.access_token')
```

2. Execute migrator action and pass the JWT access token in HTTP Authorization header:

```
curl http://haproxy.localtest.me:8080/v2/config \
    -H "Authorization: Bearer $access_token"
```

## Miscellaneous

You can copy JWT access token (haproxy log or Keycloak REST API) and decode it on https://jwt.io.

You can verify the signature of the JWT token by providing the public key (`keycloak.pem` available in haproxy folder).

Public key can be also fetched from:

```
curl http://keycloak.localtest.me:9080/auth/realms/master/
```

> The response is a JSON and the public key is returned as a string. To be a valid PEM format you need to add `-----BEGIN PUBLIC KEY-----` header, `-----END PUBLIC KEY-----` footer, and break that string into lines of 64 characters. Compare `keycloak.pem` with the above response.
