package healthgo

import (
	"encoding/json"
	"net/http"
	"runtime"
	"sync"
	"time"
)

var mu sync.Mutex
var checks []Config

const (
	statusOK                 = "OK"
	statusPartiallyAvailable = "Partially Available"
	statusUnavailable        = "Unavailable"
)

// Config carries the parameters to run the check.
type Config struct {
	Name      string
	Timeout   time.Duration
	SkipOnErr bool
	Check     func() error
}

// Check represents the health check response.
type Check struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Failures  map[string]string `json:"failures,omitempty"`
	System    `json:"system"`
}

// System runtime variables about the go process.
type System struct {
	Version          string `json:"version"`
	GoroutinesCount  int    `json:"goroutines_count"`
	TotalAllocBytes  int    `json:"total_alloc_bytes"`
	HeapObjectsCount int    `json:"heap_objects_count"`
	AllocBytes       int    `json:"alloc_bytes"`
}

// Register registers a check to be evaluated each given period.
func Register(c Config) {
	mu.Lock()
	defer mu.Unlock()
	checks = append(checks, c)
}

// Handler returns an HTTP handler (http.HandlerFunc)
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		status := statusOK
		now := time.Now()
		failures := make(map[string]string)

		for _, c := range checks {
			err := c.Check()
			if err != nil {
				failures[c.Name] = err.Error()
				if c.SkipOnErr && status != statusUnavailable {
					status = statusPartiallyAvailable
				} else {
					status = statusUnavailable
				}
			}
		}

		c := Check{
			Status:    status,
			Timestamp: now,
			Failures:  failures,
			System:    systemMetrics(),
		}

		data, err := json.Marshal(c)
		if err != nil {
			return
		}

		w.Header().Set("Content-Type", "application/json")
		code := http.StatusOK
		if status == statusUnavailable {
			code = http.StatusServiceUnavailable
		}
		w.WriteHeader(code)
		w.Write(data)
	})
}

func systemMetrics() System {
	s := runtime.MemStats{}
	runtime.ReadMemStats(&s)
	return System{
		Version:          runtime.Version(),
		GoroutinesCount:  runtime.NumGoroutine(),
		TotalAllocBytes:  int(s.TotalAlloc),
		HeapObjectsCount: int(s.HeapObjects),
		AllocBytes:       int(s.Alloc),
	}
}
