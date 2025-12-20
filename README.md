# Load Balancer

**A Layer 7 HTTP load balancer written in Go, with a focus on concurrency, high performance, and clean architecture.**

## Key Features

- **Thread-safe Round Robin** using atomic operations → minimal locking and high throughput
- **Atomic-cached healthy backend pool** → zero-allocation healthy backend retrieval
- **Pluggable load balancing strategies** (Strategy Pattern) → easy to extend with new algorithms
- **Health monitoring** with unified Monitor → reliable backend availability
- **Context propagation** throughout request flow → proper cancellation and tracing
- **Structured logging with zap** → high-performance, correlated logs
- **Graceful shutdown** with signal handling and configurable timeout → zero-downtime deployments

## Project Structure

```
internal/          - Private application code
├── app/           - Application orchestration layer
├── balancer/      - Business logic (Strategy pattern)
├── config/        - Configuration management
├── health/        - Health monitoring domain
├── middleware/    - HTTP middleware chain
├── pool/          - Backend pool with atomic caching
├── proxy/         - HTTP reverse proxy implementation
└── server/        - HTTP server layer
```

Clean architecture with proper dependency flow. Each package has one responsibility, and everything depends on abstractions, not concretions.

## How It Works

### Request Flow

```
Client → HTTP Handler → Director → Strategy → Proxy → Backend → Response
```

1. Client sends an HTTP request
2. Request hits the HTTP Handler
3. Handler calls the Director
4. Director retrieves cached healthy backends from the Pool
5. Director passes the healthy backends to the Strategy
6. Strategy selects the next backend (```Next()```)
7. Handler forwards the request to the Proxy
8. Proxy sends the request to the selected backend and streams back the response
9. Handler returns the response to the client.

### Health Monitoring

Separate goroutine does health checks every N seconds:
- Hits each backend with a GET request
- Marks unhealthy if not 2xx-3xx response
- Auto-recovers when backends come back
- Logs status changes

### Load Balancing

Current implementation uses Round Robin:

```go
func (rr *RoundRobin) Next(backends []*pool.Backend) *pool.Backend {
    n := atomic.AddUint64(&rr.counter, 1) - 1
    return backends[n%uint64(len(backends))]
}
```

Thread-safe with atomic operations, distributes evenly across healthy backends. Easy to extend with other strategies.

## Quick Start

### Using Docker Compose (Recommended)

```bash
# Clone and navigate to the project
git clone https://github.com/extndr/load-balancer.git
cd load-balancer
# Start the demo environment with 3 backends and the load balancer
docker-compose -f demo/docker-compose.yml up -d
# Test the load balancer
curl http://localhost:1234
# Make multiple requests to see round-robin in action
for i in {1..12}; do curl http://localhost:1234; echo; done
# Stop the environment
docker-compose -f demo/docker-compose.yml down
```

## Why I Built This

I started this project as a simple exercise to explore Go’s concurrency model and horizontal scaling.  
At first, it was just a basic load balancer without health checks, proper separation of concerns, or any architectural discipline.  

Over time, it turned into a solid system with small, focused components that have clear boundaries.  
Backends are managed safely, health is monitored, and the design makes it easy to extend.  
Each part works on its own and is simple to understand, test, and maintain.

**Deliberate exclusions:**

- No service discovery like etcd → that's becoming Kubernetes territory
- No CB pattern and rate limiting → that's becoming API Gateway territory
- No production-grade test suite → focused on patterns and performance rather than coverage

**Future improvements:**

- Live reload of backend pool via API → allows updating backends without restarting the load balancer.
- Replace demo backends (hashicorp/http-echo) with real HTTP services and update health checks to target a dedicated `/health` endpoint → for more realistic application health semantics.
