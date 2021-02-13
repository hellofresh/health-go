package memcached

import (
	"context"
	"fmt"
	"strings"

	"github.com/bradfitz/gomemcache/memcache"
)

// Config is the Memcached checker configuration settings container.
type Config struct {
	// DSN is the Memcached instance connection DSN. Required.
	DSN string
}

// New creates new Memcached health check that verifies the following:
// - connection establishing
// - doing the PING command and verifying the response
func New(config Config) func(ctx context.Context) error {
	// support all DSN formats (for backward compatibility) - with and w/out schema and path part:
	// - memcached://localhost:11211/
	// - localhost:11211
	memcachedDSN := strings.TrimPrefix(config.DSN, "memcached://")
	memcachedDSN = strings.TrimSuffix(memcachedDSN, "/")

	return func(ctx context.Context) error {
		mdb := memcache.New(memcachedDSN)

		err := mdb.Ping()

		if err != nil {
			return fmt.Errorf("memcached ping failed: %w", err)
		}

		return nil
	}
}
