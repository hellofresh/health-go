package cassandra

import (
	"context"
	"errors"
	"fmt"
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
func New(config Config) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		if len(config.Hosts) < 1 || len(config.Keyspace) < 1 {
			return errors.New("keyspace name and hosts are required to initialize cassandra health check")
		}

		cluster := gocql.NewCluster(config.Hosts...)
		cluster.Keyspace = config.Keyspace

		session, err := cluster.CreateSession()
		if err != nil {
			return fmt.Errorf("cassandra health check failed on connect: %w", err)
		}

		defer session.Close()

		err = session.Query("DESCRIBE KEYSPACES;").WithContext(ctx).Exec()
		if err != nil {
			return fmt.Errorf("cassandra health check failed on describe: %w", err)
		}

		return nil
	}
}
