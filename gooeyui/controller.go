package gooeyui

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	gooeycomponents "github.com/The-Mess-NZ/gooey/pkg/components"
	gooeyipc "github.com/The-Mess-NZ/gooey/pkg/ipc"
	"github.com/The-Mess-NZ/midipunk/config"
	"github.com/The-Mess-NZ/midipunk/logger"
	"github.com/The-Mess-NZ/midipunk/midirouter"
)

const (
	defaultSocketPath  = "/tmp/gooey.sock"
	reconnectDelay     = 2 * time.Second
	actionBackPorts    = "back_ports"
	actionViewPrefix   = "view_port:"
	actionTogglePrefix = "toggle_route:"
	actionPageFirst    = "page_first"
	actionPagePrev     = "page_prev"
	actionPageNext     = "page_next"
	actionPageLast     = "page_last"
)

type sceneView string

type hardwareInputReport struct {
	Action string `json:"action,omitempty"`
}

const (
	sceneViewPorts  sceneView = "ports"
	sceneViewRoutes sceneView = "routes"
)

// Controller drives Gooey from Midipunk's current config state.
type Controller struct {
	socketPath     string
	client         *gooeyipc.Client
	config         *config.MidiPunkConfig
	routes         []*midirouter.Route
	currentView    sceneView
	portsPage      int
	routesPage     int
	selectedPortID string
}

// NewController creates a Gooey UI controller for the supplied Midipunk config.
func NewController(cfg *config.MidiPunkConfig, routes []*midirouter.Route) *Controller {
	socketPath := os.Getenv("MIDIPUNK_GOOEY_SOCKET")
	if socketPath == "" {
		socketPath = defaultSocketPath
	}
	return &Controller{socketPath: socketPath, config: cfg, routes: routes, currentView: sceneViewPorts}
}

// Run maintains the Gooey client connection and handles UI navigation events.
func (c *Controller) Run(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			c.closeClient()
			return
		}

		if c.client == nil {
			if err := c.connectAndSubmitCurrentScene(); err != nil {
				logger.Error("Gooey UI connection failed: %v", err)
				if !waitForRetry(ctx, reconnectDelay) {
					c.closeClient()
					return
				}
				continue
			}
		}

		event, err := c.client.NextEvent()
		if err != nil {
			logger.Error("Gooey UI event loop lost connection: %v", err)
			if reconnectErr := c.client.ReconnectAndResubmit(); reconnectErr != nil {
				logger.Error("Failed to reconnect to Gooey: %v", reconnectErr)
				c.closeClient()
				if !waitForRetry(ctx, reconnectDelay) {
					return
				}
				continue
			}
			logger.Info("Reconnected to Gooey and resubmitted the current Midipunk scene")
			continue
		}

		c.handleEvent(event)
	}
}

func (c *Controller) connectAndSubmitCurrentScene() error {
	client, err := gooeyipc.Dial(c.socketPath)
	if err != nil {
		return err
	}

	scene, err := c.currentScene()
	if err != nil {
		client.Close()
		return err
	}
	if err := client.SubmitScene(scene); err != nil {
		client.Close()
		return err
	}
	if err := client.RequestStatus(); err != nil {
		logger.Error("Failed to request Gooey status snapshot: %v", err)
	}

	c.client = client
	logger.Info("Connected Midipunk UI to Gooey at %s", c.socketPath)
	return nil
}

func (c *Controller) handleEvent(event gooeyipc.Event) {
	switch event.Type {
	case gooeyipc.EventComponentActivate:
		var interaction gooeycomponents.Interaction
		if err := gooeyipc.DecodeEventPayload(event, &interaction); err != nil {
			logger.Error("Failed to decode Gooey component activation: %v", err)
			return
		}
		c.handleAction(interaction.Action)

	case gooeyipc.EventInputEvent:
		var report hardwareInputReport
		if err := gooeyipc.DecodeEventPayload(event, &report); err != nil {
			logger.Error("Failed to decode Gooey hardware input event: %v", err)
			return
		}
		c.handleAction(report.Action)

	case gooeyipc.EventProtocolError:
		var protocolErr gooeyipc.ErrorPayload
		if err := gooeyipc.DecodeEventPayload(event, &protocolErr); err != nil {
			logger.Error("Failed to decode Gooey protocol error: %v", err)
			return
		}
		logger.Error("Gooey protocol error [%s] for %s: %s", protocolErr.Code, protocolErr.Action, protocolErr.Message)

	case gooeyipc.EventDiagnostics:
		var diagnostics gooeyipc.DiagnosticsPayload
		if err := gooeyipc.DecodeEventPayload(event, &diagnostics); err != nil {
			logger.Error("Failed to decode Gooey diagnostics: %v", err)
			return
		}
		if diagnostics.Status != nil {
			logger.Info("Gooey status: sceneLoaded=%t root=%s components=%d touchConfigured=%t",
				diagnostics.Status.SceneLoaded,
				diagnostics.Status.RootID,
				diagnostics.Status.ComponentCount,
				diagnostics.Status.TouchConfigured,
			)
			return
		}
		if diagnostics.Message != "" {
			logger.Info("Gooey diagnostics [%s]: %s", diagnostics.Kind, diagnostics.Message)
		}

	case gooeyipc.EventSceneLoaded:
		var ack gooeyipc.AckPayload
		if err := gooeyipc.DecodeEventPayload(event, &ack); err == nil {
			logger.Info("Gooey scene loaded: root=%s version=%s", ack.ID, ack.Version)
		}
	}
}

func (c *Controller) handleAction(action string) {
	if action == "" {
		return
	}

	changed, err := c.applyAction(action)
	if err != nil {
		logger.Error("Failed to handle Gooey action %s: %v", action, err)
		return
	}
	if !changed {
		return
	}
	if err := c.replaceCurrentScene(action); err != nil {
		logger.Error("Failed to replace Gooey scene after action %s: %v", action, err)
	}
}

func (c *Controller) applyAction(action string) (bool, error) {
	switch {
	case action == actionBackPorts:
		if c.currentView == sceneViewPorts {
			return false, nil
		}
		c.currentView = sceneViewPorts
		c.selectedPortID = ""
		c.routesPage = 0
		return true, nil
	case strings.HasPrefix(action, actionViewPrefix):
		portID := strings.TrimPrefix(action, actionViewPrefix)
		if _, ok := findPort(c.config, portID); !ok {
			return false, sceneErrorf("unknown port %q", portID)
		}
		if c.currentView == sceneViewRoutes && c.selectedPortID == portID {
			return false, nil
		}
		c.currentView = sceneViewRoutes
		if c.selectedPortID != portID {
			c.routesPage = 0
		}
		c.selectedPortID = portID
		return true, nil
	case strings.HasPrefix(action, actionTogglePrefix):
		if err := c.toggleRoute(strings.TrimPrefix(action, actionTogglePrefix)); err != nil {
			return false, err
		}
		return true, nil
	case action == actionPageFirst:
		return c.setCurrentPage(0)
	case action == actionPagePrev:
		return c.setCurrentPage(c.currentPage() - 1)
	case action == actionPageNext:
		return c.setCurrentPage(c.currentPage() + 1)
	case action == actionPageLast:
		return c.setCurrentPage(c.totalPages() - 1)
	default:
		logger.Info("Ignoring unknown Gooey UI action: %s", action)
		return false, nil
	}
}

func (c *Controller) replaceCurrentScene(action string) error {
	scene, err := c.currentScene()
	if err != nil {
		return err
	}
	if c.client == nil {
		return nil
	}
	if err := c.client.ReplaceScene(scene); err != nil {
		return err
	}
	if err := c.client.RequestStatus(); err != nil {
		logger.Error("Failed to request Gooey status after action %s: %v", action, err)
	}
	logger.Info("Updated Gooey scene after action %s", action)
	return nil
}

// TODO: The current scene is determined by whether there's a selectedPortID or not. This is not the desired pattern.... But we're just testing for now.
func (c *Controller) currentScene() (gooeycomponents.SceneDocument, error) {
	if c.currentView == sceneViewRoutes && c.selectedPortID != "" {
		return buildRoutesScene(c.config, c.selectedPortID, c.routesPage)
	}
	return buildPortsScene(c.config, c.portsPage)
}

func (c *Controller) currentPage() int {
	if c.currentView == sceneViewRoutes {
		return c.routesPage
	}
	return c.portsPage
}

func (c *Controller) setCurrentPage(page int) (bool, error) {
	totalPages := c.totalPages()
	if totalPages < 1 {
		return false, nil
	}
	nextPage := clampPageIndex(page, totalPages)
	if nextPage == c.currentPage() {
		return false, nil
	}
	if c.currentView == sceneViewRoutes {
		c.routesPage = nextPage
		return true, nil
	}
	c.portsPage = nextPage
	return true, nil
}

func (c *Controller) totalPages() int {
	if c.currentView == sceneViewRoutes && c.selectedPortID != "" {
		return routesPageCount(c.config, c.selectedPortID)
	}
	return portsPageCount(c.config)
}

func (c *Controller) toggleRoute(indexText string) error {
	index, err := strconv.Atoi(indexText)
	if err != nil {
		return sceneErrorf("invalid route index %q", indexText)
	}
	if index < 0 || index >= len(c.config.Routes) {
		return sceneErrorf("route index %d out of range", index)
	}

	enabled := !c.config.Routes[index].IsEnabled()
	c.config.Routes[index].SetEnabled(enabled)
	if index < len(c.routes) && c.routes[index] != nil {
		c.routes[index].SetEnabled(enabled)
	}
	logger.Info("Route %d now %s", index, enabledState(enabled))
	return nil
}

func (c *Controller) closeClient() {
	if c.client != nil {
		_ = c.client.Close()
		c.client = nil
	}
}

func waitForRetry(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func sceneErrorf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

func enabledState(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}
