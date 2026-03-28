import { IconMasksTheater } from "@tabler/icons-react"
import * as React from "react"

import { getAppConfig } from "@/api/channels"

const asRecord = (value: unknown): Record<string, unknown> =>
  value && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : {}
const asArray = (value: unknown): unknown[] => (Array.isArray(value) ? value : [])
const asString = (value: unknown): string => (typeof value === "string" ? value : "")

export interface SidebarPersonaNavItem {
  key: string
  title: string
  to: "/config"
  hash: string
  icon: typeof IconMasksTheater
}

export function useSidebarPersonas() {
  const [personas, setPersonas] = React.useState<SidebarPersonaNavItem[]>([])

  const reload = React.useCallback((shouldApply?: () => boolean) => {
    getAppConfig()
      .then((config) => {
        if (shouldApply && !shouldApply()) return
        const agents = asRecord(asRecord(config).agents)
        const list = asArray(agents.list)
        setPersonas(
          list
            .map((item, index) => {
              const record = asRecord(item)
              const id = asString(record.id)
              if (!id) return null
              return {
                key: id,
                title: asString(record.name) || id || `Persona ${index + 1}`,
                to: "/config" as const,
                hash: `persona-${id}`,
                icon: IconMasksTheater,
              }
            })
            .filter((item): item is SidebarPersonaNavItem => item !== null),
        )
      })
      .catch(() => {
        if (shouldApply && !shouldApply()) return
        setPersonas([])
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
