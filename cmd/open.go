package cmd

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/skarlsson/ws-manager/internal/config"
	"github.com/skarlsson/ws-manager/internal/kitty"
	"github.com/skarlsson/ws-manager/internal/state"
	"github.com/skarlsson/ws-manager/internal/zellij"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open <workspace>",
	Short: "Open a workspace in a new kitty window with zellij",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		ws, err := config.LoadWorkspace(name)
		if err != nil {
			return fmt.Errorf("workspace %q not found: %w", name, err)
		}

		// Check if already active
		st, _ := state.Load(name)
		if st.Active && kitty.IsRunning(st.KittyPID) {
			fmt.Printf("Workspace %q is already open (PID %d)\n", name, st.KittyPID)
			return nil
		}

		// Generate layout
		layoutPath, err := zellij.GenerateLayout(ws)
		if err != nil {
			return fmt.Errorf("generating layout: %w", err)
		}

		// Clean up any dead zellij session with the same name
		session := zellij.SessionName(name)
		zellij.CleanupSession(session)

		// Launch kitty
		title := fmt.Sprintf("ws: %s", name)
		pid, err := kitty.Launch(name, ws.Dir, title)
		if err != nil {
			return fmt.Errorf("launching kitty: %w", err)
		}

		// Wait for kitty socket to be ready
		socket := kitty.SocketPath(name)
		zellijCmd := zellij.LaunchCommand(session, layoutPath, ws.Dir)

		// Wait for kitty socket to become available
		if err := waitForSocket(socket, 5*time.Second); err != nil {
			fmt.Printf("Warning: kitty socket not ready: %v\n", err)
		}

		if err := kitty.SendText(socket, zellijCmd); err != nil {
			fmt.Printf("Warning: could not auto-start zellij: %v\n", err)
			fmt.Println("Start it manually with: zellij --session", session, "--layout", layoutPath)
		}

		// Save state
		st = state.WorkspaceState{
			Name:          name,
			KittyPID:      pid,
			ZellijSession: session,
			Active:        true,
		}
		if err := state.Save(st); err != nil {
			return fmt.Errorf("saving state: %w", err)
		}

		fmt.Printf("Opened workspace %q (kitty PID %d, zellij session %q)\n", name, pid, session)
		return nil
	},
}

func waitForSocket(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			// Socket file exists, try connecting
			conn, err := net.Dial("unix", path)
			if err == nil {
				conn.Close()
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for %s", path)
}

func init() {
	rootCmd.AddCommand(openCmd)
}
