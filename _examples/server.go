package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/hellofresh/health-go"
	healthHttp "github.com/hellofresh/health-go/checks/http"
	healthMySql "github.com/hellofresh/health-go/checks/mysql"
	healthPg "github.com/hellofresh/health-go/checks/postgres"
)

func main() {
	// custom health check example
	health.Register(health.Config{
		Name:      "rabbitmq",
		Timeout:   time.Second * 5,
		SkipOnErr: true,
		Check:     func() error { return errors.New("failed during rabbitmq health check") },
	})

	// custom health check example
	health.Register(health.Config{
		Name:  "mongodb",
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
			DSN: `postgres://test:test@postgres:5432/test?sslmode=disable`,
		}),
	})

	// mysql health check example
	health.Register(health.Config{
		Name:      "mysql-check",
		Timeout:   time.Second * 5,
		SkipOnErr: true,
		Check: healthMySql.New(healthMySql.Config{
			DSN: `test:test@tcp(mysql:3306)/test?charset=utf8`,
		}),
	})

	http.Handle("/status", health.Handler())
	http.ListenAndServe(":3000", nil)
}
