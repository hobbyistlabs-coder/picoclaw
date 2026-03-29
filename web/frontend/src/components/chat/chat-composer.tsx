import { IconArrowUp } from "@tabler/icons-react"
import type { KeyboardEvent } from "react"
import { useTranslation } from "react-i18next"
import TextareaAutosize from "react-textarea-autosize"

import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

interface ChatComposerProps {
  input: string
  onInputChange: (value: string) => void
  onSend: () => void
  isConnected: boolean
  hasDefaultModel: boolean
}

export function ChatComposer({
  input,
  onInputChange,
  onSend,
  isConnected,
  hasDefaultModel,
}: ChatComposerProps) {
  const { t } = useTranslation()
  const canInput = isConnected && hasDefaultModel
  const disabledReason = !isConnected
    ? t("chat.composer.gatewayRequired")
    : !hasDefaultModel
      ? t("chat.composer.modelRequired")
      : t("chat.composer.ready")

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.nativeEvent.isComposing) return
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault()
      onSend()
    }
  }

  return (
    <div className="shrink-0 px-4 pt-4 pb-[calc(1rem+env(safe-area-inset-bottom))] md:px-8 md:pb-8 lg:px-24 xl:px-48">
      <div className="border-border/80 bg-card/90 mx-auto flex max-w-[1000px] flex-col rounded-[1.75rem] border p-3 shadow-xl shadow-black/25 backdrop-blur">
        <TextareaAutosize
          value={input}
          onChange={(e) => onInputChange(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={t("chat.placeholder")}
          disabled={!canInput}
          className={cn(
            "max-h-[200px] min-h-[60px] resize-none border-0 bg-transparent px-2 py-1 text-[15px] shadow-none transition-colors focus-visible:ring-0 focus-visible:outline-none dark:bg-transparent",
            !canInput && "cursor-not-allowed",
          )}
          minRows={1}
          maxRows={8}
        />

        <div className="mt-3 flex items-center justify-between gap-3 px-1">
          <div className="min-w-0">
            <p className="text-primary text-[11px] font-medium tracking-[0.22em] uppercase">
              {disabledReason}
            </p>
            <p className="text-muted-foreground mt-1 text-xs">
              {t("chat.composer.shortcut")}
            </p>
          </div>

          <Button
            size="icon"
            className="bg-primary text-primary-foreground hover:bg-primary/85 size-8 rounded-full transition-transform active:scale-95"
            onClick={onSend}
            disabled={!input.trim() || !canInput}
          >
            <IconArrowUp className="size-4" />
          </Button>
        </div>
      </div>
    </div>
  )
}
