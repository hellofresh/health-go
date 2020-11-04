package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	checkErr = "failed during RabbitMQ health check"
)

func TestRegisterWithNoName(t *testing.T) {
	h, err := New()
	require.NoError(t, err)

	err = h.Register(Config{
		Name: "",
		Check: func(context.Context) error {
			return nil
		},
	})
	require.Error(t, err, "health check registration with empty name should return an error")
}

func TestDoubleRegister(t *testing.T) {
	h, err := New()
	require.NoError(t, err)

	healthCheckName := "health-check"

	conf := Config{
		Name: healthCheckName,
		Check: func(context.Context) error {
			return nil
		},
	}

	err = h.Register(conf)
	require.NoError(t, err, "the first registration of a health check should not return an error, but got one")

	err = h.Register(conf)
	assert.Error(t, err, "the second registration of a health check config should return an error, but did not")

	err = h.Register(Config{
		Name: healthCheckName,
		Check: func(context.Context) error {
			return errors.New("health checks registered")
		},
	})
	assert.Error(t, err, "registration with same name, but different details should still return an error, but did not")
}

func TestHealthHandler(t *testing.T) {
	h, err := New()
	require.NoError(t, err)

	res := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://localhost/status", nil)
	require.NoError(t, err)

	err = h.Register(Config{
		Name:      "rabbitmq",
		SkipOnErr: true,
		Check:     func(context.Context) error { return errors.New(checkErr) },
	})
	require.NoError(t, err)

	err = h.Register(Config{
		Name:  "mongodb",
		Check: func(context.Context) error { return nil },
	})
	require.NoError(t, err)

	err = h.Register(Config{
		Name:      "snail-service",
		SkipOnErr: true,
		Timeout:   time.Second * 1,
		Check: func(context.Context) error {
			time.Sleep(time.Second * 2)
			return nil
		},
	})
	require.NoError(t, err)

	handler := h.Handler()
	handler.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code, "status handler returned wrong status code")

	body := make(map[string]interface{})
	err = json.NewDecoder(res.Body).Decode(&body)
	require.NoError(t, err)

	assert.Equal(t, string(StatusPartiallyAvailable), body["status"], "body returned wrong status")

	failure, ok := body["failures"]
	assert.True(t, ok, "body returned nil failures field")

	f, ok := failure.(map[string]interface{})
	assert.True(t, ok, "body returned nil failures.rabbitmq field")

	assert.Equal(t, checkErr, f["rabbitmq"], "body returned wrong status for rabbitmq")
	assert.Equal(t, string(StatusTimeout), f["snail-service"], "body returned wrong status for snail-service")
}

func TestHealth_Measure(t *testing.T) {
	t.Logf("Num CPUs: %d", runtime.NumCPU())

	h, err := New(WithChecks(Config{
		Name:      "check1",
		Timeout:   time.Second,
		SkipOnErr: false,
		Check: func(context.Context) error {
			time.Sleep(time.Second * 10)
			return errors.New("check1")
		},
	}, Config{
		Name:      "check2",
		Timeout:   time.Second * 2,
		SkipOnErr: false,
		Check: func(context.Context) error {
			time.Sleep(time.Second * 10)
			return errors.New("check2")
		},
	}))
	require.NoError(t, err)

	startedAt := time.Now()
	result := h.Measure(context.Background())
	elapsed := time.Since(startedAt)

	// both checks should run concurrently and should fail with timeout,
	// so should take not less than 2 sec, but less than 5 that is check time
	require.GreaterOrEqual(t, elapsed.Milliseconds(), (time.Second * 2).Milliseconds())
	require.Less(t, elapsed.Milliseconds(), (time.Second * 10).Milliseconds())

	assert.Equal(t, StatusUnavailable, result.Status)
	assert.Equal(t, string(StatusTimeout), result.Failures["check1"])
	assert.Equal(t, string(StatusTimeout), result.Failures["check2"])
}
