import { IconFile } from "@tabler/icons-react"
import ReactMarkdown from "react-markdown"
import remarkGfm from "remark-gfm"

import { CodeBlock } from "@/components/chat/code-block"

export interface Attachment {
  id: string
  url: string
  name: string
  type: string
}

interface UserMessageProps {
  content: string
  displayName: string
  attachments?: Attachment[]
}

export function UserMessage({
  content,
  displayName,
  attachments = [],
}: UserMessageProps) {
  return (
    <div className="flex w-full flex-col items-end gap-2">
      {/* Display Name Badge */}
      <div className="flex items-center gap-2 px-1">
        <span className="rounded-full border border-indigo-500/20 bg-indigo-500/10 px-2.5 py-1 text-[11px] font-bold tracking-[0.15em] text-indigo-400 uppercase shadow-inner backdrop-blur-md">
          {displayName}
        </span>
      </div>

      {/* Message Bubble Container */}
      <div className="max-w-[85%] rounded-[2rem] rounded-tr-sm border border-indigo-500/20 bg-[linear-gradient(135deg,rgba(99,102,241,0.12),rgba(99,102,241,0.05))] px-5 py-4 text-[15px] leading-relaxed text-white/90 shadow-[0_8px_32px_rgba(0,0,0,0.12)] backdrop-blur-xl sm:max-w-[74%] md:px-6">
        {/* Render Attachments (if any) */}
        {attachments.length > 0 && (
          <div className="mb-4 flex flex-wrap gap-2">
            {attachments.map((file) => {
              const isImage = file.type.startsWith("image/")
              return isImage ? (
                <div
                  key={file.id}
                  className="relative h-32 w-32 overflow-hidden rounded-2xl border border-white/[0.05] bg-white/[0.02] shadow-inner"
                >
                  <img
                    src={file.url}
                    alt={file.name}
                    className="h-full w-full object-cover transition-transform duration-300 hover:scale-110"
                  />
                </div>
              ) : (
                <div
                  key={file.id}
                  className="flex items-center gap-2 rounded-full border border-white/[0.05] bg-white/[0.03] px-3 py-1.5 shadow-inner backdrop-blur-xl"
                >
                  <div className="flex h-6 w-6 items-center justify-center rounded-full bg-white/5">
                    <IconFile className="h-3 w-3 text-white/60" />
                  </div>
                  <span className="max-w-[150px] truncate text-xs text-white/80">
                    {file.name}
                  </span>
                </div>
              )
            })}
          </div>
        )}

        {/* Markdown Prose */}
        <div className="prose prose-invert prose-p:my-2 prose-pre:my-2 max-w-none text-inherit">
          <ReactMarkdown
            remarkPlugins={[remarkGfm]}
            components={{
              code(props) {
                const { className, children } = props
                const match = /language-(\w+)/.exec(className || "")
                const code = String(children).replace(/\n$/, "")

                // Inline code styling matching the dark glass theme
                if (!match) {
                  return (
                    <code className="rounded-md border border-white/[0.05] bg-white/[0.05] px-1.5 py-0.5 font-mono text-[0.9em] text-indigo-300">
                      {children}
                    </code>
                  )
                }

                return <CodeBlock code={code} language={match[1]} />
              },
              // Optional: Soften link colors to match the palette
              a(props) {
                return (
                  <a
                    {...props}
                    className="text-indigo-400 underline decoration-indigo-400/30 underline-offset-4 transition-colors hover:decoration-indigo-400"
                  >
                    {props.children}
                  </a>
                )
              },
            }}
          >
            {content}
          </ReactMarkdown>
        </div>
      </div>
    </div>
  )
}
