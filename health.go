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

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/trace"
)

// Status type represents health status
type Status string

// Possible health statuses
const (
	StatusOK                 Status = "OK"
	StatusPartiallyAvailable Status = "Partially Available"
	StatusUnavailable        Status = "Unavailable"
	StatusTimeout            Status = "Timeout during health check"
)

type (
	// CheckFunc is the func which executes the check.
	CheckFunc func(context.Context) error

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
		Status Status `json:"status"`
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

	// Health is the health-checks container
	Health struct {
		mu     sync.Mutex
		checks map[string]Config

		tp                  trace.TracerProvider
		instrumentationName string
	}

	checkResponse struct {
		name      string
		skipOnErr bool
		err       error
	}
)

// New instantiates and build new health check container
func New(opts ...Option) (*Health, error) {
	h := &Health{
		checks: make(map[string]Config),
		tp:     trace.NewNoopTracerProvider(),
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

	code := http.StatusOK
	if c.Status == StatusUnavailable {
		code = http.StatusServiceUnavailable
	}
	w.WriteHeader(code)
	w.Write(data)
}

type checkSpan struct {
	ctx  context.Context
	span trace.Span
}

func newCheckSpan(ctx context.Context, tracer trace.Tracer, name string) checkSpan {
	var cs checkSpan
	cs.ctx, cs.span = tracer.Start(ctx, name)
	return cs
}

// Measure runs all the registered health checks and returns summary status
func (h *Health) Measure(ctx context.Context) Check {
	h.mu.Lock()
	defer h.mu.Unlock()

	tracer := h.tp.Tracer(h.instrumentationName)

	ctx, span := tracer.Start(ctx, "health.Measure")
	defer span.End()

	status := StatusOK
	total := len(h.checks)
	failures := make(map[string]string)
	resChan := make(chan checkResponse, total)
	checkSpans := make(map[string]checkSpan)

	span.SetAttributes(label.Int("checks", total))

	var wg sync.WaitGroup
	wg.Add(total)

	go func() {
		defer close(resChan)

		wg.Wait()
	}()

	for _, c := range h.checks {
		checkSpans[c.Name] = newCheckSpan(ctx, tracer, c.Name)

		go func(c Config) {
			defer wg.Done()

			select {
			case resChan <- checkResponse{c.Name, c.SkipOnErr, c.Check(ctx)}:
			default:
			}
		}(c)

	loop:
		for {
			select {
			case <-time.After(c.Timeout):
				failures[c.Name] = string(StatusTimeout)
				status = getAvailability(status, c.SkipOnErr)

				cs := checkSpans[c.Name]
				cs.span.SetStatus(codes.Error, string(StatusTimeout))
				cs.span.End()

				break loop
			case res := <-resChan:
				cs := checkSpans[res.name]

				if res.err != nil {
					failures[res.name] = res.err.Error()
					status = getAvailability(status, res.skipOnErr)

					cs.span.RecordError(res.err)
				}

				cs.span.End()

				break loop
			}
		}
	}

	span.SetAttributes(label.String("status", string(status)))

	return newCheck(status, failures)
}

func newCheck(s Status, failures map[string]string) Check {
	return Check{
		Status:    s,
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

func getAvailability(s Status, skipOnErr bool) Status {
	if skipOnErr && s != StatusUnavailable {
		return StatusPartiallyAvailable
	}

	return StatusUnavailable
}
