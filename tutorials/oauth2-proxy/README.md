# Securing migrator with OAuth2

In this tutorial I will show you how to setup OAuth2 authorization in front of migrator.

## oauth2-proxy

I will use oauth2-proxy project. It supports multiple OAuth2 providers. To name a few: Google, Facebook, GitHub, LinkedIn, Azure, Keycloak, login.gov, or any OpenID Connect compatible provider. For the sake of simplicity I will re-use oauth2-proxy local-environment which creates and setups a ready-to-use Keycloak server.

To learn more about oauth2-proxy visit https://github.com/oauth2-proxy/oauth2-proxy.

To learn more about Keycloak visit https://www.keycloak.org.

## Docker setup

I re-used `docker-compose.yaml` from oauth2-proxy local-environment and updated it to provision the following services:

* keycloak - the Identity and Access Management service, available at: http://keycloak.localtest.me:9080
* oauth2-proxy - proxy that protects migrator and connects to keycloak for OAuth2 authorization, available at: http://gateway.localtest.me:4180
* migrator - deployed internally and accessible only from oauth2-proxy and only by authorized users

> Note: above setup doesn't have a database as this is to only illustrate how to setup oauth2-proxy.

To build the test environment execute:

```
docker-compose up -d
```

Access http://gateway.localtest.me:4180/ to initiate a login cycle and authenticate with user `admin@example.com` and password `password`. After a successful login you will see migrator.

Access http://keycloak.localtest.me:9080 to play around with Keycloak.

## GitHub setup

OAuth2 can be easily setup in GitHub. Open `oauth2-proxy.cfg` and comment lines 9 and below. Follow https://docs.github.com/en/free-pro-team@latest/developers/apps/creating-an-oauth-app documentation to create OAuth2 application and then setup the following 3 parameters in the config file:

```
provider="github"
client_id="XXX"
client_secret="XXX"
```

In case of GitHub you can use additional out-of-the-box features to limit users who can access migrator. For example you can limit access to particular users, team, repository, or organisation. For a full list of GitHub provider features check out oauth2-proxy documentation: https://oauth2-proxy.github.io/oauth2-proxy/auth-configuration#github-auth-provider.
