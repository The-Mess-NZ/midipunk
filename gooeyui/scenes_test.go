package gooeyui

import (
	"image"
	"testing"

	gooeycomponents "github.com/The-Mess-NZ/gooey/pkg/components"
)

func TestBuildPortsSceneAddsSoftButtonsAndBindings(t *testing.T) {
	scene, err := buildPortsScene(testConfig(), 0)
	if err != nil {
		t.Fatalf("buildPortsScene() error = %v", err)
	}

	bar := scene.Root.Children[len(scene.Root.Children)-1]
	if bar.Type != "soft_button_bar" {
		t.Fatalf("soft bar type = %q, want %q", bar.Type, "soft_button_bar")
	}
	if got := []string{bar.SoftButtons[0].Label, bar.SoftButtons[1].Label, bar.SoftButtons[2].Label, bar.SoftButtons[3].Label}; got[0] != "" || got[1] != "" || got[2] != ">" || got[3] != ">>" {
		t.Fatalf("ports soft buttons = %v, want [  > >>]", got)
	}
	if len(scene.InputBindings) != 2 || scene.InputBindings[0].ID != softButtonThreeID || scene.InputBindings[1].ID != softButtonFourID {
		t.Fatalf("ports inputBindings = %+v, want button3/button4 only", scene.InputBindings)
	}
}

func TestBuildRoutesSceneAddsSoftButtonsAndBindings(t *testing.T) {
	scene, err := buildRoutesScene(testConfig(), "port-1", 0)
	if err != nil {
		t.Fatalf("buildRoutesScene() error = %v", err)
	}

	bar := scene.Root.Children[len(scene.Root.Children)-1]
	if got := []string{bar.SoftButtons[0].Label, bar.SoftButtons[1].Label, bar.SoftButtons[2].Label, bar.SoftButtons[3].Label}; got[0] != "ports" || got[1] != "" || got[2] != "next" || got[3] != "" {
		t.Fatalf("routes soft buttons = %v, want [ports  next ]", got)
	}
	if len(scene.InputBindings) != 2 || scene.InputBindings[0].ID != softButtonOneID || scene.InputBindings[1].ID != softButtonThreeID {
		t.Fatalf("routes inputBindings = %+v, want button1/button3 only", scene.InputBindings)
	}
}

func TestPortsSoftBarIsPinnedToBottom(t *testing.T) {
	scene, err := buildPortsScene(testConfig(), 1)
	if err != nil {
		t.Fatalf("buildPortsScene() error = %v", err)
	}
	assertSoftBarBottom(t, scene, "ports-soft-bar")
}

func TestRoutesSoftBarIsPinnedToBottom(t *testing.T) {
	scene, err := buildRoutesScene(testConfig(), "port-1", 0)
	if err != nil {
		t.Fatalf("buildRoutesScene() error = %v", err)
	}
	assertSoftBarBottom(t, scene, "routes-soft-bar")
}

func assertSoftBarBottom(t *testing.T, scene gooeycomponents.SceneDocument, componentID string) {
	t.Helper()
	components, err := gooeycomponents.BuildScene(scene, image.Rect(0, 0, 320, 240))
	if err != nil {
		t.Fatalf("BuildScene() error = %v", err)
	}
	bar := findComponentByID(components, componentID)
	if bar == nil {
		t.Fatalf("component %q not found", componentID)
	}
	if got := bar.BoundingBox().Dy(); got != softBarHeight {
		t.Fatalf("soft bar height = %d, want %d", got, softBarHeight)
	}
	if got := bar.BoundingBox().Max.Y; got != 240 {
		t.Fatalf("soft bar bottom = %d, want 240", got)
	}
}

func findComponentByID(components []gooeycomponents.Component, id string) gooeycomponents.Component {
	for _, component := range components {
		if component.ID() == id {
			return component
		}
	}
	return nil
}
