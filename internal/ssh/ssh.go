package ssh

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Run executes a command on the remote host and returns its output.
// Uses -A for agent forwarding so remote can use local SSH keys.
func Run(target, command string) (string, error) {
	cmd := exec.Command("ssh", "-A", target, command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ssh %s: %w\n%s", target, err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

// RunWithTimeout executes a command with a ConnectTimeout.
func RunWithTimeout(target, command string, timeout time.Duration) (string, error) {
	secs := int(timeout.Seconds())
	if secs < 1 {
		secs = 1
	}
	cmd := exec.Command("ssh", "-A", "-o", fmt.Sprintf("ConnectTimeout=%d", secs), target, command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ssh %s: %w\n%s", target, err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

// EnsureGitHubHostKey adds github.com to known_hosts on the remote if not already present.
func EnsureGitHubHostKey(target string) error {
	_, err := Run(target, "ssh-keygen -F github.com >/dev/null 2>&1 || ssh-keyscan -t ed25519 github.com >> ~/.ssh/known_hosts 2>/dev/null")
	return err
}

// InteractiveCommand returns the full SSH command string for sending to kitty via send-text.
// Ensures ~/.local/bin is in PATH since non-login shells may not source .bashrc.
func InteractiveCommand(target, command string) string {
	return fmt.Sprintf("ssh -A %s -t 'export PATH=\"$HOME/.local/bin:$PATH\" && %s'\n", target, command)
}

// CheckConnection verifies SSH connectivity to a host.
func CheckConnection(target string) error {
	_, err := RunWithTimeout(target, "echo ok", 5*time.Second)
	return err
}

// GetArch returns the remote machine architecture (e.g. "x86_64", "aarch64").
func GetArch(target string) (string, error) {
	return RunWithTimeout(target, "uname -m", 5*time.Second)
}

// CheckZellijSession checks if a named zellij session exists on the remote host.
func CheckZellijSession(target, session string) bool {
	out, err := RunWithTimeout(target, "zellij list-sessions --short 2>/dev/null", 5*time.Second)
	if err != nil {
		return false
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == session {
			return true
		}
	}
	return false
}

// RemoteStatus represents a workspace status returned from a remote ws instance.
type RemoteStatus struct {
	Name       string `json:"name"`
	Dir        string `json:"dir"`
	Branch     string `json:"branch"`
	Task       string `json:"task,omitempty"`
	Active     bool   `json:"active"`
	Claude     bool   `json:"claude"`
	ClaudeTime string `json:"claude_cpu_time,omitempty"`
}

// GetRemoteStatuses queries all workspace statuses from a remote host.
func GetRemoteStatuses(target string) ([]RemoteStatus, error) {
	out, err := RunWithTimeout(target, "export PATH=\"$HOME/.local/bin:$PATH\" && ws status 2>/dev/null", 10*time.Second)
	if err != nil {
		return nil, err
	}
	var statuses []RemoteStatus
	if err := json.Unmarshal([]byte(out), &statuses); err != nil {
		return nil, fmt.Errorf("parsing remote status: %w", err)
	}
	return statuses, nil
}

// CopyFile copies a local file to a remote path via scp.
func CopyFile(target, localPath, remotePath string) error {
	cmd := exec.Command("scp", localPath, fmt.Sprintf("%s:%s", target, remotePath))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("scp to %s: %w\n%s", target, err, strings.TrimSpace(string(out)))
	}
	return nil
}
