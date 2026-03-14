package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/skarlsson/ws-manager/internal/config"
	"github.com/skarlsson/ws-manager/internal/git"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new workspace interactively",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.EnsureDirs(); err != nil {
			return err
		}

		reader := bufio.NewReader(os.Stdin)
		prompt := func(label, defaultVal string) string {
			if defaultVal != "" {
				fmt.Printf("%s [%s]: ", label, defaultVal)
			} else {
				fmt.Printf("%s: ", label)
			}
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if input == "" {
				return defaultVal
			}
			return input
		}

		// 1. Name
		name := prompt("Workspace name", "")
		if name == "" {
			return fmt.Errorf("workspace name is required")
		}
		if config.WorkspaceExists(name) {
			return fmt.Errorf("workspace %q already exists", name)
		}

		// 2. Directory
		cwd, _ := os.Getwd()
		defaultDir := filepath.Join(cwd, name)
		dir := prompt("Directory", defaultDir)
		dir, _ = filepath.Abs(dir)

		// 3. Ensure directory exists — clone if needed
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			repo := prompt("Directory doesn't exist. Git repo URL to clone (empty to create empty dir)", "")
			if repo != "" {
				fmt.Printf("Cloning %s into %s...\n", repo, dir)
				clone := exec.Command("git", "clone", "--recursive", repo, dir)
				clone.Stdout = os.Stdout
				clone.Stderr = os.Stderr
				if err := clone.Run(); err != nil {
					return fmt.Errorf("cloning repo: %w", err)
				}
			} else {
				fmt.Printf("Creating directory %s...\n", dir)
				if err := os.MkdirAll(dir, 0755); err != nil {
					return fmt.Errorf("creating directory: %w", err)
				}
			}
		} else {
			fmt.Printf("Using existing directory %s\n", dir)
		}

		// 4. Layout
		layout := prompt("Layout", "default")

		// 5. Auto-start claude
		autoClaudeStr := prompt("Auto-start claude in left pane? (y/n)", "y")
		autoClaude := strings.ToLower(autoClaudeStr) == "y" || strings.ToLower(autoClaudeStr) == "yes"

		// 6. Setup commands
		fmt.Println("Setup commands (one per line, empty line to finish):")
		var setupCmds []string
		for {
			line, _ := reader.ReadString('\n')
			line = strings.TrimSpace(line)
			if line == "" {
				break
			}
			setupCmds = append(setupCmds, line)
		}

		// Detect default branch from the repo
		defaultBranch := git.DefaultBranch(dir)

		ws := config.Workspace{
			Name:          name,
			Dir:           dir,
			DefaultBranch: defaultBranch,
			Layout:        layout,
			AutoClaude:    autoClaude,
			SetupCommands: setupCmds,
		}

		if err := config.SaveWorkspace(ws); err != nil {
			return fmt.Errorf("saving workspace: %w", err)
		}

		fmt.Printf("\nWorkspace %q created!\n", name)
		fmt.Printf("Config: %s\n", config.WorkspacePath(name))
		fmt.Printf("Open it with: ws open %s\n", name)

		// Run setup commands if any
		if len(setupCmds) > 0 {
			runSetup := prompt("Run setup commands now? (y/n)", "y")
			if strings.ToLower(runSetup) == "y" || strings.ToLower(runSetup) == "yes" {
				for _, c := range setupCmds {
					fmt.Printf("Running: %s\n", c)
					setup := exec.Command("bash", "-c", c)
					setup.Dir = dir
					setup.Stdout = os.Stdout
					setup.Stderr = os.Stderr
					if err := setup.Run(); err != nil {
						fmt.Printf("Warning: command failed: %v\n", err)
					}
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
