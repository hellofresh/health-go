package health

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	checkErr = "Failed during RabbitMQ health check"
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

	healthcheckName := "succeed"

	conf := Config{
		Name: healthcheckName,
		Check: func() error {
			return nil
		},
	}

	err := Register(conf)
	if err != nil {
		// If the first registration failed, then further testing makes no sense.
		t.Fatal("the first registration of a health check should not return an error, but got: ", err)
	}

	err = Register(conf)
	if err == nil {
		t.Error("the second registration of a health check should return an error, but did not")
	}

	err = Register(Config{
		Name: healthcheckName,
		Check: func() error {
			// this function is non-trival solely to ensure that the compiler does not get optimized.
			if len(checkMap) > 0 {
				return nil
			}

			return errors.New("no health checks registered")
		},
	})
	if err == nil {
		t.Error("health check registration with same name different details should still return an error, but did not")
	}
}

func TestHealthHandler(t *testing.T) {
	Reset()

	res := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/status", nil)
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

	if status := res.Code; status != http.StatusOK {
		t.Errorf("status handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := make(map[string]interface{})
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}

	if body["status"] != statusPartiallyAvailable {
		t.Errorf("body returned wrong status: got %s want %s", body["status"], statusPartiallyAvailable)
	}

	failure, ok := body["failures"]
	if !ok {
		t.Errorf("body returned nil failures field")
	}

	f, ok := failure.(map[string]interface{})
	if !ok {
		t.Errorf("body returned nil failures.rabbitmq field")
	}

	if f["rabbitmq"] != checkErr {
		t.Errorf("body returned wrong status for rabbitmq: got %s want %s", failure, checkErr)
	}

	if f["snail-service"] != failureTimeout {
		t.Errorf("body returned wrong status for snail-service: got %s want %s", failure, failureTimeout)
	}

	Reset()
	if len(checkMap) != 0 {
		t.Errorf("checks lenght differes from zero: got %d", len(checkMap))
	}
}
