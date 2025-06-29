# StriGo Examples

This directory contains example implementations of the StriGo rate limiter using different storage backends, architectures, and scenarios.

## Prerequisites

Before running the examples, make sure you have the following services running:

- Redis server (default: localhost:6379)
- Memcached server (default: localhost:11211)

## Examples

### 🚀 Web Server (`web/`) - **Recommended**

A comprehensive example showing integration with the Fiber web framework. This example demonstrates:

- **Multiple rate limiting scenarios**: API endpoints with different limits
- **Both Redis and Memcached**: Support for different storage backends
- **Middleware usage**: Clean integration with web framework
- **Different strategies**: Various algorithms and time windows
- **Real-world patterns**: Error handling and response management

**Features:**

- ✅ **Production-ready**: Error handling and proper middleware integration
- ✅ **Multiple backends**: Redis, Memcached, and memory storage
- ✅ **Framework integration**: Clean Fiber middleware patterns
- ✅ **Easy to test**: Simple curl commands for testing
- ✅ **Well-documented**: Clear examples and usage patterns

To run:

```bash
cd web
go run main.go
```

### Basic Usage (`basic/`)

Demonstrates standalone usage of StriGo with both Redis and Memcached backends. This example shows:

- Basic rate limiter configuration
- Direct usage without web framework
- Request simulation and testing

To run:

```bash
cd basic
go run main.go
```

## Testing Examples

### Test the Web Server (Recommended)

```bash
cd web
go run main.go
```

Then test the endpoints in another terminal:

```bash
# Test Redis endpoint (5 requests per 10 seconds)
for i in {1..6}; do curl http://localhost:3000/redis; echo; done

# Test Memcached endpoint (3 requests per 5 seconds)
for i in {1..4}; do curl http://localhost:3000/memcached; echo; done

# Test Advanced endpoint (10 requests per minute)
for i in {1..11}; do curl http://localhost:3000/advanced; echo; done
```

Expected output for rate limiting:

```
✅ Request allowed! Remaining: 4 requests
✅ Request allowed! Remaining: 3 requests
✅ Request allowed! Remaining: 2 requests
✅ Request allowed! Remaining: 1 requests
✅ Request allowed! Remaining: 0 requests
❌ Rate limit exceeded! Try again in 10 seconds
```

### Test Basic Example

```bash
cd basic
go run main.go
```

## Example Comparison

| Example | Architecture | Complexity | Production Ready | Recommended For            |
| ------- | ------------ | ---------- | ---------------- | -------------------------- |
| **web** | **Single**   | **Medium** | **✅ Yes**       | **Web applications**       |
| basic   | Standalone   | Low        | Partial          | Learning and understanding |

## Key Features Demonstrated

1. **Web Framework Integration**: Clean Fiber middleware patterns
2. **Multiple Storage Backends**: Redis, Memcached, and in-memory storage
3. **Different Rate Limiting Strategies**: Various algorithms and time windows
4. **Error Handling**: Graceful error responses and fallbacks
5. **Real-world Usage**: Practical examples for web applications
6. **Easy Testing**: Simple curl commands for validation

## Getting Started

For web applications, we recommend starting with the **web** example as it demonstrates best practices for integrating StriGo with web frameworks and provides practical patterns for production use.

The examples are well-documented and include comprehensive testing instructions. Users can easily modify the configurations to test different scenarios or adapt the code for their own applications.
