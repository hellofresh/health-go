// Package influxdb implements a health check for InfluxDB instance.
package influxdb

import (
	"context"
	"errors"
	"fmt"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

type Config struct {
	// URL is an InfluxDB API host
	URL string
}

// New returns a check function. It uses InfluxDB health api
// to get status of the instance.
func New(config Config) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		// since only health api will be used, we don't need to pass
		// any Authorization data (token in this case)
		client := influxdb2.NewClient(config.URL, "")
		defer client.Close()

		h, err := client.Health(ctx)

		if err != nil {
			return fmt.Errorf("InfluxDB health check failed: %w", err)
		}

		// any status different from "pass" is considered as failed
		if h.Status != domain.HealthCheckStatusPass {
			return errors.New("InfluxDB health check failed, didn't get PASS status")
		}

		return nil
	}
}
