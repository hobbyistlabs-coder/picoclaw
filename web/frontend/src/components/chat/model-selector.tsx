import { useTranslation } from "react-i18next"

import type { ModelInfo } from "@/api/models"
import { ChatModelOption } from "@/components/chat/chat-model-option"
import { ChatModelTrigger } from "@/components/chat/chat-model-trigger"
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectSeparator,
  SelectTrigger,
} from "@/components/ui/select"

interface ModelSelectorProps {
  defaultModelName: string
  apiKeyModels: ModelInfo[]
  oauthModels: ModelInfo[]
  localModels: ModelInfo[]
  onValueChange: (modelName: string) => void
}

export function ModelSelector({
  defaultModelName,
  apiKeyModels,
  oauthModels,
  localModels,
  onValueChange,
}: ModelSelectorProps) {
  const { t } = useTranslation()
  const allModels = [...apiKeyModels, ...oauthModels, ...localModels]
  const selectedModel = allModels.find(
    (model) => model.model_name === defaultModelName,
  )

  return (
    <Select value={defaultModelName} onValueChange={onValueChange}>
      <SelectTrigger
        size="sm"
        className="bg-slate/6 hover:bg-slate/10 min-h-11 w-[240px] max-w-[240px] min-w-0 overflow-hidden rounded-2xl border-white/10 px-3 py-1 text-white/88 shadow-none backdrop-blur-sm transition-colors hover:border-white/20 focus-visible:ring-0 sm:w-[320px] sm:max-w-[320px]"
      >
        <ChatModelTrigger
          model={selectedModel}
          placeholder={t("chat.noModel")}
        />
      </SelectTrigger>
      <SelectContent
        position="popper"
        className="w-[min(32rem,calc(100vw-2rem))] border-white/10 bg-slate-950/96 text-white"
      >
        {apiKeyModels.length > 0 && (
          <SelectGroup>
            <SelectLabel>{t("chat.modelGroup.apikey")}</SelectLabel>
            {apiKeyModels.map((model) => (
              <SelectItem
                key={model.index}
                value={model.model_name}
                textValue={model.model_name}
              >
                <ChatModelOption model={model} />
              </SelectItem>
            ))}
          </SelectGroup>
        )}
        {apiKeyModels.length > 0 &&
          (oauthModels.length > 0 || localModels.length > 0) && (
            <SelectSeparator />
          )}

        {oauthModels.length > 0 && (
          <SelectGroup>
            <SelectLabel>{t("chat.modelGroup.oauth")}</SelectLabel>
            {oauthModels.map((model) => (
              <SelectItem
                key={model.index}
                value={model.model_name}
                textValue={model.model_name}
              >
                <ChatModelOption model={model} />
              </SelectItem>
            ))}
          </SelectGroup>
        )}
        {oauthModels.length > 0 &&
          (localModels.length > 0 || apiKeyModels.length > 0) && (
            <SelectSeparator />
          )}

        {localModels.length > 0 && (
          <SelectGroup>
            <SelectLabel>{t("chat.modelGroup.local")}</SelectLabel>
            {localModels.map((model) => (
              <SelectItem
                key={model.index}
                value={model.model_name}
                textValue={model.model_name}
              >
                <ChatModelOption model={model} />
              </SelectItem>
            ))}
          </SelectGroup>
        )}
      </SelectContent>
    </Select>
  )
}
