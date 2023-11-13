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
		shutdown, session, err := initSession(config)
		if err != nil {
			return fmt.Errorf("cassandra health check failed on connect: %w", err)
		}

		defer shutdown()

		if err != nil {
			return fmt.Errorf("cassandra health check failed on connect: %w", err)
		}

		err = session.Query("SELECT * FROM system_schema.keyspaces;").WithContext(ctx).Exec()
		if err != nil {
			return fmt.Errorf("cassandra health check failed on describe: %w", err)
		}

		return nil
	}
}

func initSession(c Config) (func(), *gocql.Session, error) {
	if c.Session != nil {
		return func() {}, c.Session, nil
	}

	if len(c.Hosts) < 1 || len(c.Keyspace) < 1 {
		return nil, nil, errors.New("cassandra cluster config or keyspace name and hosts are required to initialize cassandra health check")
	}

	cluster := gocql.NewCluster(c.Hosts...)
	cluster.Keyspace = c.Keyspace
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, nil, err
	}

	return session.Close, session, err
}
