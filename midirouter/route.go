package midirouter

import "sync"

/*
A Route takes multiple inputs, and distributes the MIDI
messages to the given outputs.
*/
type Route struct {
	mu      sync.RWMutex
	inputs  []*RouteInput
	outputs []*RouteOutput
	enabled bool
	Label   string // Label for UI
}

// NewRoute creates a new Route
func NewRoute(inputs []*RouteInput, outputs []*RouteOutput, label string, enabled bool) *Route {
	return &Route{
		inputs:  inputs,
		outputs: outputs,
		enabled: enabled,
		Label:   label,
	}
}

// GetInputs returns the route inputs
func (r *Route) GetInputs() []*RouteInput {
	return r.inputs
}

// GetOutputs returns the route outputs
func (r *Route) GetOutputs() []*RouteOutput {
	return r.outputs
}

// AddInput adds an input to the route
func (r *Route) AddInput(input *RouteInput) {
	r.inputs = append(r.inputs, input)
}

// AddOutput adds an output to the route
func (r *Route) AddOutput(output *RouteOutput) {
	r.outputs = append(r.outputs, output)
}

// IsEnabled reports whether the route is active.
func (r *Route) IsEnabled() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.enabled
}

// SetEnabled toggles whether the route should actively route MIDI messages.
func (r *Route) SetEnabled(enabled bool) {
	r.mu.Lock()
	r.enabled = enabled
	r.mu.Unlock()
}
