package cmd

import (
	"fmt"

	"github.com/skarlsson/ws-manager/internal/kitty"
	"github.com/skarlsson/ws-manager/internal/monitor"
	"github.com/skarlsson/ws-manager/internal/state"
	"github.com/spf13/cobra"
)

func captureHome(name string) error {
	st, err := state.Load(name)
	if err != nil {
		return fmt.Errorf("loading state for %q: %w", name, err)
	}
	if !st.Active || !kitty.IsRunning(st.KittyPID) {
		return fmt.Errorf("workspace %q is not running", name)
	}

	winID, err := kitty.PlatformWindowID(name)
	if err != nil {
		return fmt.Errorf("getting window ID for %q: %w", name, err)
	}

	x, y, err := monitor.GetWindowPosition(winID)
	if err != nil {
		return fmt.Errorf("getting position for %q: %w", name, err)
	}

	st.HomeX = x
	st.HomeY = y
	st.HomeCaptured = true
	if err := state.Save(st); err != nil {
		return fmt.Errorf("saving state for %q: %w", name, err)
	}

	fmt.Printf("  %s: home set to (%d, %d)\n", name, x, y)
	return nil
}

var captureCmd = &cobra.Command{
	Use:   "capture [workspace]",
	Short: "Snapshot current window positions as home positions",
	Long:  "Captures the current position of active workspace windows and saves them as home positions. Without arguments, captures all active workspaces.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			return captureHome(args[0])
		}

		active, err := state.ListActive()
		if err != nil {
			return fmt.Errorf("listing active workspaces: %w", err)
		}

		if len(active) == 0 {
			fmt.Println("No active workspaces.")
			return nil
		}

		var captured int
		for _, ws := range active {
			if err := captureHome(ws.Name); err != nil {
				fmt.Printf("  %s: skipped (%v)\n", ws.Name, err)
				continue
			}
			captured++
		}
		fmt.Printf("Captured home positions for %d workspace(s).\n", captured)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(captureCmd)
}
