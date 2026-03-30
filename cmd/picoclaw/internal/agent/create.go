package agent

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"jane/cmd/picoclaw/internal"
	"jane/pkg/config"
	"jane/pkg/runtimepaths"
)

func NewCreateCommand() *cobra.Command {
	var (
		name        string
		workspace   string
		sysPrompt   string
		model       string
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new agent persona",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !interactive && name == "" {
				return fmt.Errorf("name is required when not in interactive mode")
			}
			return createAgentCmd(name, workspace, sysPrompt, model, interactive)
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Agent name")
	cmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Workspace path")
	cmd.Flags().StringVarP(&sysPrompt, "system-prompt", "p", "", "System prompt / instructions")
	cmd.Flags().StringVarP(&model, "model", "m", "", "Model configuration (primary)")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode")

	return cmd
}

func createAgentCmd(name, workspace, sysPrompt, model string, interactive bool) error {
	cfg, err := internal.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	if interactive {
		fmt.Printf("%s Creating new agent persona...\n\n", internal.Logo)

		reader := bufio.NewReader(os.Stdin)

		if name == "" {
			fmt.Print("Agent Name: ")
			nameInput, _ := reader.ReadString('\n')
			nameInput = strings.TrimSpace(nameInput)
			if nameInput != "" {
				name = nameInput
			}
		}

		if workspace == "" {
			fmt.Printf(
				"Workspace path (default: %s/workspace/%s): ",
				runtimepaths.HomeDir(),
				strings.ToLower(strings.ReplaceAll(name, " ", "_")),
			)
			workspaceInput, _ := reader.ReadString('\n')
			workspaceInput = strings.TrimSpace(workspaceInput)
			if workspaceInput != "" {
				workspace = workspaceInput
			}
		}

		if sysPrompt == "" {
			fmt.Print("System Prompt (optional): ")
			sysPromptInput, _ := reader.ReadString('\n')
			sysPromptInput = strings.TrimSpace(sysPromptInput)
			if sysPromptInput != "" {
				sysPrompt = sysPromptInput
			}
		}
	}

	if name == "" {
		return fmt.Errorf("agent name is required")
	}

	id := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	id = strings.ReplaceAll(id, "_", "-")

	if id == "" {
		id = uuid.New().String()[:8]
	}

	if workspace == "" {
		workspace = filepath.Join(
			runtimepaths.HomeDir(),
			"workspace",
			strings.ToLower(strings.ReplaceAll(name, " ", "_")),
		)
	}

	var modelCfg *config.AgentModelConfig
	if model != "" {
		modelCfg = &config.AgentModelConfig{Primary: model}
	}

	newAgent := config.AgentConfig{
		ID:           id,
		Name:         name,
		Workspace:    workspace,
		SystemPrompt: sysPrompt,
		Model:        modelCfg,
	}

	cfg.Agents.List = append(cfg.Agents.List, newAgent)

	if err := config.SaveConfig(internal.GetConfigPath(), cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("\n✅ Successfully created agent persona '%s' with ID '%s'\n", name, id)
	return nil
}
