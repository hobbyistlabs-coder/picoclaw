import { IconWand } from "@tabler/icons-react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { getWorkspaceFileHistory } from "@/api/prompt-history"
import {
  type WorkspaceBootstrapFile,
  getWorkspaceBootstrapFiles,
  updateWorkspaceBootstrapFile,
} from "@/api/workspace"
import { MarkdownPromptEditor } from "@/components/config/markdown-prompt-editor"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

const ORDER = ["AGENTS.md", "IDENTITY.md", "SOUL.md", "USER.md"] as const

function normalizeMarkdown(content: string) {
  const normalized = content.replace(/\r\n/g, "\n").replace(/\r/g, "\n")
  return normalized && !normalized.endsWith("\n")
    ? `${normalized}\n`
    : normalized
}

function markdownLabel(name: string) {
  return name.replace(".md", "")
}

export function WorkspacePromptsSection() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [selectedName, setSelectedName] = useState<string>("AGENTS.md")

  // Only store the user's unsaved edits here.
  // We no longer need to sync the entire fetched payload into state.
  const [drafts, setDrafts] = useState<Record<string, string>>({})

  const { data, isLoading, error } = useQuery({
    queryKey: ["workspace", "bootstrap"],
    queryFn: getWorkspaceBootstrapFiles,
  })

  const selectedFile = useMemo(() => {
    return data?.files.find((file) => file.name === selectedName) ?? null
  }, [data, selectedName])

  const { data: historyData, isLoading: historyLoading } = useQuery({
    queryKey: ["workspace", "bootstrap", selectedName, "history"],
    queryFn: () => getWorkspaceFileHistory(selectedName),
    enabled: Boolean(selectedName),
  })

  // Derive the baseline and current content directly during render
  const currentBaseline = selectedFile?.content ?? ""
  // If there's a draft for this file, use it. Otherwise, fall back to the server data.
  const currentContent = drafts[selectedName] ?? currentBaseline
  const isDirty = currentContent !== currentBaseline

  const saveMutation = useMutation({
    mutationFn: async (file: WorkspaceBootstrapFile) =>
      // We know `drafts[file.name]` exists here because the save button is disabled otherwise
      updateWorkspaceBootstrapFile(file.name, drafts[file.name] ?? ""),
    onSuccess: (saved) => {
      toast.success(t("pages.config.workspace_files.save_success"))

      // Clear the local draft so the UI falls back to the newly fetched server data
      setDrafts((prev) => {
        const next = { ...prev }
        delete next[saved.name]
        return next
      })

      void queryClient.invalidateQueries({
        queryKey: ["workspace", "bootstrap"],
      })
    },
    onError: (err) => {
      toast.error(
        err instanceof Error
          ? err.message
          : t("pages.config.workspace_files.save_error"),
      )
    },
  })

  const handleReset = () => {
    // Simply delete the draft to revert to the baseline
    setDrafts((prev) => {
      const next = { ...prev }
      delete next[selectedName]
      return next
    })
  }

  return (
    <Card size="sm">
      <CardHeader className="border-border border-b">
        <CardTitle>{t("pages.config.sections.workspace_files")}</CardTitle>
        <CardDescription>
          {t("pages.config.workspace_files.description")}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4 pt-5">
        {isLoading ? (
          <div className="text-muted-foreground text-sm">
            {t("labels.loading")}
          </div>
        ) : error || !data ? (
          <div className="text-destructive text-sm">
            {t("pages.config.workspace_files.load_error")}
          </div>
        ) : (
          <>
            <div className="flex flex-wrap gap-2">
              {ORDER.map((name) => {
                const active = name === selectedName
                return (
                  <Button
                    key={name}
                    variant={active ? "default" : "outline"}
                    size="sm"
                    onClick={() => setSelectedName(name)}
                  >
                    {markdownLabel(name)}
                  </Button>
                )
              })}
            </div>

            <div className="flex flex-wrap items-center justify-between gap-3">
              <div className="text-muted-foreground text-xs">
                {selectedFile?.path ?? ""}
              </div>
            </div>

            <MarkdownPromptEditor
              value={currentContent}
              onChange={(value) =>
                setDrafts((prev) => ({ ...prev, [selectedName]: value }))
              }
              placeholder={t("pages.config.workspace_files.placeholder")}
              disabled={saveMutation.isPending}
              revisions={historyData?.revisions ?? []}
              historyLoading={historyLoading}
            />

            <div className="flex justify-end gap-2">
              <Button
                variant="outline"
                onClick={() =>
                  setDrafts((prev) => ({
                    ...prev,
                    [selectedName]: normalizeMarkdown(currentContent),
                  }))
                }
                disabled={saveMutation.isPending}
              >
                <IconWand className="size-4" />
                {t("pages.config.workspace_files.format")}
              </Button>
              <Button
                variant="outline"
                onClick={handleReset}
                disabled={!isDirty || saveMutation.isPending}
              >
                {t("common.reset")}
              </Button>
              <Button
                onClick={() =>
                  selectedFile && saveMutation.mutate(selectedFile)
                }
                disabled={!isDirty || saveMutation.isPending}
              >
                {saveMutation.isPending ? t("common.saving") : t("common.save")}
              </Button>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  )
}
