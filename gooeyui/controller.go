package gooeyui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	gooeycomponents "github.com/The-Mess-NZ/gooey/pkg/components"
	gooeyipc "github.com/The-Mess-NZ/gooey/pkg/ipc"
	"github.com/The-Mess-NZ/midipunk/config"
	"github.com/The-Mess-NZ/midipunk/logger"
)

const (
	defaultSocketPath = "/tmp/gooey.sock"
	reconnectDelay    = 2 * time.Second
	actionBackPorts   = "back_ports"
	actionViewPrefix  = "view_port:"
)

// Controller drives Gooey from Midipunk's current config state.
type Controller struct {
	socketPath     string
	client         *gooeyipc.Client
	config         *config.MidiPunkConfig
	selectedPortID string
}

// NewController creates a Gooey UI controller for the supplied Midipunk config.
func NewController(cfg *config.MidiPunkConfig) *Controller {
	socketPath := os.Getenv("MIDIPUNK_GOOEY_SOCKET")
	if socketPath == "" {
		socketPath = defaultSocketPath
	}
	return &Controller{socketPath: socketPath, config: cfg}
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

	switch {
	case action == actionBackPorts:
		c.selectedPortID = ""
	case strings.HasPrefix(action, actionViewPrefix):
		c.selectedPortID = strings.TrimPrefix(action, actionViewPrefix)
	default:
		logger.Info("Ignoring unknown Gooey UI action: %s", action)
		return
	}

	scene, err := c.currentScene()
	if err != nil {
		logger.Error("Failed to build Gooey scene for action %s: %v", action, err)
		return
	}
	if err := c.client.ReplaceScene(scene); err != nil {
		logger.Error("Failed to replace Gooey scene after action %s: %v", action, err)
		return
	}
	if err := c.client.RequestStatus(); err != nil {
		logger.Error("Failed to request Gooey status after action %s: %v", action, err)
	}
	logger.Info("Updated Gooey scene after action %s", action)
}

// TODO: The current scene is determined by whether there's a selectedPortID or not. This is not the desired pattern.... But we're just testing for now.
func (c *Controller) currentScene() (gooeycomponents.SceneDocument, error) {
	if c.selectedPortID == "" {
		return buildPortsScene(c.config)
	}
	return buildRoutesScene(c.config, c.selectedPortID)
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
