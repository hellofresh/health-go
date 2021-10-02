package neo4j

import (
	"context"
	"time"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// Config is the Neo4j checker configuration settings container.
type Config struct {
	// DSN is the Neo4j instance connection DSN, supports bolt:// and neo4j:// formats. Required.
	DSN string

	// Username is the username credential to connect to the Neo4j instance. Required.
	Username string

	// Password is the password credential to connect to the Neo4j instance. Required.
	Password string

	// ConnectionAcquisitionTimeout is the maximum amount of time to  create a new connection
	// Negative values result in an infinite wait time where 0 value results in no timeout which
	// results in immediate failure when there are no available connections.
	ConnectionAcquisitionTimeout time.Duration

	// SocketConnectTimeout timeout that will be set on underlying sockets.
	// Values less than or equal to 0 results in no timeout being applied.
	SocketConnectTimeout time.Duration
}

// New creates new Neo4j health check that verifies the connection with the instance
func New(config Config) func(ctx context.Context) error {
	return func(_ context.Context) error {
		cfgFn := func(c *neo4j.Config) {
			c.ConnectionAcquisitionTimeout = config.ConnectionAcquisitionTimeout
			c.SocketConnectTimeout = config.SocketConnectTimeout
		}

		driver, err := neo4j.NewDriver(config.DSN, neo4j.BasicAuth(config.Username, config.Password, ""), cfgFn)
		if err != nil {
			panic(err)
		}
		defer driver.Close()

		return driver.VerifyConnectivity()
	}
}
