# VOID API Gateway

## Description
VOID is an API gateway with following features:

- service discovery via
  - [eureka](https://cloud.spring.io/spring-cloud-netflix/reference/html/)
  - [disco](https://github.com/slink-go/disco)
  - [static config](https://github.com/slink-go/void/blob/master/api-gateway/discovery/static_client.go)
- requests authentication via REST authentication service
- rate limiting (TBD)
- request timeouts (TBD)
- circuit breaker (TBD)

## Usage

docker-compose:
```yaml

services:
  void-gateway:
    deploy:
      resources:
        limits:
          memory: "50M"
    image: slinkgo/void:0.0.2
    container_name: void-gateway
    environment:

      - "GATEWAY_NAME=GW"
      - "SERVICE_PORT=3000"

      - "MONITORING_ENABLED=true"
      - "MONITORING_PORT=3001"

      - "REGISTRY_REFRESH_INITIAL_DELAY=2s"
      - "REGISTRY_REFRESH_INTERVAL=10s"

      - "AUTH_ENABLED=true"
      - "AUTH_ENDPOINT=http://auth/api/token/exchange"
      - "AUTH_RESPONSE_MAPPING_FILE_PATH=/auth_mapping.json"
      - "AUTH_CACHE_TTL=10s"

      - "EUREKA_CLIENT_ENABLED=true"
      - "EUREKA_URL=http://eureka:8761/eureka"
      - "EUREKA_LOGIN=eureka"
      - "EUREKA_PASSWORD=eureka"
      - "EUREKA_HEARTBEAT_INTERVAL=5s"
      - "EUREKA_REFRESH_INTERVAL=10s"

      - "DISCO_CLIENT_ENABLED=true"
      - "DISCO_URL=http://disco:8081"
      - "DISCO_LOGIN=disco"
      - "DISCO_PASSWORD=disco"
      - "DISCO_HEARTBEAT_INTERVAL=5s"
      - "DISCO_REFRESH_INTERVAL=10s"

      - "RATE_LIMIT_RPM=150"

      - "GO_ENV=dev"
      - "LOGGING_LEVEL_ROOT=INFO"
    ports:
      - "3021:3000"
      - "3022:3001"

```
Full config see [here](https://github.com/slink-go/void/blob/master/app/run/docker-compose.yml)

## Reverse Proxy
With dynamic discovery enabled (Disco and/or Eureka client is enabled) VOID tries to get target service name from request URL Path applying following path conversion rules:
```text
a) http://{host}/api/SERVICE-A/some/rest/endpoint     -> http://{SERVICE-A-HOST}:{SERVICE-A-PORT}/api/some/rest/endpoint
b) http://{host}/SERVICE-A/api/some/rest/endpoint     -> http://{SERVICE-A-HOST}:{SERVICE-A-PORT}/api/some/rest/endpoint
c) http://{host}/api/SERVICE-A/api/some/rest/endpoint -> http://{SERVICE-A-HOST}:{SERVICE-A-PORT}/api/some/rest/endpoint
d) http://{host}/SERVICE-A/some/rest/endpoint         -> http://{SERVICE-A-HOST}:{SERVICE-A-PORT}/some/rest/endpoint
```

If multiple instances are discovered for resolved service name, VOID will load balance between all of them using round-robin algorithm.

### Circuit breaker
TBD: implement circuit breaker for dead peers (if requests to some instance of service fail, this instance should be 
removed from load balancing until "circuit" is restored)

## Request Authentication
If `AUTH_ENABLED` flag is set to `true`, VOID tries to authenticate incoming requests. Authentication is performed on 
configured auth service (`AUTH_ENDPOINT`). Authentication in fact is exchanging auth token to user details data. 
Auth token is taken from `Autorization: Bearer` request header, or from `AuthToken` request cookie. If authentication 
is successful, VOID expects authentication server to send user details JSON in response. Then VOID injects user details into
request context via headers. Authentication server response is mapped to headers via mapping file. For example, 
authentication server sends following user details:
```json
{
  "id": 1,
  "email": "jsmith@email.com",
  "firstName": "John",
  "lastName": "Smith",
  "status": "ACTIVE",
  "org": {
    "id": 2,
    "active": true
  }
}

```
So VOID should be configured with following mapping file (`AUTH_RESPONSE_MAPPING_FILE_PATH`): 

```json
{
  "id": "Ctx-User-Id",
  "email": "Ctx-User-Email",
  "firstName": "Ctx-User-First-Name",
  "lastName": "Ctx-User-Last-Name",
  "status": "Ctx-User-Status",
  "org": {
    "id": "Ctx-Org-Id",
    "active": "Ctx-Org-Active"
  }
}
```
i.e. each field, which should be mapped to context, should be mentioned in mapping file with related context field (header name)

So in our example following request headers will be set:
```text
  Ctx-User-Id:          "1"
  Ctx-User-Email:       "jsmith@email.com"
  Ctx-User-First-Name:  "John"
  Ctx-User-Last-Name:   "Smith"
  Ctx-User-Status:      "ACTIVE"
  Ctx-Org-Id:           "2"
  Ctx-Org-Active:       "true"
```

These headers will be set before proxying request to backend services, thus allowing backend to have all the needed 
authentication data directly in request context.

To prevent authentication service overload, auth caching is used, so that once received user details for given auth token,
VOID won't make subsequent authentication requests to authentication service, until user details data is expired. Expiration
timeout is set via `AUTH_CACHE_TTL` variable (should be reasonably low value, i.e. 10-30 seconds). 

### Auth Skip
TBD: skip authentication for certain URL patterns

## Rate Limiting
TBD: implement waiting/denying rate limiters

## Request Timeouts
TBD: implement request timeouts (with configurable skip URL patterns)

## SSL Support
TBD: enable SSL support