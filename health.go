package health

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

type Health struct {
	mu       sync.Mutex
	checkMap map[string]Config
	//WithSysMetrics is the conditional to process system metrics
	//If it true log and response system metrics
	WithSysMetrics bool
	// ErrorLogFunc is the callback function for errors logging during check.
	// If not set logging is skipped.
	ErrorLogFunc func(err error, details string, extra ...interface{})
	// DebugLogFunc is the callback function for debug logging during check.
	// If not set logging is skipped.
	DebugLogFunc func(...interface{})
}

func New(WithSysMetrics bool, ErrorLogFunc func(err error, details string, extra ...interface{}), DebugLogFunc func(...interface{})) Health {
	if ErrorLogFunc == nil {
		ErrorLogFunc = func(err error, details string, extra ...interface{}) {}
	}
	if DebugLogFunc == nil {
		DebugLogFunc = func(...interface{}) {}
	}

	return Health{
		mu:             sync.Mutex{},
		checkMap:       make(map[string]Config),
		WithSysMetrics: WithSysMetrics,
		ErrorLogFunc: ErrorLogFunc,
		DebugLogFunc: DebugLogFunc,
	}
}

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

// Register allot of checks
func (h *Health) BulkRegister(c ...Config) (err error) {
	for _, cf := range c {
		err = h.Register(cf)
		if err != nil {
			return
		}
	}
	return
}

// Register registers a check config to be performed.
func (h *Health) Register(c Config) error {
	if c.Timeout == 0 {
		c.Timeout = time.Second * 2
	}

	if c.Name == "" {
		return errors.New("health check must have a name to be registered")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.checkMap[c.Name]; ok {
		return fmt.Errorf("health check %s is already registered", c.Name)
	}

	h.checkMap[c.Name] = c

	return nil
}

//Execute a health check standalone if the flag chosen is true
func (h *Health) HealthCheckStandaloneMode(flagName string) {
	b := flag.Bool(flagName, false, "Flag used to ability the health check mode")
	flag.Parse()
	if *b {
		h.ExecuteStandalone()
	}
}

// Handler returns an HTTP handler (http.HandlerFunc).
func (h *Health) Handler() http.Handler {
	return http.HandlerFunc(h.HandlerFunc)
}

// HandlerFunc is the HTTP handler function.
func (h *Health) HandlerFunc(w http.ResponseWriter, r *http.Request) {
	h.DebugLogFunc("Handling health check func")
	c := h.ExecuteCheck()
	h.logCheck(c)
	data, err := json.Marshal(c)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.ErrorLogFunc(err, "Error to parse health check result")
		return
	}

	code := http.StatusOK
	if c.Status == statusUnavailable {
		code = http.StatusServiceUnavailable
	}
	w.WriteHeader(code)
	_, _ = w.Write(data)
}

//Execute health check base on config map
func (h *Health) ExecuteCheck() Check {
	h.mu.Lock()
	defer h.mu.Unlock()

	status := statusOK
	total := len(h.checkMap)
	failures := make(map[string]string)
	resChan := make(chan checkResponse, total)

	var wg sync.WaitGroup
	wg.Add(total)

	go func() {
		defer close(resChan)
		wg.Wait()
	}()

	for _, c := range h.checkMap {
		go func(c Config) {
			h.DebugLogFunc("Executing health check:", c.Name)
			defer wg.Done()
			select {
			case resChan <- checkResponse{c.Name, c.SkipOnErr, c.Check()}:
			default:
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

	c := h.newCheck(status, failures)
	return c
}

//Execute health check in standalone which if it is not ok return code 1 to system
func (h *Health) ExecuteStandalone() {
	c := h.ExecuteCheck()
	h.logCheck(c)
	if c.Status == statusOK {
		os.Exit(0)
	}
	os.Exit(1)
}

// Reset unregisters all previously set check configs
func (h *Health) Reset() {
	h.mu.Lock()
	h.DebugLogFunc("Reseting health check configs")
	defer h.mu.Unlock()

	h.checkMap = make(map[string]Config)
}

func (h *Health) logCheck(c Check) {
	b, e := json.MarshalIndent(c, "", " ")
	if e != nil {
		h.ErrorLogFunc(e, "Error to parse Health Check Result")
	}
	h.DebugLogFunc("Health Check Result:\n", string(b))
}

func (h *Health) newCheck(status string, failures map[string]string) Check {
	c := Check{
		Status:    status,
		Timestamp: time.Now(),
		Failures:  failures,
	}

	if h.WithSysMetrics {
		c.System = newSystemMetrics()
	}

	return c
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
