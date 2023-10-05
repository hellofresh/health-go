package cassandra

import (
	"context"
	"errors"
	"fmt"

	"github.com/gocql/gocql"
)

// Config is the Cassandra checker configuration settings container.
type Config struct {
	// Hosts is a list of Cassandra hosts. Optional if ClusterConfig is supplied.
	Hosts []string
	// Keyspace is the Cassandra keyspace to which you want to connect. Optional if ClusterConfig is supplied.
	Keyspace string
	// ClusterConfig is a struct to configure the default cluster implementation of gocql. Can  be used in place of
	// Hosts and Keyspace for more customized behavior.
	// Optional if Hosts & Keyspace are supplied.
	ClusterConfig *gocql.ClusterConfig
}

// New creates new Cassandra health check that verifies the following:
// - that a connection can be established through creating a session
// - that queries can be executed by describing keyspaces
func New(config Config) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		var session *gocql.Session
		var err error
		if config.ClusterConfig == nil && (len(config.Hosts) < 1 || len(config.Keyspace) < 1) {
			return errors.New("cassandra cluster config or keyspace name and hosts are required to initialize cassandra health check")
		}

		if config.ClusterConfig != nil {
			session, err = config.ClusterConfig.CreateSession()
		} else {
			cluster := gocql.NewCluster(config.Hosts...)
			cluster.Keyspace = config.Keyspace
			session, err = cluster.CreateSession()
		}

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
