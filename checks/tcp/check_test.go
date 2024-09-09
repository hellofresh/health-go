package tcp

import (
	"github.com/stretchr/testify/require"
	"log"
	"net"
	"strconv"
	"testing"
)

var tcp Server

func init() {
	// Start the new server
	tcp, err := NewServer("tcp", ":1123")
	if err != nil {
		log.Println("error starting TCP server")
		return
	}

	// Run the servers in goroutines to stop blocking
	go func() {
		tcp.Run()
	}()
}

func TestNew(t *testing.T) {

	t.Run("service is available", func(t *testing.T) {

		check := Config{
			Host:           "127.0.0.1",
			Port:           1123,
			RequestTimeout: defaultRequestTimeout,
		}
		conn, err := net.DialTimeout("tcp", check.Host+":"+strconv.Itoa(check.Port), check.RequestTimeout)
		require.NoError(t, err)
		defer conn.Close()
	})

	t.Run("service is not available", func(t *testing.T) {
		check := Config{
			Host:           "127.0.0.1",
			Port:           1124,
			RequestTimeout: defaultRequestTimeout,
		}

		conn, err := net.DialTimeout("tcp", check.Host+":"+strconv.Itoa(check.Port), check.RequestTimeout)
		require.Error(t, err)
		defer conn.Close()

	})
}
