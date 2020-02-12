# health-go
[![Build Status](https://travis-ci.com/sensedia/health-go.svg?branch=master)](https://travis-ci.com/sensedia/health-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/sensedia/health-go)](https://goreportcard.com/report/github.com/sensedia/health-go)
[![Go Doc](https://godoc.org/github.com/sensedia/health-go?status.svg)](https://godoc.org/github.com/sensedia/health-go)

* Standalone Health Check
* Customizable Health Check entry
* Exposes an HTTP handler that retrieves health status of the application
* Implements some generic checkers for the following services:
  * RabbitMQ
  * PostgreSQL
  * Redis
  * HTTP
  * MongoDB
  * MySQL

## Usage
The library exports `Execute` function to implements a custom health check entry; `HealthStandaloneMode` function execute the health check and exit the program with right status;
 `Handler` and `HandlerFunc` functions which are fully compatible with `net/http`.

### Custom Health Check entry

```go
import(
"fmt"
"github.com/sensedia/health-go"

)
func main(){
	//Custom func to handle a internal errors and log
	ErrorLogFunc := func(err error, details string, extra ...interface{}) {
		fmt.Println("Errors:\n", err, "\nDetails:\n", details)
	}
	//Custom func to print debug logs
	DebugLogFunc := func(args ...interface{}) {
		fmt.Println(args)
	}
	//Both func can be nil and is optional

	h := health.New(true, ErrorLogFunc, DebugLogFunc)

	// custom health check example (success)
	h.Register(health.Config{
		Name:  "some-custom-check-success",
		Check: func() error { return nil },
	})

	c := customHealthCheckEntry(&h)
	fmt.Println(c)
}

func customHealthCheckEntry(healthCheck *health.Health) health.Check{
	//personalized logic
	return healthCheck.ExecuteCheck()
}
```

### Standalone
This category of health check can be used when the service uses the Cloud Events Specification, where it is not allowed to create endpoints or create another server in the application. The standalone mode is run by Kubernetes' Liveness Probe, running the application itself with the health check flag and verifying the integrity of the service's external dependencies.
```go
package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/sensedia/health-go"
)

func main() {
	//Custom func to handle a internal errors and log
	ErrorLogFunc := func(err error, details string, extra ...interface{}) {
		fmt.Println("Errors:\n", err, "\nDetails:\n", details)
	}
	//Custom func to print debug logs
	DebugLogFunc := func(args ...interface{}) {
		fmt.Println(args)
	}
	//Both func can be nil

	healthCheck := health.New(true, ErrorLogFunc, DebugLogFunc)

	// custom health check example (success)
	healthCheck.Register(health.Config{
		Name:  "some-custom-check-success",
		Check: func() error { return nil },
	})

 	//if the flag hc is true, health check will be executed and exit te program
	healthCheck.HealthCheckStandaloneMode("hc")

	http.Handle("/status", healthCheck.Handler())
	http.ListenAndServe(":3000", nil)
}

```
### Handler

```go
package main

import (
  "net/http"
  "time"

  "github.com/sensedia/health-go"
  healthMysql "github.com/sensedia/health-go/checks/mysql"
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
  
  health.Register(health.Config{
    Name:      "mysql",
    Timeout:   time.Second * 2,
    SkipOnErr: false,
    Check: healthMysql.New(healthMysql.Config{
      DSN: "test:test@tcp(0.0.0.0:31726)/test?charset=utf8",
    },
  })

  http.Handle("/status", health.Handler())
  http.ListenAndServe(":3000", nil)
}
```

### HandlerFunc
```go
package main

import (
  "net/http"
  "time"

  "github.com/go-chi/chi"
  "github.com/sensedia/health-go"
  healthMysql "github.com/sensedia/health-go/checks/mysql"
)

func main() {
  health.Register(health.Config{
    Name: "rabbitmq",
    Timeout: time.Second*5,
    SkipOnErr: true,
    Check: func() error {
      // rabbitmq health check implementation goes here
    }),
  })

  health.Register(health.Config{
    Name: "mongodb",
    Check: func() error {
      // mongo_db health check implementation goes here
    },
  })
  
  health.Register(health.Config{
    Name:      "mysql",
    Timeout:   time.Second * 2,
    SkipOnErr: false,
    Check: healthMysql.New(healthMysql.Config{
      DSN:               "test:test@tcp(0.0.0.0:31726)/test?charset=utf8",
    },
  })

  r := chi.NewRouter()
  r.Get("/status", health.HandlerFunc)
  http.ListenAndServe(":3000", nil)
}
```

For more examples please check [here](https://github.com/sensedia/health-go/blob/master/_examples/server.go)
## API Documentation

### `GET /status`

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
    "rabbitmq": "Failed during rabbitmq health check"
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
    "mongodb": "Failed during mongodb health check"
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

### Note
This project is a fork of hellofresh health-go, the original content is available in GitHub [@hellofresh](https://github.com/hellofresh) 

---
> GitHub [@sensedia](https://github.com/sensedia) &nbsp;&middot;&nbsp;

