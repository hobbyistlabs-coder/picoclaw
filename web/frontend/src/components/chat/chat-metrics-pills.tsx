import { useTranslation } from "react-i18next"

import type { ChatMetrics } from "@/lib/chat-metrics"
import {
  formatEstimatedCost,
  formatTokenCount,
  hasChatMetrics,
} from "@/lib/chat-metrics"

interface ChatMetricsPillsProps {
  metrics?: ChatMetrics | null
}

export function ChatMetricsPills({ metrics }: ChatMetricsPillsProps) {
  const { t } = useTranslation()

  if (!hasChatMetrics(metrics)) return null

  return (
    <div className="flex flex-wrap items-center gap-2 text-xs">
      <span className="bg-muted text-muted-foreground rounded-full px-2.5 py-1">
        {t("chat.metrics.tools")}: {metrics?.tool_calls ?? 0}
      </span>
      <span className="bg-muted text-muted-foreground rounded-full px-2.5 py-1">
        {t("chat.metrics.tokens")}:{" "}
        {formatTokenCount(metrics?.total_tokens ?? 0)}
      </span>
      <span className="bg-muted text-muted-foreground rounded-full px-2.5 py-1">
        {t("chat.metrics.cost")}: {formatEstimatedCost(metrics)}
      </span>
    </div>
  )
}
