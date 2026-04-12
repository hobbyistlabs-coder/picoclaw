export interface ChatMetrics {
  tool_calls: number
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
  estimated_cost_usd?: number
  has_estimated_cost?: boolean
}

export interface ChatMetricPricing {
  price_per_m_token?: number
  input_price_per_m_token?: number
  output_price_per_m_token?: number
}

export interface ChatCostDebugInfo {
  reason:
    | "missing_metrics"
    | "backend_estimate"
    | "missing_pricing"
    | "no_priced_rates"
    | "computed"
  amount: number | null
  inputRate: number
  outputRate: number
  hasTokens: boolean
}

export function addChatMetrics(
  base?: ChatMetrics | null,
  next?: ChatMetrics | null,
): ChatMetrics | null {
  if (!base && !next) return null
  const hasEstimatedCost =
    Boolean(base?.has_estimated_cost) || Boolean(next?.has_estimated_cost)
  return {
    tool_calls: (base?.tool_calls ?? 0) + (next?.tool_calls ?? 0),
    prompt_tokens: (base?.prompt_tokens ?? 0) + (next?.prompt_tokens ?? 0),
    completion_tokens:
      (base?.completion_tokens ?? 0) + (next?.completion_tokens ?? 0),
    total_tokens: (base?.total_tokens ?? 0) + (next?.total_tokens ?? 0),
    estimated_cost_usd:
      (base?.estimated_cost_usd ?? 0) + (next?.estimated_cost_usd ?? 0),
    has_estimated_cost: hasEstimatedCost,
  }
}

export function sumChatMetrics(
  metrics: Array<ChatMetrics | null | undefined>,
): ChatMetrics | null {
  return metrics.reduce<ChatMetrics | null>((total, current) => {
    return addChatMetrics(total, current)
  }, null)
}

export function hasChatMetrics(metrics?: ChatMetrics | null): boolean {
  return Boolean(
    metrics &&
    (metrics.tool_calls > 0 ||
      metrics.prompt_tokens > 0 ||
      metrics.completion_tokens > 0 ||
      metrics.total_tokens > 0 ||
      metrics.has_estimated_cost),
  )
}

export function formatTokenCount(value: number): string {
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(2)}M`
  if (value >= 1_000) return `${(value / 1_000).toFixed(1)}K`
  return value.toString()
}

export function estimateChatCost(
  metrics?: ChatMetrics | null,
  pricing?: ChatMetricPricing | null,
): number | null {
  return getChatCostDebugInfo(metrics, pricing).amount
}

export function getChatCostDebugInfo(
  metrics?: ChatMetrics | null,
  pricing?: ChatMetricPricing | null,
): ChatCostDebugInfo {
  if (!metrics) {
    return {
      reason: "missing_metrics",
      amount: null,
      inputRate: 0,
      outputRate: 0,
      hasTokens: false,
    }
  }
  if (metrics.has_estimated_cost) {
    return {
      reason: "backend_estimate",
      amount: metrics.estimated_cost_usd ?? 0,
      inputRate: 0,
      outputRate: 0,
      hasTokens: metrics.total_tokens > 0,
    }
  }
  if (!pricing) {
    return {
      reason: "missing_pricing",
      amount: null,
      inputRate: 0,
      outputRate: 0,
      hasTokens: metrics.total_tokens > 0,
    }
  }

  const inputRate =
    pricing.input_price_per_m_token ?? pricing.price_per_m_token ?? 0
  const outputRate =
    pricing.output_price_per_m_token ?? pricing.price_per_m_token ?? 0

  if (inputRate <= 0 && outputRate <= 0) {
    return {
      reason: "no_priced_rates",
      amount: null,
      inputRate,
      outputRate,
      hasTokens: metrics.total_tokens > 0,
    }
  }

  let amount =
    (metrics.prompt_tokens * inputRate) / 1_000_000 +
    (metrics.completion_tokens * outputRate) / 1_000_000

  if (amount === 0 && metrics.total_tokens > 0) {
    const fallbackRate = inputRate > 0 ? inputRate : outputRate
    amount = (metrics.total_tokens * fallbackRate) / 1_000_000
  }

  return {
    reason: "computed",
    amount,
    inputRate,
    outputRate,
    hasTokens: metrics.total_tokens > 0,
  }
}

export function formatEstimatedCost(
  metrics?: ChatMetrics | null,
  pricing?: ChatMetricPricing | null,
): string {
  const amount = estimateChatCost(metrics, pricing)
  if (amount === null) return "n/a"
  return amount < 0.01 ? `$${amount.toFixed(4)}` : `$${amount.toFixed(2)}`
}
