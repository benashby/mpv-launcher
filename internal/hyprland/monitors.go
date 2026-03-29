package hyprland

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)

// Monitor represents a single Hyprland monitor as returned by -j/monitors.
type Monitor struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	RefreshRate float64 `json:"refreshRate"`
	Focused     bool    `json:"focused"`
	Disabled    bool    `json:"disabled"`
}

// IsAvailable reports whether the process is running inside Hyprland.
func IsAvailable() bool {
	return os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != ""
}

// socketPath returns the path to the Hyprland command socket.
func socketPath() (string, error) {
	sig := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if sig == "" {
		return "", fmt.Errorf("HYPRLAND_INSTANCE_SIGNATURE not set")
	}
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		return "", fmt.Errorf("XDG_RUNTIME_DIR not set")
	}
	return fmt.Sprintf("%s/hypr/%s/.socket.sock", runtimeDir, sig), nil
}

// GetMonitors queries the Hyprland socket for the current monitor list.
// Only active (non-disabled) monitors are returned.
func GetMonitors() ([]Monitor, error) {
	path, err := socketPath()
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("unix", path)
	if err != nil {
		return nil, fmt.Errorf("connecting to hyprland socket: %w", err)
	}
	defer conn.Close()

	if _, err := conn.Write([]byte("-j/monitors")); err != nil {
		return nil, fmt.Errorf("sending monitors command: %w", err)
	}

	var buf []byte
	tmp := make([]byte, 4096)
	for {
		n, err := conn.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil { // io.EOF signals end of response
			break
		}
	}

	var all []Monitor
	if err := json.Unmarshal(buf, &all); err != nil {
		return nil, fmt.Errorf("parsing monitor JSON: %w", err)
	}

	active := make([]Monitor, 0, len(all))
	for _, m := range all {
		if !m.Disabled {
			active = append(active, m)
		}
	}
	return active, nil
}
