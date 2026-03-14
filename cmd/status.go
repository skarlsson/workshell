package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/skarlsson/ws-manager/internal/config"
	"github.com/skarlsson/ws-manager/internal/git"
	"github.com/skarlsson/ws-manager/internal/kitty"
	"github.com/skarlsson/ws-manager/internal/process"
	"github.com/skarlsson/ws-manager/internal/state"
	"github.com/spf13/cobra"
)

type WorkspaceStatus struct {
	Name       string `json:"name"`
	Dir        string `json:"dir"`
	Branch     string `json:"branch"`
	Task       string `json:"task,omitempty"`
	Active     bool   `json:"active"`
	Focused    bool   `json:"focused"`
	Claude     bool   `json:"claude"`
	ClaudeCPU  int64  `json:"claude_cpu_secs"`
	ClaudeTime string `json:"claude_cpu_time,omitempty"`
}

var statusCmd = &cobra.Command{
	Use:   "status [workspace]",
	Short: "Output workspace status as JSON (for integrations)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		focused := state.LoadFocused()

		if len(args) == 1 {
			// Single workspace
			s, err := getWorkspaceStatus(args[0], focused)
			if err != nil {
				return err
			}
			return outputJSON(s)
		}

		// All workspaces
		workspaces, err := config.ListWorkspaces()
		if err != nil {
			return fmt.Errorf("listing workspaces: %w", err)
		}

		var statuses []WorkspaceStatus
		for _, ws := range workspaces {
			s, _ := getWorkspaceStatus(ws.Name, focused)
			statuses = append(statuses, s)
		}
		return outputJSON(statuses)
	},
}

func getWorkspaceStatus(name, focused string) (WorkspaceStatus, error) {
	ws, err := config.LoadWorkspace(name)
	if err != nil {
		return WorkspaceStatus{}, fmt.Errorf("workspace %q not found: %w", name, err)
	}

	st, _ := state.Load(name)
	active := st.Active && kitty.IsRunning(st.KittyPID)

	// Read actual branch from disk, fall back to config
	branch := ws.CurrentBranch
	if git.IsGitRepo(ws.Dir) {
		if b, err := git.CurrentBranch(ws.Dir); err == nil {
			branch = b
		}
	}

	s := WorkspaceStatus{
		Name:    ws.Name,
		Dir:     ws.Dir,
		Branch:  branch,
		Task:    ws.CurrentTask,
		Active:  active,
		Focused: name == focused,
	}

	if active {
		ci := process.GetClaudeInfo(st.ZellijSession)
		s.Claude = ci.Running
		s.ClaudeCPU = ci.CPUSecs
		s.ClaudeTime = ci.CPUTime
	}

	return s, nil
}

func outputJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
