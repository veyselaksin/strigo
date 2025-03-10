# StriGo Examples

This directory contains example implementations of the StriGo rate limiter using different storage backends and scenarios.

## Prerequisites

Before running the examples, make sure you have the following services running:

- Redis server (default: localhost:6379)
- Memcached server (default: localhost:11211)

## Examples

### Basic Usage (`basic/main.go`)

Demonstrates standalone usage of StriGo with both Redis and Memcached backends. This example shows:
- Basic rate limiter configuration
- Direct usage without web framework
- Request simulation

To run:
```bash
go run basic/main.go
```

### Web Server (`web/main.go`)

Shows integration with the Fiber web framework, demonstrating:
- Middleware usage
- Multiple rate limit rules
- Different strategies and time windows
- Both Redis and Memcached backends

To run:
```bash
go run web/main.go
```

Then test the endpoints:
- Redis endpoint: `http://localhost:3000/redis`
- Memcached endpoint: `http://localhost:3000/memcached`
- Advanced endpoint: `http://localhost:3000/advanced`

## Testing with curl

Test the web server endpoints:

```bash
# Test Redis endpoint
for i in {1..6}; do curl http://localhost:3000/redis; echo; done

# Test Memcached endpoint
for i in {1..6}; do curl http://localhost:3000/memcached; echo; done

# Test Advanced endpoint
for i in {1..11}; do curl http://localhost:3000/advanced; echo; done
```
```

These examples demonstrate:
1. Basic usage of both Redis and Memcached backends
2. Web framework integration with Fiber
3. Different rate limiting strategies
4. Multiple rules and time windows
5. Proper error handling and resource cleanup
6. Real-world usage patterns

The examples are well-documented and include instructions for testing. Users can easily modify the configurations to test different scenarios or adapt the code for their own needs.