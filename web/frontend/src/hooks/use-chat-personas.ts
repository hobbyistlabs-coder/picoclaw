import * as React from "react"

import { getAppConfig } from "@/api/channels"

const AUTO_PERSONA_ID = "__auto__"
const MAIN_PERSONA_ID = "main"

const asRecord = (value: unknown): Record<string, unknown> =>
  value && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : {}
const asArray = (value: unknown): unknown[] => (Array.isArray(value) ? value : [])
const asString = (value: unknown): string => (typeof value === "string" ? value : "")

export interface ChatPersonaOption {
  id: string
  label: string
}

export function useChatPersonas() {
  const [personas, setPersonas] = React.useState<ChatPersonaOption[]>([
    { id: AUTO_PERSONA_ID, label: "Auto" },
  ])

  const reload = React.useCallback((shouldApply?: () => boolean) => {
    getAppConfig()
      .then((config) => {
        if (shouldApply && !shouldApply()) return

        const agents = asRecord(asRecord(config).agents)
        const list = asArray(agents.list)
        const configured = new Map<string, ChatPersonaOption>()

        configured.set(AUTO_PERSONA_ID, { id: AUTO_PERSONA_ID, label: "Auto" })
        configured.set(MAIN_PERSONA_ID, { id: MAIN_PERSONA_ID, label: "Main" })

        list.forEach((item, index) => {
          const record = asRecord(item)
          const id = asString(record.id).trim()
          if (!id) return
          configured.set(id, {
            id,
            label: asString(record.name).trim() || id || `Persona ${index + 1}`,
          })
        })

        setPersonas(Array.from(configured.values()))
      })
      .catch(() => {
        if (shouldApply && !shouldApply()) return
        setPersonas([
          { id: AUTO_PERSONA_ID, label: "Auto" },
          { id: MAIN_PERSONA_ID, label: "Main" },
        ])
      })
  }, [])

  React.useEffect(() => {
    let active = true
    const handleConfigUpdate = () => reload(() => active)

    reload(() => active)
    window.addEventListener("jane-config-updated", handleConfigUpdate)

    return () => {
      active = false
      window.removeEventListener("jane-config-updated", handleConfigUpdate)
    }
  }, [reload])

  return personas
}
