import {
  IconBrain,
  IconCheck,
  IconChevronDown,
  IconCopy,
  IconTool,
} from "@tabler/icons-react"
import { useState } from "react"
import ReactMarkdown from "react-markdown"
import remarkGfm from "remark-gfm"

import { ChatMetricsPills } from "@/components/chat/chat-metrics-pills"
import { Button } from "@/components/ui/button"
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible"
import type { ChatToolEvent } from "@/hooks/use-pico-chat"
import { formatMessageTime } from "@/hooks/use-pico-chat"
import type { ChatMetrics } from "@/lib/chat-metrics"

interface AssistantMessageProps {
  content: string
  timestamp?: string | number
  metrics?: ChatMetrics
  reasoningContent?: string
  toolEvents?: ChatToolEvent[]
  pending?: boolean
}

export function AssistantMessage({
  content,
  timestamp = "",
  metrics,
  reasoningContent,
  toolEvents = [],
  pending = false,
}: AssistantMessageProps) {
  const [isCopied, setIsCopied] = useState(false)
  const formattedTimestamp =
    timestamp !== "" ? formatMessageTime(timestamp) : ""
  const hasReasoning = Boolean(reasoningContent?.trim())
  const hasToolEvents = toolEvents.length > 0

  const handleCopy = () => {
    navigator.clipboard.writeText(content).then(() => {
      setIsCopied(true)
      setTimeout(() => setIsCopied(false), 2000)
    })
  }

  return (
    <div className="group flex w-full flex-col gap-1.5">
      <div className="text-muted-foreground flex items-center justify-between gap-2 px-1 text-xs opacity-70">
        <div className="flex items-center gap-2">
          <span>Jane AI</span>
          {formattedTimestamp && (
            <>
              <span className="opacity-50">•</span>
              <span>{formattedTimestamp}</span>
            </>
          )}
        </div>
      </div>

      <ChatMetricsPills metrics={metrics} />

      {hasReasoning && (
        <Collapsible defaultOpen={pending}>
          <div className="rounded-xl border border-amber-200/70 bg-amber-50/80">
            <CollapsibleTrigger className="flex w-full items-center justify-between gap-3 px-4 py-3 text-left">
              <span className="flex items-center gap-2 text-sm font-medium">
                <IconBrain className="size-4" />
                Thinking
              </span>
              <IconChevronDown className="size-4 opacity-60" />
            </CollapsibleTrigger>
            <CollapsibleContent className="border-t px-4 py-3 text-sm leading-6 whitespace-pre-wrap">
              {reasoningContent}
            </CollapsibleContent>
          </div>
        </Collapsible>
      )}

      {hasToolEvents && (
        <Collapsible defaultOpen={pending}>
          <div className="rounded-xl border border-slate-200 bg-slate-50/80">
            <CollapsibleTrigger className="flex w-full items-center justify-between gap-3 px-4 py-3 text-left">
              <span className="flex items-center gap-2 text-sm font-medium">
                <IconTool className="size-4" />
                Tool Activity
              </span>
              <IconChevronDown className="size-4 opacity-60" />
            </CollapsibleTrigger>
            <CollapsibleContent className="border-t px-4 py-3">
              <div className="flex flex-col gap-3">
                {toolEvents.map((event) => (
                  <div
                    key={event.id}
                    className="rounded-lg border bg-white p-3"
                  >
                    <div className="flex flex-wrap items-center gap-2 text-sm">
                      <span className="font-medium">
                        {event.label || event.name}
                      </span>
                      <span className="rounded-full bg-slate-100 px-2 py-0.5 text-xs uppercase">
                        {event.kind}
                      </span>
                      <span className="rounded-full bg-slate-900 px-2 py-0.5 text-xs text-white uppercase">
                        {event.status}
                      </span>
                    </div>
                    {event.summary && (
                      <p className="mt-2 text-sm text-slate-600">
                        {event.summary}
                      </p>
                    )}
                    {event.arguments && (
                      <pre className="mt-2 overflow-x-auto rounded-md bg-slate-950 p-3 text-xs text-slate-50">
                        {JSON.stringify(event.arguments, null, 2)}
                      </pre>
                    )}
                    {event.result && (
                      <p className="mt-2 text-sm whitespace-pre-wrap text-slate-600">
                        {event.result}
                      </p>
                    )}
                  </div>
                ))}
              </div>
            </CollapsibleContent>
          </div>
        </Collapsible>
      )}

      <div className="bg-card text-card-foreground relative overflow-hidden rounded-xl border">
        <div className="prose dark:prose-invert prose-p:my-2 prose-pre:my-2 prose-pre:rounded-lg prose-pre:border prose-pre:bg-zinc-950 prose-pre:p-3 max-w-none p-4 text-[15px] leading-relaxed">
          {content ? (
            <ReactMarkdown remarkPlugins={[remarkGfm]}>{content}</ReactMarkdown>
          ) : (
            <p className="text-muted-foreground m-0 text-sm">
              {pending ? "Working..." : "No final response content."}
            </p>
          )}
        </div>
        {content && (
          <Button
            variant="ghost"
            size="icon"
            className="bg-background/50 hover:bg-background/80 absolute top-2 right-2 h-7 w-7 opacity-0 transition-opacity group-hover:opacity-100"
            onClick={handleCopy}
          >
            {isCopied ? (
              <IconCheck className="h-4 w-4 text-green-500" />
            ) : (
              <IconCopy className="text-muted-foreground h-4 w-4" />
            )}
          </Button>
        )}
      </div>
    </div>
  )
}
