import type { ModelInfo } from "@/api/models"
import { formatModelPrice } from "@/lib/model-pricing"

import { ModelCapabilityBadges } from "../models/model-capability-badges"

function getCredentialLabel(model: ModelInfo) {
  if (model.auth_method === "oauth") return "OAuth"
  return model.api_key || "Unconfigured"
}

export function ChatModelOption({ model }: { model: ModelInfo }) {
  const price = formatModelPrice(model)

  return (
    <div className="flex min-w-0 flex-col gap-1">
      <span className="truncate text-sm font-medium">{model.model_name}</span>
      <div className="text-muted-foreground flex min-w-0 items-center gap-2 text-[11px]">
        {price ? (
          <span className="bg-muted rounded px-1.5 py-0.5 font-medium">
            {price}
          </span>
        ) : null}
        <span className="truncate font-mono">{getCredentialLabel(model)}</span>
      </div>
      <ModelCapabilityBadges model={model.catalog} limit={3} />
    </div>
  )
}
