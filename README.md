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
    Name: "rabbit_mq",
    Timeout: time.Second*5,
    SkipOnErr: true,
    CheckFunc: func() error {
      // rabbitmq health check implementation goes here
    })

  health.Register(health.Config{
    Name: "mongodb",
    Timeout: time.Second*5,
    CheckFunc: func() error {
      // mongo_db health check implementation goes here
  })

  http.Handle("/status", health.Handler())
	http.ListenAndServe(":3000", nil)
}
```

```
$ http GET localhost:8000/status
HTTP/1.1 503 Service Unavailable
{
  "status": "Unavailable",
  "timestamp": "2017-01-01T00:00:00.413567856+033:00",
  "failures": {
    "mongodb": "Error"
  },
  "system": {
    "alloc_bytes": 234523,
    "goroutines_count": 4,
    "heap_objects_count": 21323,
    "total_alloc_bytes": 21321,
    "version": "go1.8"
  }
}
```
