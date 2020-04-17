package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/hellofresh/health-go/v3"
	healthHttp "github.com/hellofresh/health-go/v3/checks/http"
	healthMySql "github.com/hellofresh/health-go/v3/checks/mysql"
	healthPg "github.com/hellofresh/health-go/v3/checks/postgres"
)

func main() {
	// custom health check example (fail)
	health.Register(health.Config{
		Name:      "some-custom-check-fail",
		Timeout:   time.Second * 5,
		SkipOnErr: true,
		Check:     func() error { return errors.New("failed during custom health check") },
	})

	// custom health check example (success)
	health.Register(health.Config{
		Name:  "some-custom-check-success",
		Check: func() error { return nil },
	})

	// http health check example
	health.Register(health.Config{
		Name:      "http-check",
		Timeout:   time.Second * 5,
		SkipOnErr: true,
		Check: healthHttp.New(healthHttp.Config{
			URL: `http://example.com`,
		}),
	})

	// postgres health check example
	health.Register(health.Config{
		Name:      "postgres-check",
		Timeout:   time.Second * 5,
		SkipOnErr: true,
		Check: healthPg.New(healthPg.Config{
			DSN: `postgres://test:test@0.0.0.0:32783/test?sslmode=disable`,
		}),
	})

	// mysql health check example
	health.Register(health.Config{
		Name:      "mysql-check",
		Timeout:   time.Second * 5,
		SkipOnErr: true,
		Check: healthMySql.New(healthMySql.Config{
			DSN: `test:test@tcp(0.0.0.0:32778)/test?charset=utf8`,
		}),
	})

	// rabbitmq aliveness test example.
	// Use it if your app has access to RabbitMQ management API.
	// This endpoint declares a test queue, then publishes and consumes a message. Intended for use by monitoring tools. If everything is working correctly, will return HTTP status 200.
	// As the default virtual host is called "/", this will need to be encoded as "%2f".
	health.Register(health.Config{
		Name:      "rabbit-aliveness-check",
		Timeout:   time.Second * 5,
		SkipOnErr: true,
		Check: healthHttp.New(healthHttp.Config{
			URL: `http://guest:guest@0.0.0.0:32780/api/aliveness-test/%2f`,
		}),
	})

	http.Handle("/status", health.Handler())
	http.ListenAndServe(":3000", nil)
}
