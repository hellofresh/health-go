package main

import (
	"errors"
	"net/http"
	"time"

	health "github.com/hellofresh/health-go"
)

func main() {
	health.Register(health.Config{
		Name:      "rabbitmq",
		Timeout:   time.Second * 5,
		SkipOnErr: true,
		Check:     func() error { return errors.New("Failed duriung rabbitmq health check") },
	})

	health.Register(health.Config{
		Name:  "mongodb",
		Check: func() error { return nil },
	})

	http.Handle("/status", health.Handler())
	http.ListenAndServe(":3000", nil)
}
