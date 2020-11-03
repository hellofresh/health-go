package health

import "fmt"

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
