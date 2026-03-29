package hyprland

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
)

// mockSocket sets up a temp dir with the Hyprland socket path structure and serves response once.
// Sets HYPRLAND_INSTANCE_SIGNATURE and XDG_RUNTIME_DIR env vars via t.Setenv.
func mockSocket(t *testing.T, response []byte) {
	t.Helper()
	dir := t.TempDir()
	sockDir := filepath.Join(dir, "hypr", "test-sig")
	os.MkdirAll(sockDir, 0755)
	sockPath := filepath.Join(sockDir, ".socket.sock")

	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("failed to listen on mock socket: %v", err)
	}
	t.Cleanup(func() { ln.Close() })

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 64)
		conn.Read(buf)
		conn.Write(response)
	}()

	t.Setenv("HYPRLAND_INSTANCE_SIGNATURE", "test-sig")
	t.Setenv("XDG_RUNTIME_DIR", dir)
}

func TestGetMonitors_ParsesJSON(t *testing.T) {
	monitors := []Monitor{
		{ID: 0, Name: "DP-1", Description: "ASUS XG27JCG", Width: 5120, Height: 2880, RefreshRate: 165.0, Focused: true, Disabled: false},
		{ID: 1, Name: "HDMI-A-2", Description: "LG Display", Width: 1920, Height: 1080, RefreshRate: 60.0, Focused: false, Disabled: false},
	}
	data, _ := json.Marshal(monitors)
	mockSocket(t, data)

	got, err := GetMonitors()
	if err != nil {
		t.Fatalf("GetMonitors() error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 monitors, got %d", len(got))
	}
	if got[0].Name != "DP-1" {
		t.Errorf("expected DP-1, got %s", got[0].Name)
	}
	if got[1].Name != "HDMI-A-2" {
		t.Errorf("expected HDMI-A-2, got %s", got[1].Name)
	}
}

func TestGetMonitors_FiltersDisabled(t *testing.T) {
	monitors := []Monitor{
		{ID: 0, Name: "DP-1", Disabled: false},
		{ID: 1, Name: "HDMI-A-2", Disabled: true}, // should be filtered
		{ID: 2, Name: "DP-2", Disabled: false},
	}
	data, _ := json.Marshal(monitors)
	mockSocket(t, data)

	got, err := GetMonitors()
	if err != nil {
		t.Fatalf("GetMonitors() error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 active monitors, got %d", len(got))
	}
	for _, m := range got {
		if m.Disabled {
			t.Errorf("disabled monitor %q should have been filtered out", m.Name)
		}
	}
}

func TestIsAvailable(t *testing.T) {
	t.Setenv("HYPRLAND_INSTANCE_SIGNATURE", "")
	if IsAvailable() {
		t.Error("IsAvailable() should return false when env var is empty")
	}

	t.Setenv("HYPRLAND_INSTANCE_SIGNATURE", "some-sig")
	if !IsAvailable() {
		t.Error("IsAvailable() should return true when env var is set")
	}
}

func TestGetMonitors_NoEnvVar(t *testing.T) {
	t.Setenv("HYPRLAND_INSTANCE_SIGNATURE", "")
	_, err := GetMonitors()
	if err == nil {
		t.Error("GetMonitors() should return error when HYPRLAND_INSTANCE_SIGNATURE is not set")
	}
}
