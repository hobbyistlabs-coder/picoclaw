import type { ModelInfo } from "@/api/models"
import { parseOpenRouterPrice } from "@/lib/openrouter"

function formatRate(rate?: number) {
  return rate ? `$${rate}/M` : null
}

export function formatModelPrice(model: ModelInfo) {
  const input = formatRate(
    model.input_price_per_m_token ??
      parseOpenRouterPrice(model.catalog?.pricing.prompt),
  )
  const output = formatRate(
    model.output_price_per_m_token ??
      parseOpenRouterPrice(model.catalog?.pricing.completion),
  )
  const flat = formatRate(
    model.price_per_m_token ??
      parseOpenRouterPrice(model.catalog?.pricing.prompt),
  )

  if (input && output) {
    if (input === output) return input
    return `In ${input} · Out ${output}`
  }
  return input || output || flat
}
