package cmd

import (
	"fmt"

	"github.com/skarlsson/ws-manager/internal/config"
	"github.com/skarlsson/ws-manager/internal/git"
	"github.com/skarlsson/ws-manager/internal/kitty"
	"github.com/skarlsson/ws-manager/internal/monitor"
	"github.com/skarlsson/ws-manager/internal/state"
	"github.com/spf13/cobra"
)

// bringToFront moves a workspace to the work monitor, restoring the previously
// focused workspace to its home position (multi mode) or minimizing others (single mode).
func bringToFront(name string) error {
	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		return err
	}

	winID, err := kitty.PlatformWindowID(name)
	if err != nil {
		return err
	}

	mode := cfg.FocusMode
	if mode == "" {
		mode = "multi"
	}

	prev := state.LoadFocused()

	if mode == "single" {
		// Minimize all other active workspaces
		active, _ := state.ListActive()
		for _, other := range active {
			if other.Name == name || !kitty.IsRunning(other.KittyPID) {
				continue
			}
			otherWinID, err := kitty.PlatformWindowID(other.Name)
			if err == nil {
				monitor.MinimizeWindow(otherWinID)
			}
		}
	} else {
		// Multi mode: move previous workspace back to its home position
		if prev != "" && prev != name {
			prevSt, err := state.Load(prev)
			if err == nil && prevSt.Active && kitty.IsRunning(prevSt.KittyPID) {
				prevWinID, err := kitty.PlatformWindowID(prev)
				if err == nil && prevSt.HomeCaptured {
					monitor.MoveWindow(prevWinID, prevSt.HomeX, prevSt.HomeY)
				}
			}
		}
	}

	// Move target to work monitor
	if cfg.WorkMonitor != "" {
		mon, err := monitor.GetMonitor(cfg.WorkMonitor)
		if err == nil {
			monitor.MoveWindow(winID, mon.X, mon.Y)
		}
	}

	// Activate
	if err := monitor.ActivateWindow(winID); err != nil {
		return err
	}

	// Refresh window title with current branch
	refreshTitle(name)

	state.SaveFocused(name)
	return nil
}

func refreshTitle(name string) {
	ws, err := config.LoadWorkspace(name)
	if err != nil {
		return
	}
	title := fmt.Sprintf("ws: %s", name)
	if git.IsGitRepo(ws.Dir) {
		if branch, err := git.CurrentBranch(ws.Dir); err == nil {
			title = fmt.Sprintf("ws: %s [%s]", name, branch)
		}
	}
	kitty.SetTitle(name, title)
}

var rotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Cycle to the next active workspace on the work monitor",
	RunE: func(cmd *cobra.Command, args []string) error {
		active, err := state.ListActive()
		if err != nil {
			return fmt.Errorf("listing active workspaces: %w", err)
		}

		var running []state.WorkspaceState
		for _, st := range active {
			if kitty.IsRunning(st.KittyPID) {
				running = append(running, st)
			}
		}

		if len(running) == 0 {
			fmt.Println("No active workspaces to rotate.")
			return nil
		}

		current := state.LoadRotateIndex()
		next := (current + 1) % len(running)
		state.SaveRotateIndex(next)

		target := running[next]
		if err := bringToFront(target.Name); err != nil {
			return fmt.Errorf("focusing %q: %w", target.Name, err)
		}

		fmt.Printf("Rotated to workspace %q (%d/%d)\n", target.Name, next+1, len(running))
		return nil
	},
}

var focusCmd = &cobra.Command{
	Use:   "focus <workspace>",
	Short: "Bring a workspace to the work monitor, restore the previous one",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		st, err := state.Load(name)
		if err != nil || !st.Active || !kitty.IsRunning(st.KittyPID) {
			return fmt.Errorf("workspace %q is not running", name)
		}

		if err := bringToFront(name); err != nil {
			return fmt.Errorf("focusing %q: %w", name, err)
		}

		fmt.Printf("Focused workspace %q\n", name)
		return nil
	},
}

var unfocusCmd = &cobra.Command{
	Use:   "unfocus",
	Short: "Send the currently focused workspace back to its home position",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := state.LoadFocused()
		if name == "" {
			fmt.Println("No workspace currently focused.")
			return nil
		}

		st, err := state.Load(name)
		if err != nil || !st.Active || !kitty.IsRunning(st.KittyPID) {
			state.SaveFocused("")
			return fmt.Errorf("workspace %q is no longer running", name)
		}

		if !st.HomeCaptured {
			fmt.Printf("Workspace %q has no saved home position. Run 'ws capture' first.\n", name)
			return nil
		}

		winID, err := kitty.PlatformWindowID(name)
		if err != nil {
			return err
		}

		if err := monitor.MoveWindow(winID, st.HomeX, st.HomeY); err != nil {
			return fmt.Errorf("moving %q home: %w", name, err)
		}

		state.SaveFocused("")
		fmt.Printf("Sent workspace %q back home\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rotateCmd)
	rootCmd.AddCommand(focusCmd)
	rootCmd.AddCommand(unfocusCmd)
}
