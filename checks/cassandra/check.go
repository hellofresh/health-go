package cassandra

import (
	"context"
	"errors"
	"fmt"

	"github.com/gocql/gocql"
)

// Config is the Cassandra checker configuration settings container.
type Config struct {
	// Hosts is a list of Cassandra hosts. Optional if Session is supplied.
	Hosts []string
	// Keyspace is the Cassandra keyspace to which you want to connect. Optional if Session is supplied.
	Keyspace string
	// Session is a gocql session and can be used in place of Hosts and Keyspace. Recommended.
	// Optional if Hosts & Keyspace are supplied.
	Session *gocql.Session
}

// New creates new Cassandra health check that verifies that a connection exists and can be used to query the cluster.
func New(config Config) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		var session *gocql.Session
		var err error
		if config.Session == nil && (len(config.Hosts) < 1 || len(config.Keyspace) < 1) {
			return errors.New("cassandra cluster config or keyspace name and hosts are required to initialize cassandra health check")
		}

		if config.Session != nil {
			session = config.Session
		} else {
			cluster := gocql.NewCluster(config.Hosts...)
			cluster.Keyspace = config.Keyspace
			session, err = cluster.CreateSession()
			defer session.Close()
		}

		if err != nil {
			return fmt.Errorf("cassandra health check failed on connect: %w", err)
		}

		err = session.Query("DESCRIBE KEYSPACES;").WithContext(ctx).Exec()
		if err != nil {
			return fmt.Errorf("cassandra health check failed on describe: %w", err)
		}

		return nil
	}
}
