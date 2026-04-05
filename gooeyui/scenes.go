package gooeyui

import (
	"fmt"
	"strings"

	gooeycomponents "github.com/The-Mess-NZ/gooey/pkg/components"
	"github.com/The-Mess-NZ/midipunk/config"
)

func buildPortsScene(cfg *config.MidiPunkConfig) (gooeycomponents.SceneDocument, error) {
	if cfg == nil {
		return gooeycomponents.SceneDocument{}, sceneErrorf("config is nil")
	}

	children := []gooeycomponents.SceneNode{
		makeTitleNode("ports-title", "MidiPunk Ports"),
		makeSubheadNode("ports-subtitle", fmt.Sprintf("%s • tap a port to view routes", cfg.Label)),
	}

	for _, port := range cfg.Ports {
		text := port.Label
		if text == "" {
			text = port.ID
		}
		text = fmt.Sprintf("%s\n%s %s", text, strings.ToUpper(port.Type), strings.ToUpper(port.Direction))
		children = append(children, gooeycomponents.SceneNode{
			ID:     "port-" + sanitizeID(port.ID),
			Type:   gooeycomponents.NodeTypeButton,
			Text:   text,
			Action: actionViewPrefix + port.ID,
			Bounds: &gooeycomponents.Rect{Height: 42},
			Style: &gooeycomponents.Style{
				Background:  portButtonBackground(port.Direction),
				Foreground:  "#F2F2E9",
				BorderColor: "#9BB3C3FF",
				BorderWidth: 2,
				FontSize:    13,
				TextPadding: 5,
			},
		})
	}

	children = append(children, makeFooterNode("ports-footer", fmt.Sprintf("%d configured port(s)", len(cfg.Ports))))

	return gooeycomponents.SceneDocument{
		Version: gooeycomponents.SceneVersionAlpha1,
		Root: gooeycomponents.SceneNode{
			ID:   "midipunk-ports-root",
			Type: gooeycomponents.NodeTypeContainer,
			Layout: &gooeycomponents.Layout{
				Direction: gooeycomponents.LayoutDirectionVertical,
				Gap:       8,
				Padding:   gooeycomponents.Insets{Top: 8, Right: 8, Bottom: 8, Left: 8},
			},
			Style:    &gooeycomponents.Style{Background: "#0F1418FF"},
			Children: children,
		},
	}, nil
}

func buildRoutesScene(cfg *config.MidiPunkConfig, portID string) (gooeycomponents.SceneDocument, error) {
	if cfg == nil {
		return gooeycomponents.SceneDocument{}, sceneErrorf("config is nil")
	}

	port, ok := findPort(cfg, portID)
	if !ok {
		return gooeycomponents.SceneDocument{}, sceneErrorf("unknown port %q", portID)
	}

	children := []gooeycomponents.SceneNode{
		makeTitleNode("routes-title", portSceneTitle(port)),
		makeSubheadNode("routes-subtitle", fmt.Sprintf("Routes that use port %s", port.ID)),
	}

	associated := routesForPort(cfg, portID)
	if len(associated) == 0 {
		children = append(children, gooeycomponents.SceneNode{
			ID:     "route-empty",
			Type:   gooeycomponents.NodeTypeLabel,
			Text:   "No routes use this port.",
			Bounds: &gooeycomponents.Rect{Height: 36},
			Style:  &gooeycomponents.Style{Foreground: "#D8DEE4", FontSize: 13, TextPadding: 4},
		})
	} else {
		for index, route := range associated {
			children = append(children, gooeycomponents.SceneNode{
				ID:     fmt.Sprintf("route-%s-%d", sanitizeID(portID), index),
				Type:   gooeycomponents.NodeTypeLabel,
				Text:   routeSummary(route),
				Bounds: &gooeycomponents.Rect{Height: 54},
				Style: &gooeycomponents.Style{
					Background:  "#1A252DFF",
					Foreground:  "#F2F2E9",
					BorderColor: "#415664FF",
					BorderWidth: 1,
					FontSize:    12,
					TextPadding: 5,
				},
			})
		}
	}

	children = append(children, gooeycomponents.SceneNode{
		ID:     "routes-back",
		Type:   gooeycomponents.NodeTypeButton,
		Text:   "Back",
		Action: actionBackPorts,
		Bounds: &gooeycomponents.Rect{Height: 34},
		Style: &gooeycomponents.Style{
			Background:  "#314554FF",
			Foreground:  "#F2F2E9",
			BorderColor: "#A8C2D4FF",
			BorderWidth: 2,
			FontSize:    13,
			TextPadding: 4,
		},
	})

	return gooeycomponents.SceneDocument{
		Version: gooeycomponents.SceneVersionAlpha1,
		Root: gooeycomponents.SceneNode{
			ID:   "midipunk-routes-root",
			Type: gooeycomponents.NodeTypeContainer,
			Layout: &gooeycomponents.Layout{
				Direction: gooeycomponents.LayoutDirectionVertical,
				Gap:       8,
				Padding:   gooeycomponents.Insets{Top: 8, Right: 8, Bottom: 8, Left: 8},
			},
			Style:    &gooeycomponents.Style{Background: "#0F1418FF"},
			Children: children,
		},
	}, nil
}

func makeTitleNode(id, text string) gooeycomponents.SceneNode {
	return gooeycomponents.SceneNode{
		ID:     id,
		Type:   gooeycomponents.NodeTypeLabel,
		Text:   text,
		Bounds: &gooeycomponents.Rect{Height: 26},
		Style: &gooeycomponents.Style{
			Background:  "#1A2730FF",
			Foreground:  "#F7F1D5",
			FontSize:    16,
			TextPadding: 4,
		},
	}
}

func makeSubheadNode(id, text string) gooeycomponents.SceneNode {
	return gooeycomponents.SceneNode{
		ID:     id,
		Type:   gooeycomponents.NodeTypeLabel,
		Text:   text,
		Bounds: &gooeycomponents.Rect{Height: 28},
		Style:  &gooeycomponents.Style{Foreground: "#AFC2D0", FontSize: 11, TextPadding: 4},
	}
}

func makeFooterNode(id, text string) gooeycomponents.SceneNode {
	return gooeycomponents.SceneNode{
		ID:     id,
		Type:   gooeycomponents.NodeTypeLabel,
		Text:   text,
		Bounds: &gooeycomponents.Rect{Height: 18},
		Style:  &gooeycomponents.Style{Foreground: "#7F919F", FontSize: 10, TextPadding: 4},
	}
}

func portButtonBackground(direction string) string {
	if strings.EqualFold(direction, "INPUT") {
		return "#2F5E72FF"
	}
	return "#5E4B2FFF"
}

func portSceneTitle(port config.PortConfigYAML) string {
	label := port.Label
	if label == "" {
		label = port.ID
	}
	return fmt.Sprintf("%s Routes", label)
}

func routesForPort(cfg *config.MidiPunkConfig, portID string) []config.RouteYAML {
	matches := make([]config.RouteYAML, 0)
	for _, route := range cfg.Routes {
		if routeUsesPort(route, portID) {
			matches = append(matches, route)
		}
	}
	return matches
}

func routeUsesPort(route config.RouteYAML, portID string) bool {
	for _, input := range route.Inputs {
		if input.PortID == portID {
			return true
		}
	}
	for _, output := range route.Outputs {
		if output.PortID == portID {
			return true
		}
	}
	return false
}

func routeSummary(route config.RouteYAML) string {
	label := route.Label
	if label == "" {
		label = "Route"
	}
	status := "enabled"
	if !route.IsEnabled() {
		status = "disabled"
	}
	return fmt.Sprintf("%s (%s)\n%s -> %s", label, status, joinRouteInputs(route.Inputs), joinRouteOutputs(route.Outputs))
}

func joinRouteInputs(inputs []config.RouteInputYAML) string {
	parts := make([]string, 0, len(inputs))
	for _, input := range inputs {
		part := input.PortID
		if input.Channel > 0 {
			part = fmt.Sprintf("%s ch%d", part, input.Channel)
		}
		parts = append(parts, part)
	}
	if len(parts) == 0 {
		return "?"
	}
	return strings.Join(parts, ", ")
}

func joinRouteOutputs(outputs []config.RouteOutputYAML) string {
	parts := make([]string, 0, len(outputs))
	for _, output := range outputs {
		part := output.PortID
		if output.Channel > 0 {
			part = fmt.Sprintf("%s ch%d", part, output.Channel)
		}
		parts = append(parts, part)
	}
	if len(parts) == 0 {
		return "?"
	}
	return strings.Join(parts, ", ")
}

func findPort(cfg *config.MidiPunkConfig, portID string) (config.PortConfigYAML, bool) {
	for _, port := range cfg.Ports {
		if port.ID == portID {
			return port, true
		}
	}
	return config.PortConfigYAML{}, false
}

func sanitizeID(value string) string {
	replacer := strings.NewReplacer("/", "-", " ", "-", ":", "-", ".", "-")
	return replacer.Replace(value)
}
