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

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.nativeEvent.isComposing) return
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault()
      onSend()
    }
  }

  return (
    <div className="shrink-0 px-4 pt-4 pb-[calc(1rem+env(safe-area-inset-bottom))] md:px-8 md:pb-8 lg:px-24 xl:px-48">
      <div className="border-border/80 mx-auto flex max-w-[1000px] flex-col rounded-[1.75rem] border bg-[#09131d]/92 p-3 shadow-xl shadow-black/25 backdrop-blur">
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

        <div className="mt-2 flex items-center justify-between px-1">
          <div className="flex items-center gap-1">{/* action buttons */}</div>

          <Button
            size="icon"
            className="size-8 rounded-full bg-[#74e3d5] text-[#07131b] transition-transform hover:bg-[#99efe4] active:scale-95"
            onClick={onSend}
            disabled={!input.trim() || !isConnected}
          >
            <IconArrowUp className="size-4" />
          </Button>
        </div>
      </div>
    </div>
  )
}
