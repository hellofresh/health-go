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

func TestHealthHandler(t *testing.T) {
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
