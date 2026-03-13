package kitty

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func SocketPath(wsName string) string {
	return fmt.Sprintf("/tmp/kitty-ws-%s", wsName)
}

type kittyWindow struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	PID   int    `json:"pid"`
}

type kittyTab struct {
	ID      int           `json:"id"`
	Windows []kittyWindow `json:"windows"`
}

type kittyOSWindow struct {
	ID             int        `json:"id"`
	PlatformWinID  int        `json:"platform_window_id"`
	Tabs           []kittyTab `json:"tabs"`
}

func remoteCmd(socket string, args ...string) (string, error) {
	fullArgs := append([]string{"@", "--to", "unix:" + socket}, args...)
	cmd := exec.Command("kitty", fullArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("kitty @ %s: %w\n%s", strings.Join(args, " "), err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// Launch starts a new kitty instance for a workspace.
// Returns the PID of the kitty process.
func Launch(wsName, cwd, title string) (int, error) {
	socket := SocketPath(wsName)

	cmd := exec.Command("kitty",
		"--listen-on", "unix:"+socket,
		"--directory", cwd,
		"--title", title,
		"--override", "allow_remote_control=yes",
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	// Clean environment: remove vars that prevent tools like claude from starting
	env := os.Environ()
	cleanEnv := make([]string, 0, len(env))
	for _, e := range env {
		if strings.HasPrefix(e, "CLAUDECODE=") {
			continue
		}
		cleanEnv = append(cleanEnv, e)
	}
	cmd.Env = cleanEnv

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("starting kitty: %w", err)
	}

	return cmd.Process.Pid, nil
}

// PlatformWindowID returns the X11/XWayland window ID for a workspace's kitty instance.
func PlatformWindowID(wsName string) (int, error) {
	socket := SocketPath(wsName)
	out, err := remoteCmd(socket, "ls")
	if err != nil {
		return 0, err
	}
	var osWindows []kittyOSWindow
	if err := json.Unmarshal([]byte(out), &osWindows); err != nil {
		return 0, fmt.Errorf("parsing kitty ls: %w", err)
	}
	if len(osWindows) == 0 {
		return 0, fmt.Errorf("no kitty OS windows found for %q", wsName)
	}
	return osWindows[0].PlatformWinID, nil
}

// Activate raises and focuses a workspace's kitty window using xdotool.
func Activate(wsName string) error {
	winID, err := PlatformWindowID(wsName)
	if err != nil {
		return err
	}
	cmd := exec.Command("xdotool", "windowactivate", strconv.Itoa(winID))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("xdotool windowactivate: %w\n%s", err, string(out))
	}
	return nil
}

func SendText(socket string, text string) error {
	_, err := remoteCmd(socket, "send-text", text)
	return err
}

// KillProcess sends SIGTERM to a kitty process by PID.
func KillProcess(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Signal(syscall.SIGTERM)
}

// IsRunning checks if a process with the given PID is still alive.
func IsRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}
