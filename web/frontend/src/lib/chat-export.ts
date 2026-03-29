interface ExportToolEvent {
  name: string
  status: string
  kind?: string
  toolName?: string
  summary?: string
  error?: string
  arguments?: Record<string, unknown>
  result?: string
}

interface ExportMessage {
  role: "user" | "assistant"
  content: string
  timestamp: number | string
  reasoningContent?: string
  toolEvents?: ExportToolEvent[]
}

function formatStructured(value: unknown) {
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

export function buildConversationMarkdown(
  messages: ExportMessage[],
  userName: string,
) {
  const lines = ["# Conversation Export", ""]

  messages.forEach((message, index) => {
    const speaker = message.role === "user" ? userName : "Jane AI"
    lines.push(`## ${index + 1}. ${speaker}`)
    lines.push("")
    lines.push(`_Timestamp: ${String(message.timestamp)}_`)
    lines.push("")

    if (message.content.trim()) {
      lines.push(message.content)
      lines.push("")
    }

    if (message.reasoningContent?.trim()) {
      lines.push("### Thinking")
      lines.push("")
      lines.push("```text")
      lines.push(message.reasoningContent)
      lines.push("```")
      lines.push("")
    }

    message.toolEvents?.forEach((event, toolIndex) => {
      lines.push(`### Tool Activity ${toolIndex + 1}`)
      lines.push("")
      lines.push(`- Name: ${event.toolName || event.name}`)
      lines.push(`- Kind: ${event.kind || "tool"}`)
      lines.push(`- Status: ${event.status}`)
      if (event.summary) lines.push(`- Summary: ${event.summary}`)
      if (event.arguments) {
        lines.push("")
        lines.push("```json")
        lines.push(formatStructured(event.arguments))
        lines.push("```")
      }
      if (event.result) {
        lines.push("")
        lines.push("```text")
        lines.push(formatStructured(event.result))
        lines.push("```")
      }
      if (event.error) {
        lines.push("")
        lines.push("```text")
        lines.push(event.error)
        lines.push("```")
      }
      lines.push("")
    })
  })

  return lines.join("\n")
}

export function downloadConversationMarkdown(
  markdown: string,
  sessionId: string,
) {
  const blob = new Blob([markdown], { type: "text/markdown;charset=utf-8" })
  const url = URL.createObjectURL(blob)
  const link = document.createElement("a")
  link.href = url
  link.download = `jane-ai-chat-${sessionId || "session"}.md`
  link.click()
  URL.revokeObjectURL(url)
}
