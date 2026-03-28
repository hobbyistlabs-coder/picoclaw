import { IconPlus, IconTrash } from "@tabler/icons-react"
import { useTranslation } from "react-i18next"

import type { PersonaForm } from "@/components/config/form-model"
import { Field, SwitchCardField } from "@/components/shared-form"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"

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

  return (
    <Card id={sectionId} size="sm">
      <CardHeader className="border-border border-b">
        <div className="flex items-start justify-between gap-3">
          <div className="space-y-1">
            <CardTitle>{t("pages.config.sections.personas")}</CardTitle>
            <CardDescription>
              {t("pages.config.personas_description")}
            </CardDescription>
          </div>
          <Button
            type="button"
            variant="outline"
            onClick={onAdd}
            disabled={disabled}
          >
            <IconPlus className="size-4" />
            {t("pages.config.add_persona")}
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-4 pt-4">
        {personas.length === 0 ? (
          <div className="text-muted-foreground rounded-lg border border-dashed px-4 py-6 text-sm">
            {t("pages.config.personas_empty")}
          </div>
        ) : (
          personas.map((persona, index) => (
            <Card
              key={persona.key}
              id={`persona-${persona.id.trim() || persona.key}`}
              className="border-border/70"
            >
              <CardHeader className="border-border/60 border-b pb-4">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <CardTitle className="text-base">
                      {persona.name.trim() ||
                        persona.id.trim() ||
                        `${t("pages.config.persona")} ${index + 1}`}
                    </CardTitle>
                    <CardDescription>
                      {t("pages.config.persona_card_hint")}
                    </CardDescription>
                  </div>
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => onRemove(persona.key)}
                    disabled={disabled}
                  >
                    <IconTrash className="size-4" />
                    {t("common.remove")}
                  </Button>
                </div>
              </CardHeader>
              <CardContent className="space-y-4 pt-4">
                <Field
                  label={t("pages.config.persona_id")}
                  layout="setting-row"
                >
                  <Input
                    value={persona.id}
                    disabled={disabled}
                    placeholder="coding"
                    onChange={(e) =>
                      onChange(persona.key, "id", e.target.value)
                    }
                  />
                </Field>

                <Field
                  label={t("pages.config.persona_name")}
                  layout="setting-row"
                >
                  <Input
                    value={persona.name}
                    disabled={disabled}
                    placeholder="Coding Persona"
                    onChange={(e) =>
                      onChange(persona.key, "name", e.target.value)
                    }
                  />
                </Field>

                <SwitchCardField
                  label={t("pages.config.persona_default")}
                  hint={t("pages.config.persona_default_hint")}
                  layout="setting-row"
                  checked={persona.isDefault}
                  onCheckedChange={(checked) =>
                    onChange(persona.key, "isDefault", checked)
                  }
                  disabled={disabled}
                />

                <Field
                  label={t("pages.config.persona_workspace")}
                  hint={t("pages.config.persona_workspace_hint")}
                  layout="setting-row"
                >
                  <Input
                    value={persona.workspace}
                    disabled={disabled}
                    placeholder="/root/.picoclaw/workspace/code"
                    onChange={(e) =>
                      onChange(persona.key, "workspace", e.target.value)
                    }
                  />
                </Field>

                <Field
                  label={t("pages.config.persona_model")}
                  hint={t("pages.config.persona_model_hint")}
                  layout="setting-row"
                >
                  <Input
                    value={persona.primaryModel}
                    disabled={disabled}
                    placeholder="miniMax"
                    onChange={(e) =>
                      onChange(persona.key, "primaryModel", e.target.value)
                    }
                  />
                </Field>

                <Field
                  label={t("pages.config.persona_fallbacks")}
                  hint={t("pages.config.persona_list_hint")}
                  layout="setting-row"
                >
                  <Textarea
                    value={persona.fallbackModelsText}
                    disabled={disabled}
                    className="min-h-[88px]"
                    onChange={(e) =>
                      onChange(
                        persona.key,
                        "fallbackModelsText",
                        e.target.value,
                      )
                    }
                  />
                </Field>

                <Field
                  label={t("pages.config.persona_prompt")}
                  hint={t("pages.config.persona_prompt_hint")}
                  layout="setting-row"
                >
                  <Textarea
                    value={persona.systemPrompt}
                    disabled={disabled}
                    className="min-h-[120px]"
                    onChange={(e) =>
                      onChange(persona.key, "systemPrompt", e.target.value)
                    }
                  />
                </Field>

                <Field
                  label={t("pages.config.persona_skills")}
                  hint={t("pages.config.persona_list_hint")}
                  layout="setting-row"
                >
                  <Textarea
                    value={persona.skillsText}
                    disabled={disabled}
                    className="min-h-[88px]"
                    onChange={(e) =>
                      onChange(persona.key, "skillsText", e.target.value)
                    }
                  />
                </Field>

                <Field
                  label={t("pages.config.persona_mcp_servers")}
                  hint={t("pages.config.persona_list_hint")}
                  layout="setting-row"
                >
                  <Textarea
                    value={persona.mcpServersText}
                    disabled={disabled}
                    className="min-h-[88px]"
                    onChange={(e) =>
                      onChange(persona.key, "mcpServersText", e.target.value)
                    }
                  />
                </Field>

                <Field
                  label={t("pages.config.persona_allowed_agents")}
                  hint={t("pages.config.persona_list_hint")}
                  layout="setting-row"
                >
                  <Textarea
                    value={persona.allowedAgentsText}
                    disabled={disabled}
                    className="min-h-[88px]"
                    onChange={(e) =>
                      onChange(persona.key, "allowedAgentsText", e.target.value)
                    }
                  />
                </Field>

                <Field
                  label={t("pages.config.persona_subagent_model")}
                  hint={t("pages.config.persona_subagent_model_hint")}
                  layout="setting-row"
                >
                  <Input
                    value={persona.subagentModel}
                    disabled={disabled}
                    onChange={(e) =>
                      onChange(persona.key, "subagentModel", e.target.value)
                    }
                  />
                </Field>

                <Field
                  label={t("pages.config.persona_subagent_fallbacks")}
                  hint={t("pages.config.persona_list_hint")}
                  layout="setting-row"
                >
                  <Textarea
                    value={persona.subagentFallbacksText}
                    disabled={disabled}
                    className="min-h-[88px]"
                    onChange={(e) =>
                      onChange(
                        persona.key,
                        "subagentFallbacksText",
                        e.target.value,
                      )
                    }
                  />
                </Field>
              </CardContent>
            </Card>
          ))
        )}
      </CardContent>
    </Card>
  )
}
