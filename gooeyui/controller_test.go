package gooeyui

import (
	"encoding/json"
	"testing"

	gooeyipc "github.com/The-Mess-NZ/gooey/pkg/ipc"
	"github.com/The-Mess-NZ/midipunk/config"
)

func TestControllerHandleActionPagesPortsAndRoutes(t *testing.T) {
	controller := NewController(testConfig(), nil)

	controller.handleAction(actionPageNext)
	if controller.portsPage != 1 {
		t.Fatalf("portsPage = %d, want 1", controller.portsPage)
	}

	controller.handleAction(actionViewPrefix + "port-1")
	if controller.currentView != sceneViewRoutes {
		t.Fatalf("currentView = %q, want %q", controller.currentView, sceneViewRoutes)
	}
	if controller.routesPage != 0 {
		t.Fatalf("routesPage = %d, want 0", controller.routesPage)
	}

	controller.handleAction(actionPageNext)
	if controller.routesPage != 1 {
		t.Fatalf("routesPage = %d, want 1", controller.routesPage)
	}

	controller.handleAction(actionBackPorts)
	if controller.currentView != sceneViewPorts {
		t.Fatalf("currentView = %q, want %q", controller.currentView, sceneViewPorts)
	}
	if controller.portsPage != 1 {
		t.Fatalf("portsPage after back = %d, want 1", controller.portsPage)
	}
}

func TestControllerHandleEventUsesHardwareInputAction(t *testing.T) {
	controller := NewController(testConfig(), nil)
	payload, err := json.Marshal(hardwareInputReport{Action: actionPageNext})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	controller.handleEvent(gooeyipc.Event{Type: gooeyipc.EventInputEvent, Payload: payload})
	if controller.portsPage != 1 {
		t.Fatalf("portsPage = %d, want 1", controller.portsPage)
	}
}

func testConfig() *config.MidiPunkConfig {
	enabled := true
	return &config.MidiPunkConfig{
		Label: "MidiPunk",
		Ports: []config.PortConfigYAML{
			{ID: "port-1", Type: "USB", Direction: "INPUT", Label: "Port 1"},
			{ID: "port-2", Type: "USB", Direction: "OUTPUT", Label: "Port 2"},
			{ID: "port-3", Type: "DIN", Direction: "INPUT", Label: "Port 3"},
			{ID: "port-4", Type: "DIN", Direction: "OUTPUT", Label: "Port 4"},
			{ID: "port-5", Type: "USB", Direction: "INPUT", Label: "Port 5"},
		},
		Routes: []config.RouteYAML{
			{Label: "Route 1", Enabled: &enabled, Inputs: []config.RouteInputYAML{{PortID: "port-1"}}, Outputs: []config.RouteOutputYAML{{PortID: "port-2"}}},
			{Label: "Route 2", Enabled: &enabled, Inputs: []config.RouteInputYAML{{PortID: "port-1"}}, Outputs: []config.RouteOutputYAML{{PortID: "port-3"}}},
			{Label: "Route 3", Enabled: &enabled, Inputs: []config.RouteInputYAML{{PortID: "port-1"}}, Outputs: []config.RouteOutputYAML{{PortID: "port-4"}}},
			{Label: "Route 4", Enabled: &enabled, Inputs: []config.RouteInputYAML{{PortID: "port-1"}}, Outputs: []config.RouteOutputYAML{{PortID: "port-5"}}},
		},
	}
}
