# API Gateway: TODO

### Base
1. [+] Reverse proxy
2. [+] Target resolver
3. [+] Context setup
4. [+] Rate limit
5. [+] Log request
6. [+] Static resolver (+ load balancing)
7. [+] Eureka service resolver (+ load balancing)
8. [+] Disco service resolver
9. [-] K8S service resolver
10. [+] Multiple service resolvers support (static + eureka + disco)
11. [+] Cookie AuthToken support
12. [+] AuthProvider chaining ( http header -> cookie -> ... )
13. [o] ~~Fallback (to default backend service) (?)~~
14. [-] Profiling
15. [+] Configuration & feature flags
16. [-] Handle dead peers (connection refused, host unreachable, etc)
17. [-] Remote peers filter (?) (by meta, by status, ...)
18. [+] Static resolver config from file
19. [-] Client: advertise custom address / port (for specific deployment cases) - using META
20. [-] advertise app info url (?)
21. [-] gRPC reverse proxy  (see: https://habr.com/ru/articles/645433/)
22. [+] Pattern matcher
23. [-] CSRF (?)
24. [-] Correct errors handling
25. [-] ENHANCED Pattern matcher 
26. [-] Conditional Timeout Middleware
27. [-] CORS config

### Middleware
1. [+] Auth check
2. [+] rest-auth-provider
3. [+] Timeout support (except sse/ws). Limitation: with timeout enabled streaming services does not work (all streaming endpoints should be skipped for timeout processing) 
4. [-] Metrics / latency measurement
5. [-] Bulkhead / circuit breaker / etc
6. [+] Limiter config
7. [+] Auth cache middleware
   - [+] inmem
   - [-] redis
8. [+] Rate Limiter (DELAY, DENY)
9. [-] Cookie Auth: configurable cookie name
10. [-] Use disco-client resolving capabilities (falling back to HostResolve)

### URL Pattern Matching
1. [+] auth skip urls
2. [+] timeout skip urls
3. [+] rate limit: custom config

### Procedure
```text
- latency-middleware::start
+ rate-limiter-middleware
+ proxy-target-resolver-middleware -> url -> proxy-url : (considering circuit breaker)
- circuit breaker middleware
+ headers-cleanup-middleware::start
+ auth-resolver-middleware 	-> 	header/cookie -> ctx.Set(Auth) => check token validity
+ auth-cache-middleware -> Auth -> ctx.Set(UserDetails)
+ auth-provider-middleware	->	get user details from remote peer
    + token exchange (in case of correct Auth) -> ctx.Set(UserDetails)
    - backoff
    - circuit breaker
+ locale-resolver-middleware
+ context-configurer-middleware -> set request headers (ctx.Get(UserDetails) -> ctx.SetHeader(...))
+ do proxy
- headers-cleanup-middleware::finish
- latency-middleware::finish -> set custom metrics
```

## Security
- [!!!] support "trusted proxies" for rate limiter, reverse proxy ("trusted proxy middleware") (see [here](https://adam-p.ca/blog/2022/03/x-forwarded-for/#thoughts-on-overwriting-the-xff-header))
- Good example of "detecting" client address: [envoy xff](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/headers#x-forwarded-for)
  - use_remote_address = { true | false }
  - xff_num_trusted_hops = N
  - trusted proxies (?)


### Useful links
- WS Proxy: 	
   - https://stackoverflow.com/questions/73187877/how-do-i-implement-a-wss-reverse-proxy-as-a-gin-route
   - https://dev.to/hgsgtk/reverse-http-proxy-over-websocket-in-go-part-1-13n4
   - https://gist.github.com/seblegall/2a2697fc56417b24a7ec49eb4a8d7b1b
- CSRF protection (?)
   - https://www.stackhawk.com/blog/golang-csrf-protection-guide-examples-and-how-to-enable-it/
   - https://github.com/utrack/gin-csrf
- Circuit Breaker:
   - https://gist.github.com/jerryan999/bcfdd746f3f8c2c11c3d27f1b65dfcf3
   - https://pkg.go.dev/github.com/go-kit/kit/circuitbreaker
   - https://medium.com/german-gorelkin/go-patterns-circuit-breaker-921a7489597
   - !!! https://dev.to/he110/circuitbreaker-pattern-in-go-43cn
   - !!! https://github.com/sony/gobreaker
- Hystrix:
   - https://github.com/afex/hystrix-go
- Middleware:
   - https://github.com/orgs/gin-contrib/repositories
   - https://github.com/gin-gonic/contrib/tree/master/gzip
