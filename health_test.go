package health

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	checkErr = "failed during RabbitMQ health check"
)

func TestRegisterWithNoName(t *testing.T) {
	h := New(false, nil, nil)
	err := h.Register(Config{
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
	h := New(false, nil, nil)

	if len(h.checkMap) != 0 {
		t.Errorf("checks lenght differes from zero: got %d", len(h.checkMap))
	}

	healthCheckName := "health-check"

	conf := Config{
		Name: healthCheckName,
		Check: func() error {
			return nil
		},
	}

	err := h.Register(conf)
	require.NoError(t, err, "the first registration of a health check should not return an error, but got one")

	err = h.Register(conf)
	assert.Error(t, err, "the second registration of a health check config should return an error, but did not")

	err = h.Register(Config{
		Name: healthCheckName,
		Check: func() error {
			return errors.New("health checks registered")
		},
	})
	assert.Error(t, err, "registration with same name, but different details should still return an error, but did not")
}

func TestBulkRegister(t *testing.T) {
	h := New(false, nil, nil)

	if len(h.checkMap) != 0 {
		t.Errorf("checks lenght differes from zero: got %d", len(h.checkMap))
	}

	healthCheckName1 := "health-check1"
	conf1 := Config{
		Name: healthCheckName1,
		Check: func() error {
			return nil
		},
	}

	healthCheckName2 := "health-check12"
	conf2 := Config{
		Name: healthCheckName2,
		Check: func() error {
			return nil
		},
	}

	err := h.BulkRegister(conf1, conf2)
	require.NoError(t, err, "the first registration of a health check should not return an error, but got one")
	assert.Len(t, h.checkMap, 2, "checks length diffres from two")

	err = h.BulkRegister(conf1, conf2)
	assert.Error(t, err, "the second registration of a health check config should return an error, but did not")
	assert.Len(t, h.checkMap, 2, "checks length diffres from two")
}

func TestHealthHandler(t *testing.T) {
	hc := New(false, nil, nil)

	res := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "http://localhost/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	hc.Register(Config{
		Name:      "rabbitmq",
		SkipOnErr: true,
		Check:     func() error { return errors.New(checkErr) },
	})

	hc.Register(Config{
		Name:  "mongodb",
		Check: func() error { return nil },
	})

	hc.Register(Config{
		Name:      "snail-service",
		SkipOnErr: true,
		Timeout:   time.Second * 1,
		Check: func() error {
			time.Sleep(time.Second * 2)
			return nil
		},
	})

	h := http.Handler(hc.Handler())
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
}


func TestHealthExecuteCheck(t *testing.T) {
	hc := New(false, nil, nil)

	hc.Register(Config{
		Name:      "rabbitmq",
		SkipOnErr: true,
		Check:     func() error { return errors.New(checkErr) },
	})

	hc.Register(Config{
		Name:  "mongodb",
		Check: func() error { return nil },
	})

	hc.Register(Config{
		Name:      "snail-service",
		SkipOnErr: true,
		Timeout:   time.Second * 1,
		Check: func() error {
			time.Sleep(time.Second * 2)
			return nil
		},
	})

	c := hc.ExecuteCheck()

	assert.Equal(t, statusPartiallyAvailable, c.Status, "body returned wrong status")

	failure := c.Failures
	assert.NotNil(t, failure, "body returned nil failures field")

	assert.Equal(t, checkErr, failure["rabbitmq"], "body returned wrong status for rabbitmq")
	assert.Equal(t, failureTimeout, failure["snail-service"], "body returned wrong status for snail-service")
}

func TestHealthExecuteCheckWithSysMetrics(t *testing.T) {
	hc := New(true, nil, nil)

	hc.Register(Config{
		Name:  "mongodb",
		Check: func() error { return nil },
	})

	c := hc.ExecuteCheck()
	assert.NotEqual(t, System{}, c.System)
}

func TestHealthExecuteCheckWithoutSysMetrics(t *testing.T) {
	hc := New(false, nil, nil)

	hc.Register(Config{
		Name:  "mongodb",
		Check: func() error { return nil },
	})

	c := hc.ExecuteCheck()
	assert.Equal(t, System{}, c.System)
}

func TestHealthExecuteStandaloneUnhealthy(t *testing.T) {
	// Run the crashing code when FLAG is set
	if os.Getenv("FLAG") == "1" {
		h := New(false, nil ,nil)
		h.Register(Config{
			Name:  "mongodb",
			Check: func() error { return errors.New("Error") },
		})
		h.ExecuteStandalone()
		return
	}
	// Run the test in a subprocess
	cmd := exec.Command(os.Args[0], "-test.run=TestHealthExecuteStandaloneUnhealthy")
	cmd.Env = append(os.Environ(), "FLAG=1")
	err := cmd.Run()
	// Cast the error as *exec.ExitError and compare the result
	e, ok := err.(*exec.ExitError)
	expectedErrorString := "exit status 1"
	assert.Equal(t, true, ok)
	assert.Equal(t, expectedErrorString, e.Error())
}

func TestHealthExecuteStandaloneHealthy(t *testing.T) {
	// Run the crashing code when FLAG is set
	if os.Getenv("FLAG") == "1" {
		h := New(false, nil ,nil)
		h.Register(Config{
			Name:  "mongodb",
			Check: func() error { return nil },
		})
		h.ExecuteStandalone()
		return
	}
	// Run the test in a subprocess
	cmd := exec.Command(os.Args[0], "-test.run=TestHealthExecuteStandaloneHealthy")
	cmd.Env = append(os.Environ(), "FLAG=1")
	err := cmd.Run()
	assert.Nil(t, err)
}