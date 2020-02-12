package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hellofresh/health-go"
	healthHttp "github.com/hellofresh/health-go/checks/http"
	healthMySql "github.com/hellofresh/health-go/checks/mysql"
	healthPg "github.com/hellofresh/health-go/checks/postgres"
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

	// custom health check example (fail)
	healthCheck.Register(health.Config{
		Name:      "some-custom-check-fail",
		Timeout:   time.Second * 5,
		SkipOnErr: true,
		Check:     func() error { return errors.New("failed during custom health check") },
	})

	// custom health check example (success)
	healthCheck.Register(health.Config{
		Name:  "some-custom-check-success",
		Check: func() error { return nil },
	})

	// http health check example
	healthCheck.Register(health.Config{
		Name:      "http-check",
		Timeout:   time.Second * 5,
		SkipOnErr: true,
		Check: healthHttp.New(healthHttp.Config{
			URL: `http://example.com`,
		}),
	})

	// postgres health check example
	healthCheck.Register(health.Config{
		Name:      "postgres-check",
		Timeout:   time.Second * 5,
		SkipOnErr: true,
		Check: healthPg.New(healthPg.Config{
			DSN: `postgres://test:test@0.0.0.0:32783/test?sslmode=disable`,
		}),
	})

	// mysql health check example
	mysql := health.Config{
		Name:      "mysql-check",
		Timeout:   time.Second * 5,
		SkipOnErr: true,
		Check: healthMySql.New(healthMySql.Config{
			DSN: `test:test@tcp(0.0.0.0:32778)/test?charset=utf8`,
		}),
	}

	// rabbitmq aliveness test example.
	// Use it if your app has access to RabbitMQ management API.
	// This endpoint declares a test queue, then publishes and consumes a message. Intended for use by monitoring tools. If everything is working correctly, will return HTTP status 200.
	// As the default virtual host is called "/", this will need to be encoded as "%2f".
	rabbit := health.Config{
		Name:      "rabbit-aliveness-check",
		Timeout:   time.Second * 5,
		SkipOnErr: true,
		Check: healthHttp.New(healthHttp.Config{
			URL: `http://guest:guest@0.0.0.0:32780/api/aliveness-test/%2f`,
		}),
	}

	healthCheck.BulkRegister(mysql, rabbit)

 	//if the flag hc is true, health check will be executed and exit te program
	healthCheck.HealthCheckStandaloneMode("hc")

	http.Handle("/status", healthCheck.Handler())
	http.ListenAndServe(":3000", nil)
}
