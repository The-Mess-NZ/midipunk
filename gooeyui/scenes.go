package gooeyui

import (
	"fmt"
	"strings"

	gooeycomponents "github.com/The-Mess-NZ/gooey/pkg/components"
	"github.com/The-Mess-NZ/midipunk/config"
)

// TODO: Separate scenes out into their own files.

func buildPortsScene(cfg *config.MidiPunkConfig) (gooeycomponents.SceneDocument, error) {
	if cfg == nil {
		return gooeycomponents.SceneDocument{}, sceneErrorf("config is nil")
	}

	children := []gooeycomponents.SceneNode{
		makeTitleNode("ports-title", fmt.Sprintf("%s", cfg.Label)),
	}

	for _, port := range cfg.Ports {
		children = append(children, makePortNode(cfg, port))
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

func makePortNode(cfg *config.MidiPunkConfig, port config.PortConfigYAML) gooeycomponents.SceneNode {
	counts := portRouteCounts(cfg, port.ID)
	portLabel := port.Label
	if portLabel == "" {
		portLabel = port.ID
	}

	return gooeycomponents.SceneNode{
		ID:     "port-" + sanitizeID(port.ID),
		Type:   gooeycomponents.NodeTypeContainer,
		Action: actionViewPrefix + port.ID,
		Bounds: &gooeycomponents.Rect{Height: 54},
		Layout: &gooeycomponents.Layout{Direction: gooeycomponents.LayoutDirectionHorizontal, Gap: 10, Padding: gooeycomponents.Insets{Top: 6, Right: 8, Bottom: 6, Left: 8}},
		Style: &gooeycomponents.Style{
			Background:  portButtonBackground(port.Direction),
			BorderColor: "#9BB3C3FF",
			BorderWidth: 2,
		},
		Children: []gooeycomponents.SceneNode{
			{
				ID:     "port-copy-" + sanitizeID(port.ID),
				Type:   gooeycomponents.NodeTypeContainer,
				Layout: &gooeycomponents.Layout{Direction: gooeycomponents.LayoutDirectionVertical, Gap: 2, Padding: gooeycomponents.Insets{Top: 2, Right: 0, Bottom: 2, Left: 0}},
				Children: []gooeycomponents.SceneNode{
					{
						ID:     "port-label-" + sanitizeID(port.ID),
						Type:   gooeycomponents.NodeTypeLabel,
						Text:   portLabel,
						Bounds: &gooeycomponents.Rect{Height: 18},
						Style:  &gooeycomponents.Style{Foreground: "#F2F2E9", FontSize: 14, TextPadding: 2},
					},
					{
						ID:     "port-meta-" + sanitizeID(port.ID),
						Type:   gooeycomponents.NodeTypeLabel,
						Text:   fmt.Sprintf("%s %s", strings.ToUpper(port.Type), strings.ToUpper(port.Direction)),
						Bounds: &gooeycomponents.Rect{Height: 14},
						Style:  &gooeycomponents.Style{Foreground: "#D6E4ED", FontSize: 11, TextPadding: 2},
					},
				},
			},
			{
				ID:     "port-badges-" + sanitizeID(port.ID),
				Type:   gooeycomponents.NodeTypeContainer,
				Bounds: &gooeycomponents.Rect{Width: 82},
				Layout: &gooeycomponents.Layout{Direction: gooeycomponents.LayoutDirectionVertical, Gap: 4},
				Children: []gooeycomponents.SceneNode{
					makePortBadgeNode("port-active-badge-"+sanitizeID(port.ID), fmt.Sprintf("ON %d", counts.Active), "#2F6B4FFF", "#9ED8B5FF"),
					makePortBadgeNode("port-inactive-badge-"+sanitizeID(port.ID), fmt.Sprintf("OFF %d", counts.Inactive), "#3D4650FF", "#8D9AA6FF"),
				},
			},
		},
	}
}

func makePortBadgeNode(id, text, background, border string) gooeycomponents.SceneNode {
	return gooeycomponents.SceneNode{
		ID:     id,
		Type:   gooeycomponents.NodeTypeLabel,
		Text:   text,
		Bounds: &gooeycomponents.Rect{Height: 18},
		Style: &gooeycomponents.Style{
			Background:  background,
			Foreground:  "#F2F2E9",
			BorderColor: border,
			BorderWidth: 1,
			FontSize:    10,
			TextPadding: 3,
		},
	}
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
		for _, route := range associated {
			children = append(children, gooeycomponents.SceneNode{
				ID:     fmt.Sprintf("route-card-%s-%d", sanitizeID(portID), route.Index),
				Type:   gooeycomponents.NodeTypeContainer,
				Bounds: &gooeycomponents.Rect{Height: 62},
				Layout: &gooeycomponents.Layout{Direction: gooeycomponents.LayoutDirectionHorizontal, Gap: 8, Padding: gooeycomponents.Insets{Top: 6, Right: 6, Bottom: 6, Left: 6}},
				Style: &gooeycomponents.Style{
					Background:  "#1A252DFF",
					BorderColor: "#415664FF",
					BorderWidth: 1,
				},
				Children: []gooeycomponents.SceneNode{
					{
						ID:     fmt.Sprintf("route-label-%s-%d", sanitizeID(portID), route.Index),
						Type:   gooeycomponents.NodeTypeLabel,
						Text:   routeSummary(route.RouteYAML),
						Bounds: &gooeycomponents.Rect{Width: 150},
						Style:  &gooeycomponents.Style{Foreground: "#F2F2E9", FontSize: 12, TextPadding: 4},
					},
					{
						ID:      fmt.Sprintf("route-toggle-%s-%d", sanitizeID(portID), route.Index),
						Type:    gooeycomponents.NodeTypeToggle,
						Action:  fmt.Sprintf("%s%d", actionTogglePrefix, route.Index),
						Checked: boolPtr(route.IsEnabled()),
						Style: &gooeycomponents.Style{
							Background:  toggleBackground(route.IsEnabled()),
							Foreground:  "#F2F2E9",
							BorderColor: "#B7D5C5FF",
							BorderWidth: 2,
						},
					},
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

type indexedRoute struct {
	Index int
	config.RouteYAML
}

type routeCounts struct {
	Active   int
	Inactive int
}

func routesForPort(cfg *config.MidiPunkConfig, portID string) []indexedRoute {
	matches := make([]indexedRoute, 0)
	for index, route := range cfg.Routes {
		if routeUsesPort(route, portID) {
			matches = append(matches, indexedRoute{Index: index, RouteYAML: route})
		}
	}
	return matches
}

func portRouteCounts(cfg *config.MidiPunkConfig, portID string) routeCounts {
	counts := routeCounts{}
	for _, route := range cfg.Routes {
		if !routeUsesPort(route, portID) {
			continue
		}
		if route.IsEnabled() {
			counts.Active++
			continue
		}
		counts.Inactive++
	}
	return counts
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

func boolPtr(value bool) *bool {
	return &value
}

func toggleBackground(enabled bool) string {
	if enabled {
		return "#2F6B4FFF"
	}
	return "#4A5056FF"
}
