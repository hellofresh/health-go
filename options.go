package health

import (
	"fmt"

	"go.opentelemetry.io/otel/trace"
)

// Option is the health-container options type
type Option func(*Health) error

// WithChecks adds checks to newly instantiated health-container
func WithChecks(checks ...Config) Option {
	return func(h *Health) error {
		for _, c := range checks {
			if err := h.Register(c); err != nil {
				return fmt.Errorf("could not register check %q: %w", c.Name, err)
			}
		}

		return nil
	}
}

// WithTracerProvider sets trace provider for the checks and instrumentation name that will be used
// for tracer from trace provider.
func WithTracerProvider(tp trace.TracerProvider, instrumentationName string) Option {
	return func(h *Health) error {
		h.tp = tp
		h.instrumentationName = instrumentationName

		return nil
	}
}
