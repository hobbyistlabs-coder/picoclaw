import type { OpenRouterModel } from "@/api/openrouter"
import { getCapabilityLabels } from "@/lib/openrouter"

const LABELS: Record<string, string> = {
  image: "Vision",
  file: "Files",
  audio: "Audio",
  video: "Video",
  tools: "Tools",
  json: "JSON",
  reasoning: "Reasoning",
}

export function ModelCapabilityBadges({
  model,
  limit = 4,
}: {
  model?: OpenRouterModel
  limit?: number
}) {
  const labels = getCapabilityLabels(model).slice(0, limit)
  if (!labels.length) return null

  return (
    <div className="flex flex-wrap gap-1.5">
      {labels.map((label) => (
        <span
          key={label}
          className="bg-muted text-muted-foreground rounded-full px-2 py-0.5 text-[10px] font-medium"
        >
          {LABELS[label] ?? label}
        </span>
      ))}
    </div>
  )
}
