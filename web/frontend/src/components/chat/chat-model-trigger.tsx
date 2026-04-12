import type { ModelInfo } from "@/api/models"
import { formatModelPrice } from "@/lib/model-pricing"

import { ModelCapabilityBadges } from "../models/model-capability-badges"

export function ChatModelTrigger({
  model,
  placeholder,
}: {
  model?: ModelInfo
  placeholder: string
}) {
  const price = model ? formatModelPrice(model) : null

  return (
    <div className="flex min-w-0 flex-1 flex-col items-start gap-1 overflow-hidden py-0.5">
      <div className="flex w-full min-w-0 items-center gap-2 overflow-hidden">
        <span className="block min-w-0 flex-1 truncate text-sm font-semibold text-white/92">
          {model?.model_name || placeholder}
        </span>
        {price ? (
          <span className="bg-slate/8 shrink-0 rounded-full border border-white/10 px-2 py-0.5 text-[10px] font-medium text-white/72">
            {price}
          </span>
        ) : null}
      </div>
      <div className="flex w-full items-center justify-between gap-2">
        <span className="truncate text-[10px] tracking-[0.24em] text-white/40 uppercase">
          {model?.auth_method === "oauth" ? "OAuth" : "API Key"}
        </span>
        <ModelCapabilityBadges model={model?.catalog} limit={2} />
      </div>
    </div>
  )
}
