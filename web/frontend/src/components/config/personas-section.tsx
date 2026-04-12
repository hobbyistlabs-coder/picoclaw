import { IconChevronRight, IconPlus, IconTrash } from "@tabler/icons-react"
import { useQuery } from "@tanstack/react-query"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { getPersonaPromptHistory } from "@/api/prompt-history"
import type { PersonaForm } from "@/components/config/form-model"
import { MarkdownPromptEditor } from "@/components/config/markdown-prompt-editor"
import { Field, SwitchCardField } from "@/components/shared-form"

// Reusable styling for form inputs to keep the JSX clean
const glassInputBase =
  "w-full rounded-xl border border-white/10 bg-black/20 px-4 py-2.5 text-sm text-white placeholder-white/30 transition-all focus:border-indigo-500/50 focus:bg-white/[0.05] focus:ring-1 focus:ring-indigo-500/50 focus:outline-none disabled:cursor-not-allowed disabled:opacity-50"

const glassTextareaBase =
  "min-h-[88px] w-full resize-y rounded-xl border border-white/10 bg-black/20 px-4 py-3 text-sm text-white placeholder-white/30 transition-all focus:border-indigo-500/50 focus:bg-white/[0.05] focus:ring-1 focus:ring-indigo-500/50 focus:outline-none disabled:cursor-not-allowed disabled:opacity-50"

function PersonaPromptField({
  persona,
  disabled,
  onChange,
}: {
  persona: PersonaForm
  disabled?: boolean
  onChange: (value: string) => void
}) {
  const { t } = useTranslation()
  const personaID = persona.id.trim()
  const { data, isLoading } = useQuery({
    queryKey: ["config", "persona", personaID, "system-prompt", "history"],
    queryFn: () => getPersonaPromptHistory(personaID),
    enabled: personaID.length > 0,
  })

  return (
    <Field
      label={t("pages.config.persona_prompt")}
      hint={t("pages.config.persona_prompt_hint")}
    >
      <div className="mt-2">
        <MarkdownPromptEditor
          value={persona.systemPrompt}
          onChange={onChange}
          placeholder={t("pages.config.workspace_files.placeholder")}
          disabled={disabled}
          revisions={data?.revisions ?? []}
          historyLoading={isLoading}
        />
      </div>
    </Field>
  )
}

interface PersonasSectionProps {
  personas: PersonaForm[]
  sectionId?: string
  disabled?: boolean
  onAdd: () => void
  onRemove: (key: string) => void
  onChange: <K extends keyof PersonaForm>(
    key: string,
    field: K,
    value: PersonaForm[K],
  ) => void
}

export function PersonasSection({
  personas,
  sectionId,
  disabled,
  onAdd,
  onRemove,
  onChange,
}: PersonasSectionProps) {
  const { t } = useTranslation()

  const [expandedPersonas, setExpandedPersonas] = useState<
    Record<string, boolean>
  >({})

  const togglePersona = (key: string) => {
    setExpandedPersonas((prev) => ({
      ...prev,
      [key]: !prev[key],
    }))
  }

  return (
    <div
      id={sectionId}
      className="rounded-3xl border border-white/10 bg-white/[0.02] p-6 shadow-2xl backdrop-blur-xl"
    >
      {/* Section Header */}
      <div className="mb-6 flex flex-col items-start justify-between gap-4 border-b border-white/10 pb-6 sm:flex-row sm:items-center">
        <div className="space-y-1.5">
          <h3 className="text-xl font-semibold tracking-tight text-white">
            {t("pages.config.sections.personas")}
          </h3>
          <p className="text-sm text-white/50">
            {t("pages.config.personas_description")}
          </p>
        </div>

        <button
          type="button"
          onClick={onAdd}
          disabled={disabled}
          className="group flex items-center gap-2 rounded-full border border-indigo-500/30 bg-indigo-500/20 px-5 py-2.5 text-xs font-bold tracking-wide text-indigo-50 uppercase shadow-inner transition-all hover:bg-indigo-500/30 hover:shadow-[0_0_15px_rgba(99,102,241,0.2)] active:scale-95 disabled:pointer-events-none disabled:opacity-30"
        >
          <IconPlus className="size-4 text-indigo-300 transition-transform duration-300 group-hover:rotate-90" />
          {t("pages.config.add_persona")}
        </button>
      </div>

      {/* Personas List */}
      <div className="space-y-4">
        {personas.length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-3 rounded-2xl border-2 border-dashed border-white/10 bg-white/[0.01] py-12 text-sm text-white/40">
            <IconPlus className="size-8 text-white/20" />
            {t("pages.config.personas_empty")}
          </div>
        ) : (
          personas.map((persona, index) => {
            const isExpanded = expandedPersonas[persona.key]
            const displayName =
              persona.name.trim() ||
              persona.id.trim() ||
              `${t("pages.config.persona")} ${index + 1}`

            return (
              <div
                key={persona.key}
                id={`persona-${persona.id.trim() || persona.key}`}
                className="overflow-hidden rounded-2xl border border-white/10 bg-white/[0.02] transition-all duration-300 focus-within:border-indigo-500/50 hover:border-white/20"
              >
                {/* Collapsible Header */}
                <div
                  className={`flex items-center justify-between gap-3 p-3 pl-4 transition-colors ${isExpanded ? "bg-white/[0.02]" : ""}`}
                >
                  <button
                    type="button"
                    onClick={() => togglePersona(persona.key)}
                    className="group flex flex-1 items-center gap-4 text-left focus:outline-none"
                  >
                    <div
                      className={`flex size-7 flex-shrink-0 items-center justify-center rounded-full bg-white/[0.05] transition-all duration-300 group-hover:bg-white/[0.1] ${isExpanded ? "rotate-90 bg-white/[0.1]" : ""}`}
                    >
                      <IconChevronRight className="size-4 text-white/60" />
                    </div>
                    <div>
                      <h4 className="text-[15px] font-medium text-white/90">
                        {displayName}
                      </h4>
                      <p className="mt-0.5 text-xs text-white/40">
                        {t("pages.config.persona_card_hint")}
                      </p>
                    </div>
                  </button>

                  <button
                    type="button"
                    onClick={() => onRemove(persona.key)}
                    disabled={disabled}
                    className="flex size-9 flex-shrink-0 items-center justify-center rounded-full text-white/40 transition-all hover:bg-red-500/20 hover:text-red-400 focus:ring-2 focus:ring-red-500/50 focus:outline-none disabled:opacity-20"
                    title={t("common.remove")}
                  >
                    <IconTrash className="size-4.5" />
                  </button>
                </div>

                {/* Collapsible Body */}
                {isExpanded && (
                  <div className="animate-in fade-in slide-in-from-top-2 border-t border-white/5 p-6 duration-300">
                    {/* Section: Basic Settings */}
                    <div className="space-y-5">
                      <div className="grid gap-5 sm:grid-cols-2">
                        <Field label={t("pages.config.persona_id")}>
                          <input
                            type="text"
                            value={persona.id}
                            disabled={disabled}
                            placeholder="coding"
                            onChange={(e) =>
                              onChange(persona.key, "id", e.target.value)
                            }
                            className={glassInputBase}
                          />
                        </Field>

                        <Field label={t("pages.config.persona_name")}>
                          <input
                            type="text"
                            value={persona.name}
                            disabled={disabled}
                            placeholder="Coding Persona"
                            onChange={(e) =>
                              onChange(persona.key, "name", e.target.value)
                            }
                            className={glassInputBase}
                          />
                        </Field>
                      </div>

                      <div className="grid gap-5 sm:grid-cols-2">
                        <Field
                          label={t("pages.config.persona_workspace")}
                          hint={t("pages.config.persona_workspace_hint")}
                        >
                          <input
                            type="text"
                            value={persona.workspace}
                            disabled={disabled}
                            placeholder="/root/.picoclaw/workspace"
                            onChange={(e) =>
                              onChange(persona.key, "workspace", e.target.value)
                            }
                            className={glassInputBase}
                          />
                        </Field>

                        <div className="flex items-center pt-6">
                          <SwitchCardField
                            label={t("pages.config.persona_default")}
                            hint={t("pages.config.persona_default_hint")}
                            checked={persona.isDefault}
                            onCheckedChange={(checked) =>
                              onChange(persona.key, "isDefault", checked)
                            }
                            disabled={disabled}
                          />
                        </div>
                      </div>
                    </div>

                    <div className="my-8 h-px w-full bg-gradient-to-r from-transparent via-white/10 to-transparent" />

                    {/* Section: Models Configuration */}
                    <div className="space-y-5">
                      <h5 className="mb-4 text-xs font-semibold tracking-wider text-white/40 uppercase">
                        Model Configuration
                      </h5>

                      <div className="grid gap-5 sm:grid-cols-2">
                        <Field
                          label={t("pages.config.persona_model")}
                          hint={t("pages.config.persona_model_hint")}
                        >
                          <input
                            type="text"
                            value={persona.primaryModel}
                            disabled={disabled}
                            placeholder="miniMax"
                            onChange={(e) =>
                              onChange(
                                persona.key,
                                "primaryModel",
                                e.target.value,
                              )
                            }
                            className={glassInputBase}
                          />
                        </Field>

                        <Field
                          label={t("pages.config.persona_subagent_model")}
                          hint={t("pages.config.persona_subagent_model_hint")}
                        >
                          <input
                            type="text"
                            value={persona.subagentModel}
                            disabled={disabled}
                            onChange={(e) =>
                              onChange(
                                persona.key,
                                "subagentModel",
                                e.target.value,
                              )
                            }
                            className={glassInputBase}
                          />
                        </Field>
                      </div>

                      <div className="grid gap-5 sm:grid-cols-2">
                        <Field
                          label={t("pages.config.persona_fallbacks")}
                          hint={t("pages.config.persona_list_hint")}
                        >
                          <textarea
                            value={persona.fallbackModelsText}
                            disabled={disabled}
                            onChange={(e) =>
                              onChange(
                                persona.key,
                                "fallbackModelsText",
                                e.target.value,
                              )
                            }
                            className={glassTextareaBase}
                          />
                        </Field>

                        <Field
                          label={t("pages.config.persona_subagent_fallbacks")}
                          hint={t("pages.config.persona_list_hint")}
                        >
                          <textarea
                            value={persona.subagentFallbacksText}
                            disabled={disabled}
                            onChange={(e) =>
                              onChange(
                                persona.key,
                                "subagentFallbacksText",
                                e.target.value,
                              )
                            }
                            className={glassTextareaBase}
                          />
                        </Field>
                      </div>
                    </div>

                    <div className="my-8 h-px w-full bg-gradient-to-r from-transparent via-white/10 to-transparent" />

                    {/* Section: System Prompt */}
                    <div className="mb-8">
                      <PersonaPromptField
                        persona={persona}
                        disabled={disabled}
                        onChange={(value) =>
                          onChange(persona.key, "systemPrompt", value)
                        }
                      />
                    </div>

                    {/* Section: Capabilities */}
                    <div className="space-y-5">
                      <h5 className="mb-4 text-xs font-semibold tracking-wider text-white/40 uppercase">
                        Capabilities & Access
                      </h5>

                      <div className="grid gap-5 sm:grid-cols-3">
                        <Field
                          label={t("pages.config.persona_skills")}
                          hint={t("pages.config.persona_list_hint")}
                        >
                          <textarea
                            value={persona.skillsText}
                            disabled={disabled}
                            onChange={(e) =>
                              onChange(
                                persona.key,
                                "skillsText",
                                e.target.value,
                              )
                            }
                            className={glassTextareaBase}
                          />
                        </Field>

                        <Field
                          label={t("pages.config.persona_mcp_servers")}
                          hint={t("pages.config.persona_list_hint")}
                        >
                          <textarea
                            value={persona.mcpServersText}
                            disabled={disabled}
                            onChange={(e) =>
                              onChange(
                                persona.key,
                                "mcpServersText",
                                e.target.value,
                              )
                            }
                            className={glassTextareaBase}
                          />
                        </Field>

                        <Field
                          label={t("pages.config.persona_allowed_agents")}
                          hint={t("pages.config.persona_list_hint")}
                        >
                          <textarea
                            value={persona.allowedAgentsText}
                            disabled={disabled}
                            onChange={(e) =>
                              onChange(
                                persona.key,
                                "allowedAgentsText",
                                e.target.value,
                              )
                            }
                            className={glassTextareaBase}
                          />
                        </Field>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}
