# API Gateway: TODO

### Base
1. *~~Reverse proxy~~*
2. *~~Target resolver~~*
3. *~~Context setup~~*
4. *~~Rate limit~~*
5. *~~Log request~~*
6. *~~Static resolver (+ load balancing)~~*
7. *~~Eureka service resolver (+ load balancing)~~*
8. Disco service resolver
9. K8S service resolver
10. Multiple service resolvers support (static + eureka + disco) (?)
11. Cookie AuthToken support
12. AuthProvider chaining ( http header -> cookie -> ... )
13. Fallback (to default backend service) (?)
14. Profiling
15. Configuration & feature flags
16. Handle dead peers (connection refused)
17. Remote peers filter (?) (by meta, by status, ...)
18. Static resolver config from file
19. GinBasedProxy (handle internal routes) (?)
20. Client: advertise custom address / port (for specific deployment cases)



### Middleware
1. Auth check
2. Cyberrange auth provider | [common] rest-auth-provider
3. Timeout support
4. Metrics / latency measurement
5. Bulkhead / circuit breaker / etc
6. ~~*Limiter config*~~