# VOID API Gateway

## Description
VOID is an API gateway with following features:

- service discovery via
  - [eureka](https://cloud.spring.io/spring-cloud-netflix/reference/html/)
  - [disco](https://github.com/slink-go/disco)
  - [static config](https://github.com/slink-go/void/blob/master/api-gateway/discovery/static_client.go)
- requests authentication via REST authentication service
- rate limiting
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
    image: slinkgo/void
    container_name: void-gateway
    environment:
      - "GATEWAY_NAME=GW"
      - "SERVICE_PORT=3000"
      - "MONITORING_ENABLED=true"
      - "MONITORING_PORT=3001"
      ...
    ports:
      - "3021:3000" 
      - "3022:3001"
```
Full config see [here](https://github.com/slink-go/void/blob/master/app/run/docker-compose.yml)

## Configuration
VOID configuration is performed via following environment variables:

| Variable                                              | Description                                                                     |
|-------------------------------------------------------|---------------------------------------------------------------------------------|
| **BASE**                                              |                                                                                 |
| `GATEWAY_NAME="GW"`                                   | Gateway name to register on Disco service                                       |
| `SERVICE_PORT=3000`                                   | Service port to listen on                                                       |
| `MONITORING_ENABLED=true`                             | Monitoring is enabled (if true, monitoring WebUI is started on monitoring port) |
| `MONITORING_PORT=3001`                                | Monitoring port to listen on                                                    |
| **REGISTRY**                                          |                                                                                 |
| `REGISTRY_REFRESH_INITIAL_DELAY=2s`                   | Discovered services registry refresh initial delay                              |
| `REGISTRY_REFRESH_INTERVAL=10s`                       | Discovered services registry refresh interval                                   |
| **EUREKA DISCOVERY**                                  |                                                                                 |
| `EUREKA_CLIENT_ENABLED=true`                          | Enable target service discovery via Eureka                                      |
| `EUREKA_URL=http://eureka:8761/eureka"`               | Eureka URL                                                                      |
| `EUREKA_LOGIN=eureka`                                 | Eureka login                                                                    |
| `EUREKA_PASSWORD=eureka`                              | Eureka password                                                                 |
| `EUREKA_HEARTBEAT_INTERVAL=5s`                        | Eureka heartbeat interval                                                       |
| `EUREKA_REFRESH_INTERVAL=10s`                         | Eureka refresh interval                                                         |
| **DISCO DISCOVERY**                                   |                                                                                 |
| `DISCO_CLIENT_ENABLED=true`                           | Enable target discovery via Disco                                               |
| `DISCO_URL=http://disco:8081`                         | Disco URL                                                                       |
| `DISCO_LOGIN=disco`                                   | Disco login                                                                     |
| `DISCO_PASSWORD=disco`                                | Disco password                                                                  |
| **STATIC DISCOVERY**                                  |                                                                                 |
| `STATIC_REGISTRY_FILE=./routes/registry.yml`          | Static target configuration file (json or yaml)                                 |
| **AUTHENTICATION**                                    |                                                                                 |
| `AUTH_ENABLED=true`                                   | Enable incoming requests authentication on external authentication service      |
| `AUTH_ENDPOINT=http://auth/api/token/exchange"`       | External authentication service URL                                             |
| `AUTH_METHD=GET`                                      | HTTP Method to access Authentication service (default is GET)                   |
| `AUTH_RESPONSE_MAPPING_FILE_PATH=/auth_mapping.json"` | Authentication response mapping configuration file                              |
| `AUTH_CACHE_TTL=10s`                                  | Authentication data cache TTL                                                   |
| **RATE LIMIT**                                        |                                                                                 |
| `LIMITER_MODE=DENY`                                   | Rate limiter mode (OFF, DENY, DELAY)                                            |
| `LIMITER_LIMIT=1`                                     | Rate limiter global limit (requests per time interval)                          |
| `LIMITER_PERIOD=1s`                                   | Rate limiter global time interval                                               |
| `LIMITER_CUSTOM="*/service-a/*:1:5s,..."`             | Rate limiter custom config (per path pattern): "<pattern>:<limit>:<period>,..." |
| **LOGGING**                                           |                                                                                 |
| `GO_ENV=dev`                                          | "dev" - enable "pretty" log, otherwise structured logging is used               |
| `LOGGING_LEVEL_ROOT=INFO`                             | Root logging level                                                              |
| `LOGGING_LEVEL_EUREKA_CLIENT=INFO`                    |                                                                                 |
| `LOGGING_LEVEL_DISCO_CLIENT=INFO`                     |                                                                                 |
| `LOGGING_LEVEL_STATIC_CLIENT=INFO`                    |                                                                                 |
| `LOGGING_LEVEL_DISCOVERY_REGISTRY=INFO`               |                                                                                 |
| `LOGGING_LEVEL_SERVICE_RESOLVER=INFO`                 |                                                                                 |
| `LOGGING_LEVEL_DISCO_GO=INFO`                         |                                                                                 |
| `LOGGING_LEVEL_DISCO_GO_REG=INFO`                     |                                                                                 |
| `LOGGING_LEVEL_RESOLVER_MIDDLEWARE=INFO`              |                                                                                 |
| `LOGGING_LEVEL_CONTEXT_MIDDLEWARE=INFO`               |                                                                                 |
| `LOGGING_LEVEL_CR_USER_DETAILS_PROVIDER=INFO`         |                                                                                 |
| `LOGGING_LEVEL_USER_DETAILS_CACHE=INFO`               |                                                                                 |
| `LOGGING_LEVEL_HEADER_AUTH_PROVIDER=INFO`             |                                                                                 |
| `LOGGING_LEVEL_AUTH_CHAIN=INFO`                       |                                                                                 |
| `LOGGING_LEVEL_GLOBAL_RATE_LIMITER=ERROR`             |                                                                                 |
| `LOGGING_LEVEL_GIN_GATEWAY=INFO`                      |                                                                                 |
| `LOGGING_LEVEL_RATE_LIMITER=TRACE`                    |                                                                                 |
| `LOGGING_LEVEL_GIN=WARN`                              |                                                                                 |

## Static Discovery
Static discovery is configured in `STATIC_REGISTRY_FILE`. If this variable is not set, or if configuration file can't be read, static discovery is not used. Configuration can be in either JSON or YAML format.
- JSON Example
```json
[
  {
    "name": "service-a",
    "instances": [
      "http://backend:3101",
      "http://backend:3102",
      "backend:3103"
    ]
  },
  {
    "name": "service-b",
    "instances": [
      "backend:3201",
      "http://backend:3202",
      "backend:3203"
    ]
  }
]
```
- YAML Example
```yaml
- name: service-a
  instances:
    - http://backend:3101
    - http://backend:3102
    - http://backend:3103
- name: service-b
  instances:
    - http://backend:3201
    - http://backend:3202
    - http://backend:3203
```

## Reverse Proxy
For incoming request to be proxied, VOID needs to resolve target service name & URL. Following rules are applied for service resolving: 
```text
a) http://{host}/api/SERVICE-A/some/rest/endpoint     -> http://{SERVICE-A-HOST}:{SERVICE-A-PORT}/api/some/rest/endpoint
b) http://{host}/SERVICE-A/api/some/rest/endpoint     -> http://{SERVICE-A-HOST}:{SERVICE-A-PORT}/api/some/rest/endpoint
c) http://{host}/api/SERVICE-A/api/some/rest/endpoint -> http://{SERVICE-A-HOST}:{SERVICE-A-PORT}/api/some/rest/endpoint
d) http://{host}/SERVICE-A/some/rest/endpoint         -> http://{SERVICE-A-HOST}:{SERVICE-A-PORT}/some/rest/endpoint
```

If multiple instances are discovered for resolved service name, VOID will load balance between all of them using round-robin algorithm.

### Circuit breaker
> TBD: implement circuit breaker for dead peers (if requests to some instance of service fail, this instance should be 
removed from load balancing until "circuit" is restored)

## Request Authentication
If `AUTH_ENABLED` flag is set to `true`, VOID tries to authenticate incoming requests. Authentication is performed on 
configured auth service (`AUTH_ENDPOINT`). Authentication in fact is exchanging auth token to user details data. 
Auth token is taken from `Autorization: Bearer` request header, or from `AuthToken` request cookie. If authentication 
is successful, VOID expects authentication server to send user details JSON in response. Then VOID injects user details into
request context via headers. Authentication server response is mapped to headers via mapping file. For example, 
authentication server replies to successful authentication with following user details:
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
i.e. each field, which should be mapped to context, has to be mentioned in mapping file with related context field (header name)

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
> TBD: skip authentication for certain URL patterns

## Rate Limiting
> TODO: document this feature

## Request Timeouts
> TBD: implement request timeouts (with configurable skip URL patterns)

## SSL Support
> TBD: enable SSL support