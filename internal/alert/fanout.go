package alert

import (
	"errors"
	"fmt"
)

// Fanout distributes a single alert to multiple Sink implementations.
type Fanout struct {
	sinks []Sink
}

// NewFanout creates a Fanout that delivers to all provided sinks.
func NewFanout(sinks ...Sink) *Fanout {
	return &Fanout{sinks: sinks}
}

// Send delivers the alert to every registered sink, collecting any errors.
func (f *Fanout) Send(a Alert) error {
	var errs []error
	for _, s := range f.sinks {
		if err := s.Send(a); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("fanout: %w", errors.Join(errs...))
	}
	return nil
}

// Add appends a new sink to the fanout at runtime.
func (f *Fanout) Add(s Sink) {
	f.sinks = append(f.sinks, s)
}

// Len returns the number of registered sinks.
func (f *Fanout) Len() int {
	return len(f.sinks)
}
