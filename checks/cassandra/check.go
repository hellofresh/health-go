package cassandra

import (
	"context"
	"errors"
	"fmt"
	"github.com/hellofresh/health-go/v5"

	"github.com/gocql/gocql"
)

// Config is the Cassandra checker configuration settings container.
type Config struct {
	// Hosts is a list of Cassandra hosts. At least one is required.
	Hosts []string
	// Keyspace is the Cassandra keyspace to which you want to connect. Required.
	Keyspace string
}

// New creates new Cassandra health check that verifies the following:
// - that a connection can be established through creating a session
// - that queries can be executed by describing keyspaces
func New(config Config) func(ctx context.Context) health.CheckResponse {
	return func(ctx context.Context) (checkResponse health.CheckResponse) {
		if len(config.Hosts) < 1 || len(config.Keyspace) < 1 {
			checkResponse.Error = errors.New("keyspace name and hosts are required to initialize cassandra health check")
			return
		}

		cluster := gocql.NewCluster(config.Hosts...)
		cluster.Keyspace = config.Keyspace

		session, err := cluster.CreateSession()
		if err != nil {
			checkResponse.Error = fmt.Errorf("cassandra health check failed on connect: %w", err)
			return
		}

		defer session.Close()

		err = session.Query("DESCRIBE KEYSPACES;").WithContext(ctx).Exec()
		if err != nil {
			checkResponse.Error = fmt.Errorf("cassandra health check failed on describe: %w", err)
			return
		}

		return
	}
}
