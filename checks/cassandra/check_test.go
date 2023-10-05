package cassandra

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/require"
)

const HOST = "HEALTH_GO_CASSANDRA_HOST"
const KEYSPACE = "default"

func TestNew(t *testing.T) {
	initDB(t)

	check := New(Config{
		Hosts:    getHosts(t),
		Keyspace: KEYSPACE,
	})

	err := check(context.Background())
	require.NoError(t, err)
}

func TestNew_withClusterConfig(t *testing.T) {
	initDB(t)
	check := New(Config{
		ClusterConfig: gocql.NewCluster(getHosts(t)...),
	})
	err := check(context.Background())
	require.NoError(t, err)
}

func TestNewWithError(t *testing.T) {
	check := New(Config{})

	err := check(context.Background())
	require.Error(t, err)
}

func initDB(t *testing.T) {
	t.Helper()

	cluster := gocql.NewCluster(getHosts(t)[0])

	session, err := cluster.CreateSession()
	require.NoError(t, err)

	defer session.Close()

	err = session.Query(fmt.Sprintf("CREATE KEYSPACE IF NOT EXISTS %s WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};", KEYSPACE)).Exec()
	require.NoError(t, err)

	err = session.Query(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.test (id UUID PRIMARY KEY, name text);", KEYSPACE)).Exec()
	require.NoError(t, err)
}

func getHosts(t *testing.T) []string {
	t.Helper()

	host, ok := os.LookupEnv(HOST)
	require.True(t, ok, fmt.Sprintf("Host is: %s", host))

	// "docker-compose port <service> <port>" returns 0.0.0.0:XXXX locally, change it to local port
	host = strings.Replace(host, "0.0.0.0:", "127.0.0.1:", 1)

	return []string{host}
}
