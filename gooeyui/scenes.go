package gooeyui

import (
	"fmt"
	"strings"

	gooeycomponents "github.com/The-Mess-NZ/gooey/pkg/components"
	"github.com/The-Mess-NZ/midipunk/config"
)

// TODO: Separate scenes out into their own files.

const (
	portsPageSize     = 4
	routesPageSize    = 3
	scenePadding      = 8
	sceneGap          = 6
	softBarHeight     = 22
	portsRowHeight    = 32
	routesRowHeight   = 38
	softButtonCount   = 4
	softButtonOneID   = "button1"
	softButtonTwoID   = "button2"
	softButtonThreeID = "button3"
	softButtonFourID  = "button4"
)

type paginationState struct {
	CurrentPage int
	TotalPages  int
	HasPrev     bool
	HasNext     bool
}

type softButtonSpec struct {
	Label  string
	Action string
}

func buildPortsScene(cfg *config.MidiPunkConfig, page int) (gooeycomponents.SceneDocument, error) {
	if cfg == nil {
		return gooeycomponents.SceneDocument{}, sceneErrorf("config is nil")
	}

	visiblePorts, paging := paginateItems(cfg.Ports, page, portsPageSize)
	listChildren := make([]gooeycomponents.SceneNode, 0, len(visiblePorts))
	for _, port := range visiblePorts {
		listChildren = append(listChildren, makePortNode(cfg, port))
	}
	if len(listChildren) == 0 {
		listChildren = append(listChildren, emptyStateNode("ports-empty", "No ports configured."))
	}
	slots := portsSoftButtons(paging)

	bodyChildren := []gooeycomponents.SceneNode{
		makeTitleNode("ports-title", fmt.Sprintf("%s Ports %d/%d", cfg.Label, paging.CurrentPage+1, paging.TotalPages)),
		{
			ID:       "ports-list",
			Type:     gooeycomponents.NodeTypeContainer,
			Layout:   &gooeycomponents.Layout{Direction: gooeycomponents.LayoutDirectionVertical, Gap: sceneGap},
			Children: listChildren,
		},
	}
	children := []gooeycomponents.SceneNode{
		makeSceneBodyNode("ports-body", bodyChildren),
		makeSoftButtonBarNode("ports-soft-bar", slots),
	}

	return gooeycomponents.SceneDocument{
		Version:       gooeycomponents.SceneVersionAlpha1,
		InputBindings: softButtonBindings(slots),
		Root: gooeycomponents.SceneNode{
			ID:   "midipunk-ports-root",
			Type: gooeycomponents.NodeTypeContainer,
			Layout: &gooeycomponents.Layout{
				Direction: gooeycomponents.LayoutDirectionVertical,
				Gap:       0,
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
		Bounds: &gooeycomponents.Rect{Height: portsRowHeight},
		Layout: &gooeycomponents.Layout{Direction: gooeycomponents.LayoutDirectionHorizontal, Gap: 8, Padding: gooeycomponents.Insets{Top: 4, Right: 6, Bottom: 4, Left: 6}},
		Style: &gooeycomponents.Style{
			Background:  portButtonBackground(port.Direction),
			BorderColor: "#9BB3C3FF",
			BorderWidth: 1,
		},
		Children: []gooeycomponents.SceneNode{
			{
				ID:     "port-copy-" + sanitizeID(port.ID),
				Type:   gooeycomponents.NodeTypeContainer,
				Layout: &gooeycomponents.Layout{Direction: gooeycomponents.LayoutDirectionVertical, Gap: 1, Padding: gooeycomponents.Insets{Top: 0, Right: 0, Bottom: 0, Left: 0}},
				Children: []gooeycomponents.SceneNode{
					{
						ID:     "port-label-" + sanitizeID(port.ID),
						Type:   gooeycomponents.NodeTypeLabel,
						Text:   portLabel,
						Bounds: &gooeycomponents.Rect{Height: 14},
						Style:  &gooeycomponents.Style{Foreground: "#F2F2E9", FontSize: 12, TextPadding: 1},
					},
					{
						ID:     "port-meta-" + sanitizeID(port.ID),
						Type:   gooeycomponents.NodeTypeLabel,
						Text:   fmt.Sprintf("%s %s", strings.ToUpper(port.Type), strings.ToUpper(port.Direction)),
						Bounds: &gooeycomponents.Rect{Height: 10},
						Style:  &gooeycomponents.Style{Foreground: "#D6E4ED", FontSize: 9, TextPadding: 1},
					},
				},
			},
			{
				ID:     "port-badges-" + sanitizeID(port.ID),
				Type:   gooeycomponents.NodeTypeContainer,
				Bounds: &gooeycomponents.Rect{Width: 70},
				Layout: &gooeycomponents.Layout{Direction: gooeycomponents.LayoutDirectionVertical, Gap: 2},
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
		Bounds: &gooeycomponents.Rect{Height: 12},
		Style: &gooeycomponents.Style{
			Background:  background,
			Foreground:  "#F2F2E9",
			BorderColor: border,
			BorderWidth: 1,
			FontSize:    9,
			TextPadding: 2,
		},
	}
}

func buildRoutesScene(cfg *config.MidiPunkConfig, portID string, page int) (gooeycomponents.SceneDocument, error) {
	if cfg == nil {
		return gooeycomponents.SceneDocument{}, sceneErrorf("config is nil")
	}

	port, ok := findPort(cfg, portID)
	if !ok {
		return gooeycomponents.SceneDocument{}, sceneErrorf("unknown port %q", portID)
	}

	associated := routesForPort(cfg, portID)
	visibleRoutes, paging := paginateItems(associated, page, routesPageSize)
	listChildren := make([]gooeycomponents.SceneNode, 0, len(visibleRoutes))
	if len(visibleRoutes) == 0 {
		listChildren = append(listChildren, emptyStateNode("route-empty", "No routes use this port."))
	} else {
		for _, route := range visibleRoutes {
			listChildren = append(listChildren, makeRouteNode(portID, route))
		}
	}
	slots := routesSoftButtons(paging)
	bodyChildren := []gooeycomponents.SceneNode{
		makeTitleNode("routes-title", fmt.Sprintf("%s %d/%d", portSceneTitle(port), paging.CurrentPage+1, paging.TotalPages)),
		makeSubheadNode("routes-subtitle", fmt.Sprintf("%d route(s) use port %s", len(associated), port.ID)),
		{
			ID:       "routes-list",
			Type:     gooeycomponents.NodeTypeContainer,
			Layout:   &gooeycomponents.Layout{Direction: gooeycomponents.LayoutDirectionVertical, Gap: sceneGap},
			Children: listChildren,
		},
	}
	children := []gooeycomponents.SceneNode{
		makeSceneBodyNode("routes-body", bodyChildren),
		makeSoftButtonBarNode("routes-soft-bar", slots),
	}

	return gooeycomponents.SceneDocument{
		Version:       gooeycomponents.SceneVersionAlpha1,
		InputBindings: softButtonBindings(slots),
		Root: gooeycomponents.SceneNode{
			ID:   "midipunk-routes-root",
			Type: gooeycomponents.NodeTypeContainer,
			Layout: &gooeycomponents.Layout{
				Direction: gooeycomponents.LayoutDirectionVertical,
				Gap:       0,
			},
			Style:    &gooeycomponents.Style{Background: "#0F1418FF"},
			Children: children,
		},
	}, nil
}

// TODO: Consider, the sizing like '214' is hard-coded. Could this be more % based sizing?
func makeRouteNode(portID string, route indexedRoute) gooeycomponents.SceneNode {
	return gooeycomponents.SceneNode{
		ID:     fmt.Sprintf("route-card-%s-%d", sanitizeID(portID), route.Index),
		Type:   gooeycomponents.NodeTypeContainer,
		Bounds: &gooeycomponents.Rect{Height: routesRowHeight},
		Layout: &gooeycomponents.Layout{Direction: gooeycomponents.LayoutDirectionHorizontal, Gap: 8, Padding: gooeycomponents.Insets{Top: 4, Right: 4, Bottom: 4, Left: 4}},
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
				Bounds: &gooeycomponents.Rect{Width: 214},
				Style:  &gooeycomponents.Style{Foreground: "#F2F2E9", FontSize: 11, TextPadding: 3},
			},
			{
				ID:      fmt.Sprintf("route-toggle-%s-%d", sanitizeID(portID), route.Index),
				Type:    gooeycomponents.NodeTypeToggle,
				Action:  fmt.Sprintf("%s%d", actionTogglePrefix, route.Index),
				Bounds:  &gooeycomponents.Rect{Width: 72},
				Checked: boolPtr(route.IsEnabled()),
				Style: &gooeycomponents.Style{
					Background:  toggleBackground(route.IsEnabled()),
					Foreground:  "#F2F2E9",
					BorderColor: "#B7D5C5FF",
					BorderWidth: 2,
				},
			},
		},
	}
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
		Bounds: &gooeycomponents.Rect{Height: 18},
		Style:  &gooeycomponents.Style{Foreground: "#AFC2D0", FontSize: 10, TextPadding: 3},
	}
}

func emptyStateNode(id, text string) gooeycomponents.SceneNode {
	return gooeycomponents.SceneNode{
		ID:     id,
		Type:   gooeycomponents.NodeTypeLabel,
		Text:   text,
		Bounds: &gooeycomponents.Rect{Height: 36},
		Style:  &gooeycomponents.Style{Foreground: "#D8DEE4", FontSize: 12, TextPadding: 4},
	}
}

func makeSceneBodyNode(id string, children []gooeycomponents.SceneNode) gooeycomponents.SceneNode {
	return gooeycomponents.SceneNode{
		ID:       id,
		Type:     gooeycomponents.NodeTypeContainer,
		Layout:   &gooeycomponents.Layout{Direction: gooeycomponents.LayoutDirectionVertical, Gap: sceneGap, Padding: gooeycomponents.Insets{Top: scenePadding, Right: scenePadding, Bottom: sceneGap, Left: scenePadding}},
		Children: children,
	}
}

func makeSoftButtonBarNode(id string, slots [softButtonCount]softButtonSpec) gooeycomponents.SceneNode {
	softButtons := make([]gooeycomponents.SoftButtonSlot, 0, len(slots))
	for _, slot := range slots {
		softButtons = append(softButtons, gooeycomponents.SoftButtonSlot{Label: slot.Label})
	}
	return gooeycomponents.SceneNode{
		ID:          id,
		Type:        gooeycomponents.NodeTypeSoftBar,
		Bounds:      &gooeycomponents.Rect{Height: softBarHeight},
		SoftButtons: softButtons,
		Style: &gooeycomponents.Style{
			Background:  "#314554FF",
			Foreground:  "#F2F2E9",
			BorderColor: "#A8C2D4FF",
			BorderWidth: 1,
			FontSize:    10,
			TextPadding: 1,
		},
	}
}

func softButtonBindings(slots [softButtonCount]softButtonSpec) []gooeycomponents.InputBinding {
	inputIDs := [softButtonCount]string{softButtonOneID, softButtonTwoID, softButtonThreeID, softButtonFourID}
	bindings := make([]gooeycomponents.InputBinding, 0, len(slots))
	for index, slot := range slots {
		if slot.Action == "" {
			continue
		}
		bindings = append(bindings, gooeycomponents.InputBinding{ID: inputIDs[index], OnPress: slot.Action})
	}
	return bindings
}

func portsSoftButtons(paging paginationState) [softButtonCount]softButtonSpec {
	slots := [softButtonCount]softButtonSpec{}
	if paging.CurrentPage > 0 {
		slots[0] = softButtonSpec{Label: "<<", Action: actionPageFirst}
		slots[1] = softButtonSpec{Label: "<", Action: actionPagePrev}
	}
	if paging.CurrentPage < paging.TotalPages-1 {
		slots[2] = softButtonSpec{Label: ">", Action: actionPageNext}
		slots[3] = softButtonSpec{Label: ">>", Action: actionPageLast}
	}
	return slots
}

func routesSoftButtons(paging paginationState) [softButtonCount]softButtonSpec {
	slots := [softButtonCount]softButtonSpec{{Label: "ports", Action: actionBackPorts}}
	if paging.HasPrev {
		slots[1] = softButtonSpec{Label: "prev", Action: actionPagePrev}
	}
	if paging.HasNext {
		slots[2] = softButtonSpec{Label: "next", Action: actionPageNext}
	}
	return slots
}

func portsPageCount(cfg *config.MidiPunkConfig) int {
	if cfg == nil {
		return 1
	}
	return pageCount(len(cfg.Ports), portsPageSize)
}

func routesPageCount(cfg *config.MidiPunkConfig, portID string) int {
	if cfg == nil {
		return 1
	}
	return pageCount(len(routesForPort(cfg, portID)), routesPageSize)
}

func pageCount(totalItems, pageSize int) int {
	if pageSize <= 0 || totalItems <= 0 {
		return 1
	}
	return (totalItems + pageSize - 1) / pageSize
}

func clampPageIndex(page, totalPages int) int {
	if totalPages <= 1 {
		return 0
	}
	if page < 0 {
		return 0
	}
	if page >= totalPages {
		return totalPages - 1
	}
	return page
}

func paginateItems[T any](items []T, page, pageSize int) ([]T, paginationState) {
	state := paginationState{TotalPages: pageCount(len(items), pageSize)}
	state.CurrentPage = clampPageIndex(page, state.TotalPages)
	state.HasPrev = state.CurrentPage > 0
	state.HasNext = state.CurrentPage < state.TotalPages-1
	if len(items) == 0 {
		return nil, state
	}
	start := state.CurrentPage * pageSize
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], state
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
