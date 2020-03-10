package health

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	checkErr = "failed during RabbitMQ health check"
)

func TestRegisterWithNoName(t *testing.T) {
	err := Register(Config{
		Name: "",
		Check: func() error {
			return nil
		},
	})
	if err == nil {
		t.Error("health check registration with empty name should return an error, but did not get one")
	}
}

func TestDoubleRegister(t *testing.T) {
	Reset()
	if len(checkMap) != 0 {
		t.Errorf("checks lenght differes from zero: got %d", len(checkMap))
	}

	healthCheckName := "health-check"

	conf := Config{
		Name: healthCheckName,
		Check: func() error {
			return nil
		},
	}

	err := Register(conf)
	require.NoError(t, err, "the first registration of a health check should not return an error, but got one")

	err = Register(conf)
	assert.Error(t, err, "the second registration of a health check config should return an error, but did not")

	err = Register(Config{
		Name: healthCheckName,
		Check: func() error {
			return errors.New("health checks registered")
		},
	})
	assert.Error(t, err, "registration with same name, but different details should still return an error, but did not")
}

func TestHealthHandler(t *testing.T) {
	Reset()

	res := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://localhost/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	Register(Config{
		Name:      "rabbitmq",
		SkipOnErr: true,
		Check:     func() error { return errors.New(checkErr) },
	})

	Register(Config{
		Name:  "mongodb",
		Check: func() error { return nil },
	})

	Register(Config{
		Name:      "snail-service",
		SkipOnErr: true,
		Timeout:   time.Second * 1,
		Check: func() error {
			time.Sleep(time.Second * 2)
			return nil
		},
	})

	h := http.Handler(Handler())
	h.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code, "status handler returned wrong status code")

	body := make(map[string]interface{})
	err = json.NewDecoder(res.Body).Decode(&body)
	require.NoError(t, err)

	assert.Equal(t, statusPartiallyAvailable, body["status"], "body returned wrong status")

	failure, ok := body["failures"]
	assert.True(t, ok, "body returned nil failures field")

	f, ok := failure.(map[string]interface{})
	assert.True(t, ok, "body returned nil failures.rabbitmq field")

	assert.Equal(t, checkErr, f["rabbitmq"], "body returned wrong status for rabbitmq")
	assert.Equal(t, failureTimeout, f["snail-service"], "body returned wrong status for snail-service")

	Reset()
	assert.Len(t, checkMap, 0, "checks length diffres from zero")
}
