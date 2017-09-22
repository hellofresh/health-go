package http

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// Config is the HTTP checker configuration settings container.
type Config struct {
	// URL is the remote service health check URL.
	URL string
	// RequestTimeout is the duration that health check will try to consume published test message.
	// If not set - 5 seconds
	RequestTimeout time.Duration
	// LogFunc is the callback function for errors logging during check.
	// If not set logging is skipped.
	LogFunc func(err error, details string, extra ...interface{})
}

// New creates new HTTP service health check that verifies the following:
// - connection establishing
// - getting response status from defined URL
// - verifying that status code is less than 500
func New(config Config) func() error {
	return func() error {
		if config.LogFunc == nil {
			config.LogFunc = func(err error, details string, extra ...interface{}) {}
		}

		if config.RequestTimeout == 0 {
			config.RequestTimeout = time.Second * 5
		}

		req, err := http.NewRequest(http.MethodGet, config.URL, nil)
		if err != nil {
			config.LogFunc(err, "Creating the request for the health check failed")
			return err
		}

		ctx, cancel := context.WithCancel(context.TODO())

		// Inform remote service to close the connection after the transaction is complete
		req.Header.Set("Connection", "close")
		req = req.WithContext(ctx)

		time.AfterFunc(config.RequestTimeout, func() {
			cancel()
		})

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			config.LogFunc(err, "Making the request for the health check failed")
			return err
		}
		defer res.Body.Close()

		if res.StatusCode >= http.StatusInternalServerError {
			return errors.New("remote service is not available at the moment")
		}

		return nil
	}
}
