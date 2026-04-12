package agent

import (
	"context"
	"os"
	"time"

	"jane/pkg/config"
	"jane/pkg/logger"
	"jane/pkg/runtimepaths"
	"jane/pkg/state"
	"jane/pkg/tools"
)

func configFileModTime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

func (al *AgentLoop) refreshConfigPath() {
	if al.configPath == "" {
		al.configPath = runtimepaths.ConfigPath()
	}
}

func (al *AgentLoop) reloadRuntimeConfigIfChanged(ctx context.Context) error {
	al.refreshConfigPath()
	if al.configPath == "" {
		return nil
	}

	currentModTime := configFileModTime(al.configPath)
	if !currentModTime.After(al.configModTime) {
		return nil
	}

	al.reloadMu.Lock()
	defer al.reloadMu.Unlock()

	currentModTime = configFileModTime(al.configPath)
	if !currentModTime.After(al.configModTime) {
		return nil
	}

	nextCfg, err := config.LoadConfig(al.configPath)
	if err != nil {
		return err
	}

	nextRegistry := NewAgentRegistry(nextCfg, al.provider)
	registerSharedTools(nextCfg, al.bus, nextRegistry, al.provider, al)

	if al.mediaStore != nil {
		nextRegistry.ForEachTool("send_file", func(t tools.Tool) {
			if sendFileTool, ok := t.(*tools.SendFileTool); ok {
				sendFileTool.SetMediaStore(al.mediaStore)
			}
		})
	}

	if manager := al.mcp.takeManager(); manager != nil {
		_ = manager.Close()
	}
	al.mcp = mcpRuntime{}

	if oldRegistry := al.registry; oldRegistry != nil {
		oldRegistry.Close()
	}

	al.cfg = nextCfg
	al.registry = nextRegistry
	al.configModTime = currentModTime

	if defaultAgent := al.registry.GetDefaultAgent(); defaultAgent != nil {
		al.state = state.NewManager(defaultAgent.Workspace)
	}

	logger.InfoCF("agent", "Reloaded runtime config", map[string]any{
		"config_path": al.configPath,
		"model":       nextCfg.Agents.Defaults.GetModelName(),
	})

	if nextCfg.Tools.IsToolEnabled("mcp") {
		return al.ensureMCPInitialized(ctx)
	}
	return nil
}
