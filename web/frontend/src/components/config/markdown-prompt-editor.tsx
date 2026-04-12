import { IconEye, IconFileCode, IconHistory, IconPencil } from "@tabler/icons-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import ReactMarkdown from "react-markdown"
import remarkGfm from "remark-gfm"

import type { PromptRevision } from "@/api/prompt-history"
import { CodeBlock } from "@/components/chat/code-block"
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"

type ViewMode = "edit" | "preview" | "source"

interface MarkdownPromptEditorProps {
  value: string
  onChange: (value: string) => void
  placeholder: string
  disabled?: boolean
  revisions?: PromptRevision[]
  historyLoading?: boolean
}

export function MarkdownPromptEditor({
  value,
  onChange,
  placeholder,
  disabled,
  revisions = [],
  historyLoading,
}: MarkdownPromptEditorProps) {
  const { t } = useTranslation()
  const [mode, setMode] = useState<ViewMode>("edit")

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap gap-2">
        <Button
          variant={mode === "edit" ? "default" : "outline"}
          size="sm"
          onClick={() => setMode("edit")}
          disabled={disabled}
        >
          <IconPencil className="size-4" />
          {t("pages.config.workspace_files.edit")}
        </Button>
        <Button
          variant={mode === "preview" ? "default" : "outline"}
          size="sm"
          onClick={() => setMode("preview")}
          disabled={disabled}
        >
          <IconEye className="size-4" />
          {t("pages.config.workspace_files.preview")}
        </Button>
        <Button
          variant={mode === "source" ? "default" : "outline"}
          size="sm"
          onClick={() => setMode("source")}
          disabled={disabled}
        >
          <IconFileCode className="size-4" />
          {t("pages.config.workspace_files.highlighted")}
        </Button>
      </div>

      {mode === "edit" ? (
        <Textarea
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="min-h-[240px] resize-y font-mono text-sm"
          placeholder={placeholder}
          disabled={disabled}
        />
      ) : null}

      {mode === "preview" ? (
        <div className="bg-card/70 prose prose-stone dark:prose-invert max-w-none rounded-xl border p-5">
          <ReactMarkdown remarkPlugins={[remarkGfm]}>
            {value || t("pages.config.workspace_files.empty")}
          </ReactMarkdown>
        </div>
      ) : null}

      {mode === "source" ? (
        <CodeBlock code={value || "# Empty\n"} language="markdown" />
      ) : null}

      <div className="rounded-xl border p-3">
        <div className="mb-3 flex items-center gap-2 text-sm font-medium">
          <IconHistory className="size-4" />
          {t("pages.config.revision_history")}
        </div>
        {historyLoading ? (
          <div className="text-muted-foreground text-sm">{t("labels.loading")}</div>
        ) : revisions.length === 0 ? (
          <div className="text-muted-foreground text-sm">
            {t("pages.config.no_revisions")}
          </div>
        ) : (
          <div className="space-y-2">
            {revisions.map((revision) => (
              <div
                key={revision.id}
                className="flex flex-wrap items-center justify-between gap-2 rounded-lg border px-3 py-2"
              >
                <div className="text-muted-foreground text-xs">
                  {new Date(revision.timestamp).toLocaleString()}
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onChange(revision.content)}
                  disabled={disabled}
                >
                  {t("pages.config.load_revision")}
                </Button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
