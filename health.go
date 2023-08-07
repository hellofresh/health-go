package health

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Status type represents health status
type Status string

const (
	// StatusPassing healthcheck is passing
	StatusPassing Status = "passing"
	// StatusWarning healthcheck is failing but should not fail the component
	StatusWarning Status = "warning"
	// StatusCritical healthcheck is failing should fail the component
	StatusCritical Status = "critical"
	// StatusTimeout healthcheck timed out should fail the component
	StatusTimeout Status = "timeout"
)

type (
	// CheckFunc is the func which executes the check.
	CheckFunc func(context.Context) CheckResponse

	// Config carries the parameters to run the check.
	Config struct {
		// Name is the name of the resource to be checked.
		Name string
		// Timeout is the timeout defined for every check.
		Timeout time.Duration
		// SkipOnErr if set to true, it will retrieve StatusPassing providing the error message from the failed resource.
		SkipOnErr bool
		// Check is the func which executes the check.
		Check CheckFunc
	}

	// Check represents the health check response.
	Check struct {
		// Status is the check status.
		Status Status `json:"status"`
		// Timestamp is the time in which the check occurred.
		Timestamp time.Time `json:"timestamp"`
		// Failures holds the failed checks along with their messages.
		Failures map[string]string `json:"failures,omitempty"`
		// System holds information of the go process.
		*System `json:"system,omitempty"`
		// Component holds information on the component for which checks are made
		Component `json:"component"`
	}

	CheckResponse struct {
		// Error message
		Error error

		// IsWarning if set to true, it will retrieve StatusPassing providing the error message from the failed resource.
		IsWarning bool
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

	// Component descriptive values about the component for which checks are made
	Component struct {
		// Name is the name of the component.
		Name string `json:"name"`
		// Version is the component version.
		Version string `json:"version"`
	}

	// Health is the health-checks container
	Health struct {
		mu            sync.Mutex
		checks        map[string]Config
		maxConcurrent int

		tp                  trace.TracerProvider
		instrumentationName string

		component Component

		systemInfoEnabled bool
	}
)

// New instantiates and build new health check container
func New(opts ...Option) (*Health, error) {
	h := &Health{
		checks:        make(map[string]Config),
		tp:            trace.NewNoopTracerProvider(),
		maxConcurrent: runtime.NumCPU(),
	}

	for _, o := range opts {
		if err := o(h); err != nil {
			return nil, err
		}
	}

	return h, nil
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

	if _, ok := h.checks[c.Name]; ok {
		return fmt.Errorf("health check %q is already registered", c.Name)
	}

	h.checks[c.Name] = c

	return nil
}

// Handler returns an HTTP handler (http.HandlerFunc).
func (h *Health) Handler() http.Handler {
	return http.HandlerFunc(h.HandlerFunc)
}

// HandlerFunc is the HTTP handler function.
func (h *Health) HandlerFunc(w http.ResponseWriter, r *http.Request) {
	c := h.Measure(r.Context())

	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(c)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Default status should always be failing
	// Return 503 indicating service is unavailable
	code := http.StatusServiceUnavailable

	// Return 200 when passing
	if c.Status == StatusPassing {
		code = http.StatusOK
	}

	// Return 429 indicating to try again later
	if c.Status == StatusWarning {
		code = http.StatusTooManyRequests
	}

	w.WriteHeader(code)
	w.Write(data)
}

// Measure runs all the registered health checks and returns summary status
func (h *Health) Measure(ctx context.Context) Check {
	h.mu.Lock()
	defer h.mu.Unlock()

	tracer := h.tp.Tracer(h.instrumentationName)

	ctx, span := tracer.Start(
		ctx,
		"health.Measure",
		trace.WithAttributes(attribute.Int("checks", len(h.checks))),
	)
	defer span.End()

	status := StatusPassing
	failures := make(map[string]string)

	limiterCh := make(chan bool, h.maxConcurrent)
	defer close(limiterCh)

	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)
	for _, c := range h.checks {
		limiterCh <- true
		wg.Add(1)

		go func(c Config) {
			ctx, span := tracer.Start(ctx, c.Name)
			defer func() {
				span.End()
				<-limiterCh
				wg.Done()
			}()

			resCh := make(chan CheckResponse)

			go func() {
				resCh <- c.Check(ctx)
				defer close(resCh)
			}()

			timeout := time.NewTimer(c.Timeout)

			select {
			case <-timeout.C:
				mu.Lock()
				defer mu.Unlock()

				span.SetStatus(codes.Error, string(StatusTimeout))

				failures[c.Name] = string(StatusTimeout)
				status = getAvailability(status, c.SkipOnErr)
			case res := <-resCh:
				if !timeout.Stop() {
					<-timeout.C
				}

				mu.Lock()
				defer mu.Unlock()

				if res.Error != nil {
					span.RecordError(res.Error)

					failures[c.Name] = res.Error.Error()
					status = getAvailability(status, c.SkipOnErr || res.IsWarning)
				}
			}
		}(c)
	}

	wg.Wait()
	span.SetAttributes(attribute.String("status", string(status)))

	var systemMetrics *System
	if h.systemInfoEnabled {
		systemMetrics = newSystemMetrics()
	}

	return newCheck(h.component, status, systemMetrics, failures)
}

func newCheck(c Component, s Status, system *System, failures map[string]string) Check {
	return Check{
		Status:    s,
		Timestamp: time.Now(),
		Failures:  failures,
		System:    system,
		Component: c,
	}
}

func newSystemMetrics() *System {
	s := runtime.MemStats{}
	runtime.ReadMemStats(&s)

	return &System{
		Version:          runtime.Version(),
		GoroutinesCount:  runtime.NumGoroutine(),
		TotalAllocBytes:  int(s.TotalAlloc),
		HeapObjectsCount: int(s.HeapObjects),
		AllocBytes:       int(s.Alloc),
	}
}

func getAvailability(s Status, skipOnErr bool) Status {
	if skipOnErr && s != StatusCritical {
		return StatusWarning
	}

	return StatusCritical
}
