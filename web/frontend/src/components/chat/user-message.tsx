import ReactMarkdown from "react-markdown"
import remarkGfm from "remark-gfm"

import { CodeBlock } from "@/components/chat/code-block"

interface UserMessageProps {
  content: string
  displayName: string
}

export function UserMessage({ content, displayName }: UserMessageProps) {
  return (
    <div className="flex w-full flex-col items-end gap-1.5">
      <div className="text-muted-foreground flex items-center gap-2 px-1 text-xs">
        <span className="border-primary/20 bg-primary/8 text-primary rounded-full border px-2.5 py-1 font-medium">
          {displayName}
        </span>
      </div>
      <div className="border-primary/22 bg-primary/10 text-foreground max-w-[74%] rounded-2xl rounded-tr-sm border px-5 py-3 text-[15px] leading-relaxed shadow-sm">
        <div className="prose prose-p:my-2 prose-pre:my-2 max-w-none text-inherit">
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
        </div>
      </div>
    </div>
  )
}
