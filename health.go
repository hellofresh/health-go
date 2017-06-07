package health

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

type checkResponse struct {
	name      string
	skipOnErr bool
	err       error
}

// Register registers a check config to be performed.
func Register(c Config) {
	mu.Lock()
	defer mu.Unlock()
	if c.Timeout == 0 {
		c.Timeout = time.Second * 2
	}
	checks = append(checks, c)
}

// Handler returns an HTTP handler (http.HandlerFunc).
func Handler() http.Handler {
	return http.HandlerFunc(HandlerFunc)
}

// HandlerFunc is the HTTP handler function.
func HandlerFunc(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	status := statusOK
	failures := make(map[string]string)
	resChan := make(chan checkResponse, len(checks))
	for _, c := range checks {
		go func(c Config, resChan chan checkResponse) {
			resChan <- checkResponse{c.Name, c.SkipOnErr, c.Check()}
		}(c, resChan)

	loop:
		for {
			select {
			case <-time.After(c.Timeout):
				failures[c.Name] = "Timeout during health check"
				setStatus(&status, c.SkipOnErr)
				break loop
			case res := <-resChan:
				if res.err != nil {
					failures[res.name] = res.err.Error()
					setStatus(&status, res.skipOnErr)
				}
				break loop
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	c := newCheck(status, failures)
	data, err := json.Marshal(c)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	code := http.StatusOK
	if status == statusUnavailable {
		code = http.StatusServiceUnavailable
	}
	w.WriteHeader(code)
	w.Write(data)
}

func newCheck(status string, failures map[string]string) Check {
	return Check{
		Status:    status,
		Timestamp: time.Now(),
		Failures:  failures,
		System:    newSystemMetrics(),
	}
}

func newSystemMetrics() System {
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

func setStatus(status *string, skipOnErr bool) {
	if skipOnErr && *status != statusUnavailable {
		*status = statusPartiallyAvailable
	} else {
		*status = statusUnavailable
	}
}
