package zellij

import (
	"fmt"
	"os/exec"
	"strings"
)

func run(args ...string) (string, error) {
	cmd := exec.Command("zellij", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("zellij %s: %w\n%s", strings.Join(args, " "), err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func ListSessions() ([]string, error) {
	out, err := run("list-sessions", "--short")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}

func KillSession(name string) error {
	_, err := run("kill-session", name)
	return err
}

// DeleteSession removes a dead session. Use before creating a new session with the same name.
func DeleteSession(name string) error {
	_, err := run("delete-session", name, "--force")
	return err
}

// CleanupSession kills and deletes a session, ignoring errors (session may not exist).
func CleanupSession(name string) {
	KillSession(name)
	DeleteSession(name)
}

func SessionExists(name string) bool {
	sessions, err := ListSessions()
	if err != nil {
		return false
	}
	for _, s := range sessions {
		// zellij list-sessions --short may include metadata after the name
		// e.g. "ws-foo [Created ...] (EXITED)" — match on prefix
		if s == name || strings.HasPrefix(s, name+" ") {
			return true
		}
	}
	return false
}

// LaunchCommand returns the shell command string to start zellij with a layout.
// This is meant to be sent to kitty via send-text, not executed directly.
// Uses --new-session-with-layout to always start fresh (avoids session picker).
func LaunchCommand(session, layoutPath, cwd string) string {
	if layoutPath != "" {
		return fmt.Sprintf("cd %s && zellij -s %s -n %s\n", cwd, session, layoutPath)
	}
	return fmt.Sprintf("cd %s && zellij -s %s\n", cwd, session)
}
