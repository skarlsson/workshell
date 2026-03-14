package deps

import (
	"fmt"
	"os/exec"
	"strings"
)

type ToolStatus struct {
	Name     string
	Required bool
	Found    bool
	Path     string
	Note     string
}

var tools = []struct {
	Name     string
	Required bool
	Note     string
}{
	{"kitty", true, "terminal emulator"},
	{"zellij", true, "terminal multiplexer"},
	{"git", true, "branch/task management"},
	{"xdotool", false, "window move/focus/minimize"},
	{"xprop", false, "window title updates"},
	{"gdbus", false, "monitor detection (GNOME/Mutter)"},
}

// CheckAll returns the status of all required and optional tools.
func CheckAll() []ToolStatus {
	var results []ToolStatus
	for _, t := range tools {
		ts := ToolStatus{
			Name:     t.Name,
			Required: t.Required,
			Note:     t.Note,
		}
		if path, err := exec.LookPath(t.Name); err == nil {
			ts.Found = true
			ts.Path = path
		}
		results = append(results, ts)
	}
	return results
}

// CheckRequired verifies that all required tools (kitty, zellij, git) are in PATH.
func CheckRequired() error {
	var missing []string
	for _, t := range tools {
		if !t.Required {
			continue
		}
		if _, err := exec.LookPath(t.Name); err != nil {
			missing = append(missing, t.Name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("required tools not found in PATH: %s\nRun 'ws doctor' for details or 'bash install_deps.sh' to install", strings.Join(missing, ", "))
	}
	return nil
}

// HasTool checks if a single tool is available in PATH.
func HasTool(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
