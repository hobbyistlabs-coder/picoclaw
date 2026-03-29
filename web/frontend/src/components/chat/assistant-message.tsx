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
import { CodeBlock } from "@/components/chat/code-block"
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

  const formatStructured = (value: unknown) => {
    if (!value) return ""
    if (typeof value === "string") {
      try {
        return JSON.stringify(JSON.parse(value), null, 2)
      } catch {
        return value
      }
    }
    return JSON.stringify(value, null, 2)
  }

  const getResultLanguage = (value: string) => {
    try {
      JSON.parse(value)
      return "json"
    } catch {
      return "text"
    }
  }

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
          <div className="border-primary/18 bg-primary/8 rounded-xl border">
            <CollapsibleTrigger className="flex w-full items-center justify-between gap-3 px-4 py-3 text-left">
              <span className="text-primary flex items-center gap-2 text-sm font-medium">
                <IconBrain className="size-4" />
                Thinking
              </span>
              <IconChevronDown className="size-4 opacity-60" />
            </CollapsibleTrigger>
            <CollapsibleContent className="text-foreground/88 border-t border-white/8 px-4 py-3 text-sm leading-6 whitespace-pre-wrap">
              <CodeBlock code={reasoningContent || ""} language="markdown" />
            </CollapsibleContent>
          </div>
        </Collapsible>
      )}

      {hasToolEvents && (
        <Collapsible defaultOpen={pending}>
          <div className="border-border/80 bg-card/88 rounded-xl border">
            <CollapsibleTrigger className="flex w-full items-center justify-between gap-3 px-4 py-3 text-left">
              <span className="text-foreground flex items-center gap-2 text-sm font-medium">
                <IconTool className="size-4" />
                Tool Activity
              </span>
              <IconChevronDown className="size-4 opacity-60" />
            </CollapsibleTrigger>
            <CollapsibleContent className="border-border/70 border-t px-4 py-3">
              <div className="flex flex-col gap-3">
                {toolEvents.map((event) => (
                  <div
                    key={event.id}
                    className="border-border/70 bg-background/70 rounded-xl border p-4"
                  >
                    <div className="flex flex-wrap items-center gap-2 text-sm">
                      <span className="text-foreground font-medium">
                        {event.codename || event.label || event.name}
                      </span>
                      <span className="bg-muted text-muted-foreground rounded-full px-2 py-0.5 text-xs uppercase">
                        {event.kind}
                      </span>
                      <span className="bg-primary/14 text-primary rounded-full px-2 py-0.5 text-xs font-medium uppercase">
                        {event.status}
                      </span>
                      {event.toolName && (
                        <span className="bg-secondary text-secondary-foreground rounded-full px-2 py-0.5 text-xs uppercase">
                          {event.toolName}
                        </span>
                      )}
                    </div>
                    {event.codename && event.label && (
                      <p className="text-muted-foreground mt-2 text-sm">
                        {event.label}
                      </p>
                    )}
                    {event.summary && (
                      <p className="text-foreground/78 mt-2 text-sm">
                        {event.summary}
                      </p>
                    )}
                    {typeof event.progressPercent === "number" &&
                      event.progressPercent > 0 && (
                        <p className="text-muted-foreground mt-2 text-xs">
                          Progress {event.progressPercent}%
                        </p>
                      )}
                    {event.error && (
                      <div className="mt-3">
                        <p className="text-destructive mb-2 text-[11px] font-semibold tracking-[0.2em] uppercase">
                          Execution failed
                        </p>
                        <CodeBlock code={event.error} language="text" />
                      </div>
                    )}
                    {event.arguments && (
                      <div className="mt-3">
                        <p className="text-muted-foreground mb-2 text-[11px] font-semibold tracking-[0.2em] uppercase">
                          Request
                        </p>
                        <CodeBlock
                          code={formatStructured(event.arguments)}
                          language="json"
                        />
                      </div>
                    )}
                    {event.result && (
                      <div className="mt-3">
                        <p className="text-muted-foreground mb-2 text-[11px] font-semibold tracking-[0.2em] uppercase">
                          Result
                        </p>
                        <CodeBlock
                          code={formatStructured(event.result)}
                          language={getResultLanguage(event.result)}
                        />
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </CollapsibleContent>
          </div>
        </Collapsible>
      )}

      <div className="bg-card text-card-foreground relative overflow-hidden rounded-xl border">
        <div className="prose dark:prose-invert prose-p:my-2 max-w-none p-4 text-[15px] leading-relaxed">
          {content ? (
            <ReactMarkdown
              remarkPlugins={[remarkGfm]}
              components={{
                code(props) {
                  const { className, children } = props
                  const match = /language-(\w+)/.exec(className || "")
                  const code = String(children).replace(/\n$/, "")

                  if (!match) {
                    return (
                      <code className="rounded bg-black/12 px-1.5 py-0.5 text-[0.9em]">
                        {children}
                      </code>
                    )
                  }

                  return <CodeBlock code={code} language={match[1]} />
                },
              }}
            >
              {content}
            </ReactMarkdown>
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
