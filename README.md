# API Gateway

### Base
1. *~~Reverse Proxy~~*
2. *~~Target Resolver~~*
3. *~~Context Setup~~*
4. *~~Rate Limit~~*
5. *~~Log Request~~*
6. *~~Static Resolver: Load Balancing~~*
7. Eureka Service Resolver (+ Load Balancing)
8. Disco Service Resolver
9. K8S Service Resolver
10. Multiple Service Resolvers Support (static + eureka + disco)
11. Cookie AuthToken Support
12. Fallback (to default backend service) (?)
13. Profiling
14. Configuration & feature flags
15. Handle dead peers


### Middleware
1. Auth Check
2. Cyberrange Auth Provider
3. Timeout Support
4. Metrics / Latency Measurement
5. Bulkhead / Circuit Breaker / etc

---

> RegistrarClient   ->  backend service

> RegistryClient    -> gateway

EurekaClient
- fetch registry
- self register
