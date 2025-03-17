# Golang Transport Middleware

A lightweight and modular HTTP transport middleware for Golang that enhances `http.RoundTripper` with **retry**, **logging**, and **circuit breaker** functionalities.

## Features

✅ **Retry Mechanism** - Automatically retry failed requests based on customizable strategies.  
✅ **Logging** - Logs request and response details with support for redacting sensitive information.  
✅ **Circuit Breaker** - Prevents system overload by stopping requests when failures exceed a threshold.  

## Installation

```sh
go get github.com/dangnmh/transport
```

## Usage

### Basic Setup

```go
client := &http.Client{
    Transport: NewTransportLog(http.DefaultTransport),
}

req, _ := http.NewRequest("GET", "https://api.example.com", nil)
resp, err := client.Do(req)
```

### Using Retry Transport

```go
client := &http.Client{
    Transport: NewRetryTransport(http.DefaultTransport, WithMaxRetries(3)),
}
```

### Using Circuit Breaker

```go
client := &http.Client{
    Transport: NewCircuitBreakerTransport(http.DefaultTransport, WithBreakerSettings(gobreaker.Settings{
        MaxRequests: 5,
        Interval:    10 * time.Second,
    })),
}
```

### Combining Features

```go
client := &http.Client{
    Transport: NewCircuitBreakerTransport(
        NewRetryTransport(
            NewTransportLog(http.DefaultTransport),
            WithMaxRetries(3),
        ),
        WithBreakerSettings(gobreaker.Settings{
            MaxRequests: 5,
            Interval:    10 * time.Second,
        }),
    ),
}
```

## Configuration Options

| Feature        | Option | Description |
|---------------|--------|-------------|
| **Retry** | `WithMaxRetries(n int)` | Number of retry attempts |
| **Logging** | `WithLogLevel(level slog.Level)` | Set log level (Info, Warn, Error) |
| **Circuit Breaker** | `WithBreakerSettings(settings gobreaker.Settings)` | Custom circuit breaker settings |

