import {
  IconAlertTriangle,
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
import { CodeBlock } from "@/components/chat/code-block"
import { Button } from "@/components/ui/button"
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible"
import type { ChatToolEvent } from "@/hooks/use-pico-chat"
import { formatMessageTime } from "@/hooks/use-pico-chat"
import type { ChatMetricPricing, ChatMetrics } from "@/lib/chat-metrics"

interface AssistantMessageProps {
  content: string
  timestamp?: string | number
  metrics?: ChatMetrics
  pricing?: ChatMetricPricing | null
  reasoningContent?: string
  toolEvents?: ChatToolEvent[]
  pending?: boolean
}

// Returns true when a tool event represents a failure
const isToolError = (event: ChatToolEvent): boolean => {
  if (event.error) return true
  if (event.status === "error" || event.status === "failed") return true
  if (
    typeof event.result === "string" &&
    (event.result.includes("Execution failed") ||
      event.result.includes("HTTP 4") ||
      event.result.includes("HTTP 5"))
  )
    return true
  return false
}

// Utility function to format tool event results safely
const formatStructured = (data: unknown): string => {
  if (!data) return ""

  if (typeof data === "string") {
    try {
      // Check if it's a stringified JSON object
      return JSON.stringify(JSON.parse(data), null, 2)
    } catch {
      // If it's just a normal string (like a terminal log), return as is
      return data
    }
  }

  try {
    return JSON.stringify(data, null, 2)
  } catch {
    return String(data)
  }
}

export function AssistantMessage({
  content,
  timestamp = "",
  metrics,
  pricing,
  reasoningContent,
  toolEvents = [],
  pending = false,
}: AssistantMessageProps) {
  const [isCopied, setIsCopied] = useState(false)
  const formattedTimestamp =
    timestamp !== "" ? formatMessageTime(timestamp) : ""

  const hasReasoning = Boolean(reasoningContent?.trim())
  const hasToolEvents = toolEvents.length > 0

  const handleCopy = async () => {
    if (!content) return
    try {
      await navigator.clipboard.writeText(content)
      setIsCopied(true)
      setTimeout(() => setIsCopied(false), 2000)
    } catch (err) {
      console.error("Failed to copy text: ", err)
    }
  }

  return (
    <div className="group flex w-full flex-col gap-3 pb-6">
      {/* 1. Header & Metrics */}
      <div className="flex flex-wrap items-center justify-between gap-3 px-2">
        <div className="flex items-center gap-2.5">
          <div className="flex size-6 items-center justify-center rounded-md border border-indigo-500/30 bg-indigo-500/10 shadow-sm">
            <div className="size-1.5 animate-pulse rounded-full bg-indigo-400" />
          </div>
          <span className="text-xs font-bold tracking-widest text-white/60 uppercase">
            Jane AI
          </span>
          {formattedTimestamp && (
            <span className="text-[10px] font-medium text-white/30">
              • {formattedTimestamp}
            </span>
          )}
        </div>
        <div className="opacity-0 transition-opacity duration-300 group-hover:opacity-100">
          <ChatMetricsPills metrics={metrics} pricing={pricing} />
        </div>
      </div>

      {/* 2. Process Flow (Reasoning + Tools) */}
      {(hasReasoning || hasToolEvents) && (
        <div className="flex flex-col gap-2 pl-2">
          {/* Reasoning / Cognitive Trace */}
          {hasReasoning && (
            <Collapsible defaultOpen={pending} className="w-full max-w-3xl">
              <div className="overflow-hidden rounded-xl border border-white/[0.04] bg-white/[0.01] transition-colors hover:bg-white/[0.02]">
                <CollapsibleTrigger className="group/trigger flex w-full items-center justify-between px-4 py-3">
                  <span className="flex items-center gap-2 text-[11px] font-semibold tracking-wider text-indigo-400/80 uppercase">
                    <IconBrain className="size-4" />
                    Cognitive Trace
                  </span>
                  <IconChevronDown className="size-4 text-white/30 transition-transform duration-300 group-data-[state=open]/trigger:rotate-180" />
                </CollapsibleTrigger>
                <CollapsibleContent className="border-t border-white/[0.04] bg-black/30 px-4 py-3">
                  <div className="prose-invert prose-sm font-serif leading-relaxed text-white/60 italic">
                    {reasoningContent}
                  </div>
                </CollapsibleContent>
              </div>
            </Collapsible>
          )}

          {/* Tool Events / System Execution Timeline */}
          {hasToolEvents && (() => {
            const hasAnyError = toolEvents.some(isToolError)
            const failCount = toolEvents.filter(isToolError).length
            return (
              <Collapsible defaultOpen={pending} className="w-full max-w-3xl">
                <div className="overflow-hidden rounded-xl border border-white/[0.04] bg-white/[0.01] transition-colors hover:bg-white/[0.02]">
                  <CollapsibleTrigger className="group/trigger flex w-full items-center justify-between px-4 py-3">
                    <span
                      className={`flex items-center gap-2 text-[11px] font-semibold tracking-wider uppercase ${hasAnyError ? "text-red-400/80" : "text-emerald-400/80"}`}
                    >
                      {hasAnyError ? (
                        <IconAlertTriangle className="size-4" />
                      ) : (
                        <IconTool className="size-4" />
                      )}
                      System Execution ({toolEvents.length}
                      {hasAnyError ? ` · ${failCount} failed` : ""})
                    </span>
                    <IconChevronDown className="size-4 text-white/30 transition-transform duration-300 group-data-[state=open]/trigger:rotate-180" />
                  </CollapsibleTrigger>

                  <CollapsibleContent className="border-t border-white/[0.04] bg-black/30 p-4">
                    <div className="relative ml-2 space-y-4 border-l border-white/[0.08] pb-1">
                      {toolEvents.map((event) => {
                        const hasError = isToolError(event)
                        return (
                          <div key={event.id} className="relative pl-5">
                            {/* Timeline Node — red for errors, green for success, amber for in-progress */}
                            <div
                              className={`absolute top-1.5 -left-[5px] size-2.5 rounded-full border-2 border-black ring-1 ring-white/10 ${
                                hasError
                                  ? "bg-red-500/70"
                                  : event.status === "completed"
                                    ? "bg-emerald-500/50"
                                    : "animate-pulse bg-amber-500/50"
                              }`}
                            />

                            <div className="flex flex-col gap-1.5">
                              <div className="flex items-center justify-between">
                                <span
                                  className={`font-mono text-xs font-semibold ${hasError ? "text-red-300/90" : "text-white/80"}`}
                                >
                                  {event.toolName || event.name || "Function"}
                                </span>
                                <span
                                  className={`rounded px-1.5 py-0.5 text-[9px] font-bold tracking-wider uppercase ${
                                    hasError
                                      ? "bg-red-500/15 text-red-400"
                                      : event.status === "completed"
                                        ? "bg-emerald-500/10 text-emerald-400"
                                        : "animate-pulse bg-amber-500/10 text-amber-400"
                                  }`}
                                >
                                  {hasError ? "failed" : event.status}
                                </span>
                              </div>

                              {event.summary && (
                                <p className="text-xs leading-relaxed text-white/50">
                                  {event.summary}
                                </p>
                              )}

                              {/* Error callout — shown prominently when the tool failed */}
                              {hasError && (
                                <div className="mt-1 flex items-start gap-2 rounded-lg border border-red-500/20 bg-red-500/[0.07] px-3 py-2">
                                  <IconAlertTriangle className="mt-0.5 size-3.5 shrink-0 text-red-400" />
                                  <div className="min-w-0 flex-1">
                                    <p className="text-[11px] font-semibold tracking-wide text-red-400 uppercase">
                                      Execution Failed
                                    </p>
                                    <p className="mt-0.5 break-words font-mono text-[11px] leading-relaxed text-red-300/70">
                                      {event.error ||
                                        (typeof event.result === "string"
                                          ? event.result
                                          : JSON.stringify(event.result))}
                                    </p>
                                  </div>
                                </div>
                              )}

                              {/* Raw result — only shown for successful calls */}
                              {!hasError && event.result && (
                                <div className="mt-1 overflow-hidden rounded-lg border border-white/[0.05] bg-black/50">
                                  <CodeBlock
                                    code={formatStructured(event.result)}
                                    language="json"
                                  />
                                </div>
                              )}
                            </div>
                          </div>
                        )
                      })}
                    </div>
                  </CollapsibleContent>
                </div>
              </Collapsible>
            )
          })()}
        </div>
      )}

      {/* 3. Final Output Message */}
      <div className="relative mt-1 overflow-hidden rounded-2xl border border-white/[0.08] bg-[#111114]/80 shadow-2xl backdrop-blur-xl">
        <div className="prose prose-invert max-w-none p-5 pr-12 text-[15px] leading-relaxed selection:bg-indigo-500/30">
          {content ? (
            <ReactMarkdown
              remarkPlugins={[remarkGfm]}
              components={{
                code(props) {
                  const { className, children } = props
                  const match = /language-(\w+)/.exec(className || "")
                  if (!match) {
                    return (
                      <code className="rounded-md bg-white/[0.05] px-1.5 py-0.5 font-mono text-[0.85em] text-indigo-200">
                        {children}
                      </code>
                    )
                  }
                  return (
                    <div className="my-4 overflow-hidden rounded-xl border border-white/[0.05]">
                      <CodeBlock
                        code={String(children).replace(/\n$/, "")}
                        language={match[1]}
                      />
                    </div>
                  )
                },
              }}
            >
              {content}
            </ReactMarkdown>
          ) : (
            <div className="flex items-center gap-3 py-1 text-white/40">
              <div className="flex gap-1.5">
                <div className="size-1.5 animate-bounce rounded-full bg-current [animation-delay:-0.3s]" />
                <div className="size-1.5 animate-bounce rounded-full bg-current [animation-delay:-0.15s]" />
                <div className="size-1.5 animate-bounce rounded-full bg-current" />
              </div>
              <span className="text-sm italic">Processing output...</span>
            </div>
          )}
        </div>

        {/* Copy Button */}
        {content && (
          <Button
            variant="ghost"
            size="icon"
            className="absolute top-4 right-4 size-8 rounded-lg border border-white/10 bg-white/5 text-white/50 opacity-0 transition-all duration-200 group-hover:opacity-100 hover:bg-white/10 hover:text-white"
            onClick={handleCopy}
            aria-label="Copy message content"
          >
            {isCopied ? (
              <IconCheck className="size-4 text-emerald-400" />
            ) : (
              <IconCopy className="size-4" />
            )}
          </Button>
        )}
      </div>
    </div>
  )
}
