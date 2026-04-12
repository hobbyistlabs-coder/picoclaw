import { IconArrowUp } from "@tabler/icons-react"
import { IconFile, IconPaperclip, IconUpload, IconX } from "@tabler/icons-react"
import type { KeyboardEvent } from "react"
import { type DragEvent, useRef, useState } from "react"
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
  attachments?: File[]
  onAddAttachments?: (files: File[]) => void
  onRemoveAttachment?: (index: number) => void
}

export function ChatComposer({
  input,
  onInputChange,
  onSend,
  isConnected,
  hasDefaultModel,
  attachments = [],
  onAddAttachments,
  onRemoveAttachment,
}: ChatComposerProps) {
  const { t } = useTranslation()
  const [isDragging, setIsDragging] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const canInput = isConnected && hasDefaultModel
  const disabledReason = !isConnected
    ? t("chat.composer.gatewayRequired")
    : !hasDefaultModel
      ? t("chat.composer.modelRequired")
      : t("chat.composer.ready")

  // --- Drag & Drop Handlers ---
  const handleDragOver = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    e.stopPropagation()
    if (canInput) setIsDragging(true)
  }

  const handleDragLeave = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragging(false)
  }

  const handleDrop = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragging(false)

    if (!canInput) return

    const files = Array.from(e.dataTransfer.files)
    if (files.length > 0 && onAddAttachments) {
      onAddAttachments(files)
    }
  }

  // --- Keyboard & Input Handlers ---
  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.nativeEvent.isComposing) return
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault()
      onSend()
    }
  }

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && onAddAttachments) {
      onAddAttachments(Array.from(e.target.files))
    }
    // Reset input so the same file can be selected again if removed
    if (fileInputRef.current) fileInputRef.current.value = ""
  }

  return (
    <div className="shrink-0 px-4 pt-4 pb-[calc(1rem+env(safe-area-inset-bottom))] md:px-8 md:pb-8 lg:px-24 xl:px-48">
      {/* Hidden File Input */}
      <input
        type="file"
        multiple
        className="hidden"
        ref={fileInputRef}
        onChange={handleFileSelect}
      />

      {/* Main Glassmorphic Container */}
      <div
        onDragOver={handleDragOver}
        onDragEnter={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        className={cn(
          "relative mx-auto flex max-w-[1120px] flex-col rounded-[2rem] border transition-all duration-200",
          "border-white/[0.05] bg-white/[0.02] shadow-inner backdrop-blur-xl",
          isDragging && "ring-2 ring-indigo-500/50", // Highlight when dragging
        )}
      >
        {/* Drag & Drop Overlay */}
        {isDragging && (
          <div className="absolute inset-0 z-20 flex flex-col items-center justify-center rounded-[2rem] border-2 border-dashed border-indigo-500/50 bg-indigo-500/10 backdrop-blur-md">
            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-indigo-500/20 text-indigo-500">
              <IconUpload className="h-6 w-6 animate-bounce" />
            </div>
            <p className="mt-2 text-[11px] font-bold tracking-wider text-indigo-500 uppercase">
              Drop files to attach
            </p>
          </div>
        )}

        {/* File Previews Track */}
        {attachments.length > 0 && (
          <div className="scrollbar-hide flex gap-2 overflow-x-auto p-4 pb-0">
            {attachments.map((file, index) => {
              const isImage = file.type.startsWith("image/")
              return (
                <div
                  key={`${file.name}-${index}`}
                  className="group relative flex shrink-0 items-center gap-2 rounded-full border border-white/[0.05] bg-white/[0.03] p-1.5 pr-4 shadow-inner backdrop-blur-xl transition-opacity hover:bg-white/[0.05]"
                >
                  {/* Thumbnail / Icon */}
                  <div className="flex h-8 w-8 shrink-0 items-center justify-center overflow-hidden rounded-full bg-white/5">
                    {isImage ? (
                      <img
                        src={URL.createObjectURL(file)}
                        alt={file.name}
                        className="h-full w-full object-cover"
                      />
                    ) : (
                      <IconFile className="h-4 w-4 text-white/60" />
                    )}
                  </div>

                  {/* Filename */}
                  <span className="max-w-[120px] truncate text-xs text-white/80">
                    {file.name}
                  </span>

                  {/* Remove Button */}
                  <button
                    onClick={() => onRemoveAttachment?.(index)}
                    className="absolute -top-1 -right-1 flex h-5 w-5 items-center justify-center rounded-full bg-indigo-500 text-white opacity-0 shadow-sm transition-all group-hover:opacity-100 hover:scale-110"
                  >
                    <IconX className="h-3 w-3" />
                  </button>
                </div>
              )
            })}
          </div>
        )}

        {/* Text Area */}
        <TextareaAutosize
          value={input}
          onChange={(e) => onInputChange(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={t("chat.placeholder")}
          disabled={!canInput}
          className={cn(
            "max-h-[220px] min-h-[80px] w-full resize-none border-0 bg-transparent px-6 py-4 text-[15px] leading-7 text-white shadow-none transition-colors focus-visible:ring-0 focus-visible:outline-none dark:bg-transparent",
            !canInput && "cursor-not-allowed",
          )}
          minRows={1}
          maxRows={8}
        />

        {/* Action Toolbar */}
        <div className="flex items-center justify-between gap-3 px-2 pb-2">
          <div className="flex items-center gap-1">
            {/* Attachment Button */}
            <button
              onClick={() => fileInputRef.current?.click()}
              disabled={!canInput}
              className="group flex h-9 w-9 items-center justify-center rounded-full text-white/40 transition-all hover:bg-white/10 hover:text-white disabled:opacity-20"
              title="Attach files"
            >
              <IconPaperclip className="h-5 w-5 transition-transform group-hover:scale-110" />
            </button>
          </div>

          <div className="flex items-center gap-3 pr-2">
            <div className="hidden text-right md:block">
              <p className="text-[11px] font-bold tracking-[0.2em] text-white/40 uppercase">
                {disabledReason}
              </p>
            </div>

            {/* Indigo Glass Send Button */}
            <Button
              size="icon"
              className={cn(
                "group relative h-10 w-10 rounded-full border border-indigo-500/20 bg-indigo-500/10 text-indigo-500 backdrop-blur-md transition-all",
                "hover:scale-[1.02] hover:border-indigo-400 hover:bg-indigo-500 hover:text-white hover:shadow-[0_0_15px_rgba(99,102,241,0.4)]",
                "active:scale-95 disabled:opacity-30 disabled:hover:scale-100 disabled:hover:bg-indigo-500/10 disabled:hover:text-indigo-500",
              )}
              onClick={onSend}
              disabled={
                (!input.trim() && attachments.length === 0) || !canInput
              }
            >
              <IconArrowUp className="h-5 w-5 transition-transform group-hover:rotate-90" />
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
