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
