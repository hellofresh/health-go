package healthgo

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

var checks []Config

const (
	statusOK                 = "OK"
	statusPartiallyAvailable = "Partially Available"
	statusUnavailable        = "Unavailable"
)

type Config struct {
	Name      string
	Timeout   time.Duration
	SkipOnErr bool
	Check     func() error
}

type Check struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Failures  map[string]string `json:"failures,omitempty"`
	System    `json:"system"`
}

type System struct {
	Version          string `json:"version"`
	GoroutinesCount  int    `json:"goroutines_count"`
	TotalAllocBytes  int    `json:"total_alloc_bytes"`
	HeapObjectsCount int    `json:"heap_objects_count"`
	AllocBytes       int    `json:"alloc_bytes"`
}

func Register(c Config) {
	checks = append(checks, c)
}

func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
