import { useTranslation } from "react-i18next"

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import type { ChatPersonaOption } from "@/hooks/use-chat-personas"

interface PersonaSelectorProps {
  personas: ChatPersonaOption[]
  value: string
  onValueChange: (value: string) => void
}

export function PersonaSelector({
  personas,
  value,
  onValueChange,
}: PersonaSelectorProps) {
  const { t } = useTranslation()
  const selectedPersona = personas.find((persona) => persona.id === value)
  const selectedLabel =
    value === "__auto__"
      ? t("chat.persona.auto")
      : selectedPersona?.label || t("chat.persona.auto")

  return (
    <Select value={value} onValueChange={onValueChange}>
      <SelectTrigger
        size="sm"
        className="bg-slate/6 hover:bg-slate/10 min-h-11 w-[180px] max-w-[180px] min-w-0 overflow-hidden rounded-2xl border-white/10 px-3 py-1 text-left text-white/88 shadow-none backdrop-blur-sm transition-colors hover:border-white/20 focus-visible:ring-0"
      >
        <div className="flex min-w-0 flex-1 flex-col items-start gap-0.5 overflow-hidden">
          <span className="text-[10px] tracking-[0.24em] text-white/42 uppercase">
            {t("chat.persona.selectorLabel")}
          </span>
          <SelectValue placeholder={t("chat.persona.auto")}>
            <span className="block w-full truncate text-sm font-semibold">
              {selectedLabel}
            </span>
          </SelectValue>
        </div>
      </SelectTrigger>
      <SelectContent
        position="popper"
        className="w-[240px] border-white/10 bg-slate-950/96 text-white"
      >
        {personas.map((persona) => (
          <SelectItem key={persona.id} value={persona.id}>
            {persona.id === "__auto__" ? t("chat.persona.auto") : persona.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
