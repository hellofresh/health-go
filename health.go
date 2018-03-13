package health

import (
	"encoding/json"
	"net/http"
	"runtime"
	"sync"
	"time"
)

var (
	mu     sync.Mutex
	checks []Config
)

const (
	statusOK                 = "OK"
	statusPartiallyAvailable = "Partially Available"
	statusUnavailable        = "Unavailable"
	failureTimeout           = "Timeout during health check"
)

type (
	// CheckFunc is the func which executes the check.
	CheckFunc func() error

	// Config carries the parameters to run the check.
	Config struct {
		// Name is the name of the resource to be checked.
		Name string
		// Timeout is the timeout defined for every check.
		Timeout time.Duration
		// SkipOnErr if set to true, it will retrieve StatusOK providing the error message from the failed resource.
		SkipOnErr bool
		// Check is the func which executes the check.
		Check CheckFunc
	}

	// Check represents the health check response.
	Check struct {
		// Status is the check status.
		Status string `json:"status"`
		// Timestamp is the time in which the check occurred.
		Timestamp time.Time `json:"timestamp"`
		// Failures holds the failed checks along with their messages.
		Failures map[string]string `json:"failures,omitempty"`
		// System holds information of the go process.
		System `json:"system"`
	}

	// System runtime variables about the go process.
	System struct {
		// Version is the go version.
		Version string `json:"version"`
		// GoroutinesCount is the number of the current goroutines.
		GoroutinesCount int `json:"goroutines_count"`
		// TotalAllocBytes is the total bytes allocated.
		TotalAllocBytes int `json:"total_alloc_bytes"`
		// HeapObjectsCount is the number of objects in the go heap.
		HeapObjectsCount int `json:"heap_objects_count"`
		// TotalAllocBytes is the bytes allocated and not yet freed.
		AllocBytes int `json:"alloc_bytes"`
	}

	checkResponse struct {
		name      string
		skipOnErr bool
		err       error
	}
)

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
	defer close(resChan)
	for _, c := range checks {
		go func(c Config) {
			resChan <- checkResponse{
				name:      c.Name,
				skipOnErr: c.SkipOnErr,
				err:       c.Check(),
			}
		}(c)

	loop:
		for {
			select {
			case <-time.After(c.Timeout):
				failures[c.Name] = failureTimeout
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

// Reset unregisters all previously set check configs
func Reset() {
	mu.Lock()
	defer mu.Unlock()

	checks = []Config{}
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
