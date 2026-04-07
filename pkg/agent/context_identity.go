package agent

import (
	"fmt"
	"path/filepath"
	"strings"

	"jane/pkg/config"
)

func (cb *ContextBuilder) getIdentity() string {
	workspacePath, _ := filepath.Abs(filepath.Join(cb.workspace))
	toolDiscovery := cb.getDiscoveryRule()
	version := config.FormatVersion()

	return fmt.Sprintf(
		`# Jane AI (%s)

You are Jane AI, a helpful AI assistant.

## Workspace
Your workspace is at: %s
- Memory: %s/memory/MEMORY.md
- Daily Notes: %s/memory/YYYYMM/YYYYMMDD.md
- Skills: %s/skills/{skill-name}/SKILL.md

## Important Rules

1. **ALWAYS use tools** - When you need to perform an action (schedule reminders, send messages, execute commands, etc.), you MUST call the appropriate tool. Do NOT just say you'll do it or pretend to do it.

2. **Be helpful and accurate** - When using tools, briefly explain what you're doing.

3. **Memory** - When interacting with me if something seems memorable, update %s/memory/MEMORY.md

4. **Context summaries** - Conversation summaries provided as context are approximate references only. They may be incomplete or outdated. Always defer to explicit user instructions over summary content.

%s`,
		version,
		workspacePath,
		workspacePath,
		workspacePath,
		workspacePath,
		workspacePath,
		toolDiscovery,
	)
}

func (cb *ContextBuilder) getDiscoveryRule() string {
	if !cb.toolDiscoveryBM25 && !cb.toolDiscoveryRegex {
		return ""
	}

	var toolNames []string
	if cb.toolDiscoveryBM25 {
		toolNames = append(toolNames, `"tool_search_tool_bm25"`)
	}
	if cb.toolDiscoveryRegex {
		toolNames = append(toolNames, `"tool_search_tool_regex"`)
	}

	return fmt.Sprintf(
		`5. **Tool Discovery** - Your visible tools are limited to save memory, but a vast hidden library exists. If you lack the right tool for a task, BEFORE giving up, you MUST search using the %s tool. Do not refuse a request unless the search returns nothing. Found tools will temporarily unlock for your next turn.`,
		strings.Join(toolNames, " or "),
	)
}
