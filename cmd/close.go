package cmd

import (
	"fmt"

	"github.com/skarlsson/ws-manager/internal/kitty"
	"github.com/skarlsson/ws-manager/internal/state"
	"github.com/skarlsson/ws-manager/internal/zellij"
	"github.com/spf13/cobra"
)

var closeCmd = &cobra.Command{
	Use:   "close <workspace>",
	Short: "Close a workspace session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		st, err := state.Load(name)
		if err != nil {
			return fmt.Errorf("loading state for %q: %w", name, err)
		}
		if !st.Active {
			return fmt.Errorf("workspace %q is not open", name)
		}

		// Kill zellij session
		if st.ZellijSession != "" {
			if err := zellij.KillSession(st.ZellijSession); err != nil {
				fmt.Printf("Warning: could not kill zellij session %q: %v\n", st.ZellijSession, err)
			}
		}

		// Kill kitty process
		if st.KittyPID > 0 {
			if err := kitty.KillProcess(st.KittyPID); err != nil {
				fmt.Printf("Warning: could not kill kitty process %d: %v\n", st.KittyPID, err)
			}
		}

		// Remove state
		if err := state.Remove(name); err != nil {
			fmt.Printf("Warning: could not remove state file: %v\n", err)
		}

		fmt.Printf("Closed workspace %q\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(closeCmd)
}
