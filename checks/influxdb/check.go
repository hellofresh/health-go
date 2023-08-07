// Package influxdb implements a health check for InfluxDB instance.
package influxdb

import (
	"context"
	"errors"
	"fmt"
	"github.com/hellofresh/health-go/v5"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

// Config stores InfluxDB API host and possibly parameters.
type Config struct {
	URL string
}

// New returns a check function. It uses InfluxDB health api
// to get status of the instance.
func New(config Config) func(ctx context.Context) health.CheckResponse {
	return func(ctx context.Context) (checkResponse health.CheckResponse) {
		// since only health api will be used, we don't need to pass
		// any Authorization data (token in this case)
		client := influxdb2.NewClient(config.URL, "")
		defer client.Close()

		h, err := client.Health(ctx)

		if err != nil {
			checkResponse.Error = fmt.Errorf("InfluxDB health check failed: %w", err)
			return
		}

		// any status different from "pass" is considered as failed
		if h.Status != domain.HealthCheckStatusPass {
			checkResponse.Error = errors.New("InfluxDB health check failed, didn't get PASS status")
			return
		}

		return
	}
}
