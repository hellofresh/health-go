# health-go

* Exposes an HTTP handler that retrieves health status of the application.

### Usage

```go
package main

import (
  "net/http"
  "time"

  health "github.com/hellofresh/health-go"
)

func main() {
  health.Register(health.Config{
    Name: "rabbitmq",
    Timeout: time.Second*5,
    SkipOnErr: true,
    Check: func() error {
      // rabbitmq health check implementation goes here
    },
  })

  health.Register(health.Config{
    Name: "mongodb",
    Check: func() error {
      // mongo_db health check implementation goes here
    },
  })

  http.Handle("/status", health.Handler())
  http.ListenAndServe(":3000", nil)
}
```

### API Documentation
#### `GET /status`
Get the health of the application.
- Method: `GET`
- Endpoint: `/status`
- Request:
```
curl localhost:3000/status
```
- Response:
HTTP/1.1 200 OK
```json
{
  "status": "OK",
  "timestamp": "2017-01-01T00:00:00.413567856+033:00",
  "system": {
    "version": "go1.8",
    "goroutines_count": 4,
    "total_alloc_bytes": 21321,
    "heap_objects_count": 21323,
    "alloc_bytes": 234523
  }
}
```

HTTP/1.1 200 OK
```json
{
  "status": "Partially Available",
  "timestamp": "2017-01-01T00:00:00.413567856+033:00",
  "failures": {
    "rabbitmq": "Failed duriung rabbitmq health check"
  },
  "system": {
    "version": "go1.8",
    "goroutines_count": 4,
    "total_alloc_bytes": 21321,
    "heap_objects_count": 21323,
    "alloc_bytes": 234523
  }
}
```

HTTP/1.1 503 Service Unavailable
```json
{
  "status": "Unavailable",
  "timestamp": "2017-01-01T00:00:00.413567856+033:00",
  "failures": {
    "mongodb": "Failed duriung mongodb health check"
  },
  "system": {
    "version": "go1.8",
    "goroutines_count": 4,
    "total_alloc_bytes": 21321,
    "heap_objects_count": 21323,
    "alloc_bytes": 234523
  }
}
```

## Contributing
- Fork it
- Create your feature branch (`git checkout -b my-new-feature`)
- Commit your changes (`git commit -am 'Add some feature'`)
- Push to the branch (`git push origin my-new-feature`)
- Create new Pull Request

## Badges

[![Build Status](https://travis-ci.org/hellofresh/health-go.svg?branch=master)](https://travis-ci.org/hellofresh/health-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/hellofresh/health-go)](https://goreportcard.com/report/github.com/hellofresh/health-go)
[![Go Doc](https://godoc.org/github.com/hellofresh/health-go?status.svg)](https://godoc.org/github.com/hellofresh/health-go)

---

> GitHub [@hellofresh](https://github.com/hellofresh) &nbsp;&middot;&nbsp;
> Medium [@engineering.hellofresh](https://engineering.hellofresh.com)
