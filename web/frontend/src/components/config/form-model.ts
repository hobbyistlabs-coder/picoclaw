export type JsonRecord = Record<string, unknown>

export interface PersonaForm {
  key: string
  id: string
  name: string
  isDefault: boolean
  workspace: string
  primaryModel: string
  fallbackModelsText: string
  systemPrompt: string
  skillsText: string
  mcpServersText: string
  allowedAgentsText: string
  subagentModel: string
  subagentFallbacksText: string
}

export interface CoreConfigForm {
  workspace: string
  restrictToWorkspace: boolean
  allowRemote: boolean
  maxTokens: string
  maxToolIterations: string
  summarizeMessageThreshold: string
  summarizeTokenPercent: string
  dmScope: string
  heartbeatEnabled: boolean
  heartbeatInterval: string
  devicesEnabled: boolean
  monitorUSB: boolean
  personas: PersonaForm[]
}

export interface LauncherForm {
  port: string
  publicAccess: boolean
  allowedCIDRsText: string
}

export const DM_SCOPE_OPTIONS = [
  {
    value: "per-channel-peer",
    labelKey: "pages.config.session_scope_per_channel_peer",
    labelDefault: "Per Channel + Peer",
    descKey: "pages.config.session_scope_per_channel_peer_desc",
    descDefault: "Separate context for each user in each channel.",
  },
  {
    value: "per-channel",
    labelKey: "pages.config.session_scope_per_channel",
    labelDefault: "Per Channel",
    descKey: "pages.config.session_scope_per_channel_desc",
    descDefault: "One shared context per channel.",
  },
  {
    value: "per-peer",
    labelKey: "pages.config.session_scope_per_peer",
    labelDefault: "Per Peer",
    descKey: "pages.config.session_scope_per_peer_desc",
    descDefault: "One context per user across channels.",
  },
  {
    value: "global",
    labelKey: "pages.config.session_scope_global",
    labelDefault: "Global",
    descKey: "pages.config.session_scope_global_desc",
    descDefault: "All messages share one global context.",
  },
] as const

export const EMPTY_FORM: CoreConfigForm = {
  workspace: "",
  restrictToWorkspace: true,
  allowRemote: true,
  maxTokens: "32768",
  maxToolIterations: "50",
  summarizeMessageThreshold: "20",
  summarizeTokenPercent: "75",
  dmScope: "per-channel-peer",
  heartbeatEnabled: true,
  heartbeatInterval: "30",
  devicesEnabled: false,
  monitorUSB: true,
  personas: [],
}

export const EMPTY_LAUNCHER_FORM: LauncherForm = {
  port: "18800",
  publicAccess: false,
  allowedCIDRsText: "",
}

function asRecord(value: unknown): JsonRecord {
  if (value && typeof value === "object" && !Array.isArray(value)) {
    return value as JsonRecord
  }
  return {}
}

function asString(value: unknown): string {
  return typeof value === "string" ? value : ""
}

function asArray(value: unknown): unknown[] {
  return Array.isArray(value) ? value : []
}

function asBool(value: unknown): boolean {
  return value === true
}

function asNumberString(value: unknown, fallback: string): string {
  if (typeof value === "number" && Number.isFinite(value)) {
    return String(value)
  }
  if (typeof value === "string" && value.trim() !== "") {
    return value
  }
  return fallback
}

function asStringArray(value: unknown): string[] {
  return asArray(value).filter(
    (item): item is string => typeof item === "string",
  )
}

let personaSequence = 0

function nextPersonaKey() {
  personaSequence += 1
  return `persona-${personaSequence}`
}

export function createEmptyPersona(): PersonaForm {
  return {
    key: nextPersonaKey(),
    id: "",
    name: "",
    isDefault: false,
    workspace: "",
    primaryModel: "",
    fallbackModelsText: "",
    systemPrompt: "",
    skillsText: "",
    mcpServersText: "",
    allowedAgentsText: "",
    subagentModel: "",
    subagentFallbacksText: "",
  }
}

function buildPersonaForm(value: unknown): PersonaForm {
  const record = asRecord(value)
  const model = asRecord(record.model)
  const subagents = asRecord(record.subagents)
  const subagentModel = asRecord(subagents.model)

  return {
    key: nextPersonaKey(),
    id: asString(record.id),
    name: asString(record.name),
    isDefault: asBool(record.default),
    workspace: asString(record.workspace),
    primaryModel: asString(model.primary || record.model),
    fallbackModelsText: asStringArray(model.fallbacks).join("\n"),
    systemPrompt: asString(record.system_prompt),
    skillsText: asStringArray(record.skills).join("\n"),
    mcpServersText: asStringArray(record.mcp_servers).join("\n"),
    allowedAgentsText: asStringArray(subagents.allow_agents).join("\n"),
    subagentModel: asString(subagentModel.primary || subagents.model),
    subagentFallbacksText: asStringArray(subagentModel.fallbacks).join("\n"),
  }
}

export function buildFormFromConfig(config: unknown): CoreConfigForm {
  const root = asRecord(config)
  const agents = asRecord(root.agents)
  const defaults = asRecord(agents.defaults)
  const personas = asArray(agents.list).map(buildPersonaForm)
  const session = asRecord(root.session)
  const heartbeat = asRecord(root.heartbeat)
  const devices = asRecord(root.devices)
  const tools = asRecord(root.tools)
  const exec = asRecord(tools.exec)

  return {
    workspace: asString(defaults.workspace) || EMPTY_FORM.workspace,
    restrictToWorkspace:
      defaults.restrict_to_workspace === undefined
        ? EMPTY_FORM.restrictToWorkspace
        : asBool(defaults.restrict_to_workspace),
    allowRemote:
      exec.allow_remote === undefined
        ? EMPTY_FORM.allowRemote
        : asBool(exec.allow_remote),
    maxTokens: asNumberString(defaults.max_tokens, EMPTY_FORM.maxTokens),
    maxToolIterations: asNumberString(
      defaults.max_tool_iterations,
      EMPTY_FORM.maxToolIterations,
    ),
    summarizeMessageThreshold: asNumberString(
      defaults.summarize_message_threshold,
      EMPTY_FORM.summarizeMessageThreshold,
    ),
    summarizeTokenPercent: asNumberString(
      defaults.summarize_token_percent,
      EMPTY_FORM.summarizeTokenPercent,
    ),
    dmScope: asString(session.dm_scope) || EMPTY_FORM.dmScope,
    heartbeatEnabled:
      heartbeat.enabled === undefined
        ? EMPTY_FORM.heartbeatEnabled
        : asBool(heartbeat.enabled),
    heartbeatInterval: asNumberString(
      heartbeat.interval,
      EMPTY_FORM.heartbeatInterval,
    ),
    devicesEnabled:
      devices.enabled === undefined
        ? EMPTY_FORM.devicesEnabled
        : asBool(devices.enabled),
    monitorUSB:
      devices.monitor_usb === undefined
        ? EMPTY_FORM.monitorUSB
        : asBool(devices.monitor_usb),
    personas,
  }
}

export function parseIntField(
  rawValue: string,
  label: string,
  options: { min?: number; max?: number } = {},
): number {
  const value = Number(rawValue)
  if (!Number.isInteger(value)) {
    throw new Error(`${label} must be an integer.`)
  }
  if (options.min !== undefined && value < options.min) {
    throw new Error(`${label} must be >= ${options.min}.`)
  }
  if (options.max !== undefined && value > options.max) {
    throw new Error(`${label} must be <= ${options.max}.`)
  }
  return value
}

export function parseCIDRText(raw: string): string[] {
  return parseListText(raw)
}

export function parseListText(raw: string): string[] {
  if (!raw.trim()) {
    return []
  }
  return raw
    .split(/[\n,]/)
    .map((v) => v.trim())
    .filter((v) => v.length > 0)
}
