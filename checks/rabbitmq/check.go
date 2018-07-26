package rabbitmq

import (
	"time"

	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

const (
	defaultExchange = "health_check"
)

var (
	defaultConsumeTimeout = time.Second * 3
)

type (
	// Config is the RabbitMQ checker configuration settings container.
	Config struct {
		// DSN is the RabbitMQ instance connection DSN. Required.
		DSN string
		// Vhost is the RabbitMQ virtual host. As the default virtual host is called "/", this will need to be encoded as "%2f".
		Vhost string
		// LogFunc is the callback function for errors logging during check.
		// If not set logging is skipped.
		LogFunc func(err error, details string, extra ...interface{})
	}
)

// New creates new RabbitMQ health check that verifies the following:
// - calls aliveness-test endpoint in rabbitmq
func New(config Config) func() error {
	if config.LogFunc == nil {
		config.LogFunc = func(err error, details string, extra ...interface{}) {}
	}

	if config.Vhost == `` {
		config.Vhost = `%2f`
	}

	return func() error {
		req, err := http.NewRequest(http.MethodGet, config.DSN+`api/aliveness-test/`+config.Vhost, nil)
		if err != nil {
			config.LogFunc(err, "Creating the request for the rabbitmq health check failed")
			return err
		}

		ctx, cancel := context.WithCancel(context.TODO())

		// Inform remote service to close the connection after the transaction is complete
		req.Header.Set("Connection", "close")
		req = req.WithContext(ctx)

		time.AfterFunc(defaultConsumeTimeout, func() {
			cancel()
		})

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			config.LogFunc(err, "Making the request for the  rabbitmq health check failed")
			return err
		}
		defer res.Body.Close()

		if res.StatusCode >= http.StatusInternalServerError {
			return errors.New("rabbitmq health check is not available at the moment")
		}

		body, err := ioutil.ReadAll(res.Body)

		data := map[string]interface{}{}
		json.Unmarshal(body, &data)

		if data["status"] != "ok" {
			return errors.New("rabbitmq health check is not available at the moment")
		}

		return nil
	}
}
