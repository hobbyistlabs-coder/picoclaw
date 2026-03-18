package agent

import (
	"fmt"

	"github.com/spf13/cobra"

	"jane/cmd/picoclaw/internal"
	"jane/pkg/config"
)

func NewAssignMCPCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assign-mcp <agent-id> <mcp-name>",
		Short: "Assign an MCP server to an agent persona",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return assignMCPCmd(args[0], args[1])
		},
	}
	return cmd
}

func assignMCPCmd(agentID, mcpName string) error {
	cfg, err := internal.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	found := false
	for i, a := range cfg.Agents.List {
		if a.ID == agentID {
			found = true

			// Check if already assigned
			for _, m := range a.MCPServers {
				if m == mcpName {
					fmt.Printf("⚠️ MCP server '%s' is already assigned to agent '%s'\n", mcpName, agentID)
					return nil
				}
			}

			cfg.Agents.List[i].MCPServers = append(cfg.Agents.List[i].MCPServers, mcpName)
			break
		}
	}

	if !found {
		return fmt.Errorf("agent with ID '%s' not found", agentID)
	}

	if err := config.SaveConfig(internal.GetConfigPath(), cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Successfully assigned MCP server '%s' to agent '%s'\n", mcpName, agentID)
	return nil
}

func NewRemoveMCPCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-mcp <agent-id> <mcp-name>",
		Short: "Remove an MCP server from an agent persona",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return removeMCPCmd(args[0], args[1])
		},
	}
	return cmd
}

func removeMCPCmd(agentID, mcpName string) error {
	cfg, err := internal.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	found := false
	for i, a := range cfg.Agents.List {
		if a.ID == agentID {
			found = true

			newMCPs := make([]string, 0, len(a.MCPServers))
			removed := false
			for _, m := range a.MCPServers {
				if m == mcpName {
					removed = true
					continue
				}
				newMCPs = append(newMCPs, m)
			}

			if !removed {
				fmt.Printf("⚠️ MCP server '%s' is not assigned to agent '%s'\n", mcpName, agentID)
				return nil
			}

			cfg.Agents.List[i].MCPServers = newMCPs
			break
		}
	}

	if !found {
		return fmt.Errorf("agent with ID '%s' not found", agentID)
	}

	if err := config.SaveConfig(internal.GetConfigPath(), cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✅ Successfully removed MCP server '%s' from agent '%s'\n", mcpName, agentID)
	return nil
}
