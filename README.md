# StriGo 🦉

**StriGo** is a high-performance rate limiter for Go, designed to work seamlessly with Redis, Memcached, and Dragonfly. It provides efficient and scalable request limiting to protect your applications from abuse and ensure fair resource distribution.

## Features 🚀

- 🔥 **High-performance**: Optimized for speed and efficiency.
- 🔄 **Supports multiple backends**: Redis, Memcached, and Dragonfly.
- 🛡 **Protects against abuse**: Prevents excessive API requests.
- 📏 **Flexible limit strategies**: Token Bucket, Leaky Bucket, Fixed Window, and Sliding Window.
- 📦 **Lightweight and easy to use**: Simple API for seamless integration.

## Installation 📦

```sh
go get github.com/veyselaksin/strigo
```

## Quick Start ⚡

See [examples](examples) for usage examples.

## Usage 🛠

### Creating a Rate Limiter

#### Redis Backend
```go
 limiter, err := strigo.NewLimiter(strigo.Redis, "localhost:6379")
 if err != nil {
    log.Fatal(err)
 }
 defer limiter.Close()
```

#### Memcached Backend
```go
 limiter, err := strigo.NewLimiter(strigo.Memcached, "localhost:11211")
 if err != nil {
    log.Fatal(err)
 }
 defer limiter.Close()
```

### Checking Rate Limits
```go
if limiter.Allow("user-123") {
	fmt.Println("Request allowed")
} else {
	fmt.Println("Rate limit exceeded")
}
```

## Limit Strategies 📊

StriGo supports multiple rate limiting strategies:

- **Token Bucket** (default)
- **Leaky Bucket**
- **Fixed Window**
- **Sliding Window**

Example:
```go
limiter.SetStrategy(strigo.SlidingWindow)
```

## Contributing 🤝

Contributions are welcome! Feel free to open issues or submit pull requests.

## License 📜

MIT License. See [LICENSE](LICENSE) for details.

---

Made with ❤️ by [Your Name](https://github.com/yourusername)
