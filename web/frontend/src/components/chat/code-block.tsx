import Prism from "prismjs"
import "prismjs/components/prism-bash"
import "prismjs/components/prism-json"
import "prismjs/components/prism-jsx"
import "prismjs/components/prism-markdown"
import "prismjs/components/prism-tsx"
import "prismjs/components/prism-typescript"

import { cn } from "@/lib/utils"

interface CodeBlockProps {
  code: string
  language?: string
  className?: string
}

function normalizeLanguage(language?: string) {
  const raw = (language || "").toLowerCase()
  if (raw === "js" || raw === "javascript") return "javascript"
  if (raw === "ts" || raw === "typescript") return "typescript"
  if (raw === "tsx") return "tsx"
  if (raw === "jsx") return "jsx"
  if (raw === "sh" || raw === "shell" || raw === "bash" || raw === "zsh") {
    return "bash"
  }
  if (raw === "md" || raw === "markdown") return "markdown"
  if (raw === "json") return "json"
  return "json"
}

export function CodeBlock({ code, language, className }: CodeBlockProps) {
  const normalizedLanguage = normalizeLanguage(language)
  const grammar = Prism.languages[normalizedLanguage] || Prism.languages.json
  const html = Prism.highlight(code, grammar, normalizedLanguage)

  return (
    <div
      className={cn("code-block overflow-hidden rounded-xl border", className)}
    >
      <div className="border-b border-white/8 bg-black/25 px-3 py-2 text-[10px] font-semibold tracking-[0.24em] text-[rgb(var(--theme-cream-rgb))] uppercase">
        {normalizedLanguage}
      </div>
      <pre className="overflow-x-auto bg-[rgb(var(--theme-midnight-rgb))] p-4 text-xs leading-6 text-[rgb(var(--theme-cream-rgb))]">
        <code
          className={`language-${normalizedLanguage}`}
          dangerouslySetInnerHTML={{ __html: html }}
        />
      </pre>
    </div>
  )
}
