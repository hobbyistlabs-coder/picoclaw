export interface ChatMetrics {
  tool_calls: number
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
  estimated_cost_usd?: number
  has_estimated_cost?: boolean
}

function hasUsage(metrics?: ChatMetrics | null): boolean {
  return Boolean(
    metrics &&
    (metrics.prompt_tokens > 0 ||
      metrics.completion_tokens > 0 ||
      metrics.total_tokens > 0),
  )
}

export function addChatMetrics(
  base?: ChatMetrics | null,
  next?: ChatMetrics | null,
): ChatMetrics | null {
  if (!base && !next) return null
  return {
    tool_calls: (base?.tool_calls ?? 0) + (next?.tool_calls ?? 0),
    prompt_tokens: (base?.prompt_tokens ?? 0) + (next?.prompt_tokens ?? 0),
    completion_tokens:
      (base?.completion_tokens ?? 0) + (next?.completion_tokens ?? 0),
    total_tokens: (base?.total_tokens ?? 0) + (next?.total_tokens ?? 0),
    estimated_cost_usd:
      (base?.estimated_cost_usd ?? 0) + (next?.estimated_cost_usd ?? 0),
    has_estimated_cost:
      (hasUsage(base) ? Boolean(base?.has_estimated_cost) : true) &&
      (hasUsage(next) ? Boolean(next?.has_estimated_cost) : true) &&
      (hasUsage(base) || hasUsage(next)),
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

export function formatEstimatedCost(metrics?: ChatMetrics | null): string {
  if (!metrics?.has_estimated_cost) return "n/a"
  const amount = metrics.estimated_cost_usd ?? 0
  return amount < 0.01 ? `$${amount.toFixed(4)}` : `$${amount.toFixed(2)}`
}
