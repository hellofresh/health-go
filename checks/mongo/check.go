package mongo

import "github.com/globalsign/mgo"

// Config is the MongoDB checker configuration settings container.
type Config struct {
	// DSN is the MongoDB instance connection DSN. Required.
	DSN string
	// LogFunc is the callback function for errors logging during check.
	// If not set logging is skipped.
	LogFunc func(err error, details string, extra ...interface{})
}

// New creates new MongoDB health check that verifies the following:
// - connection establishing
// - doing the ping command
func New(config Config) func() error {
	if config.LogFunc == nil {
		config.LogFunc = func(err error, details string, extra ...interface{}) {}
	}

	return func() error {
		session, err := mgo.Dial(config.DSN)
		if err != nil {
			config.LogFunc(err, "MongoDB health check failed during connect")
			return err
		}
		defer session.Close()

		err = session.Ping()
		if err != nil {
			config.LogFunc(err, "MongoDB health check failed during ping")
			return err
		}

		return nil
	}
}
