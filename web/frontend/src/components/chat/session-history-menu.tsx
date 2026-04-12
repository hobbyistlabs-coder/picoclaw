import { IconHistory, IconTrash } from "@tabler/icons-react"
import dayjs from "dayjs"
import type { RefObject } from "react"
import { useTranslation } from "react-i18next"

import type { SessionSummary } from "@/api/sessions"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { ScrollArea } from "@/components/ui/scroll-area"

interface SessionHistoryMenuProps {
  sessions: SessionSummary[]
  activeSessionId: string
  hasMore: boolean
  loadError: boolean
  loadErrorMessage: string
  observerRef: RefObject<HTMLDivElement | null>
  onOpenChange: (open: boolean) => void
  onSwitchSession: (sessionId: string) => void
  onDeleteSession: (sessionId: string) => void
}

export function SessionHistoryMenu({
  sessions,
  activeSessionId,
  hasMore,
  loadError,
  loadErrorMessage,
  observerRef,
  onOpenChange,
  onSwitchSession,
  onDeleteSession,
}: SessionHistoryMenuProps) {
  const { t } = useTranslation()

  return (
    <DropdownMenu onOpenChange={onOpenChange}>
      <DropdownMenuTrigger asChild>
        {/* Trigger: High-gloss glass effect */}
        <Button
          variant="outline"
          size="sm"
          className="group relative h-10 gap-2.5 rounded-xl border-white/[0.08] bg-white/[0.03] px-4 text-white/70 backdrop-blur-md transition-all hover:bg-white/[0.08] hover:text-white hover:ring-1 hover:ring-white/20 active:scale-95"
        >
          <IconHistory className="size-3.5 transition-transform group-hover:-rotate-12" />
          <span className="hidden text-xs font-semibold tracking-wide sm:inline">
            {t("chat.history")}
          </span>
        </Button>
      </DropdownMenuTrigger>

      <DropdownMenuContent
        align="end"
        sideOffset={8}
        className="w-[380px] overflow-hidden rounded-[1.5rem] border-white/10 bg-[#0c0c0e]/90 p-2 shadow-2xl backdrop-blur-2xl"
      >
        <div className="mb-2 px-3 pt-2 pb-1">
          <p className="text-[10px] font-bold tracking-[0.2em] text-white/30 uppercase">
            Recent Conversations
          </p>
        </div>

        <ScrollArea className="max-h-[420px] pr-1">
          {loadError && (
            <div className="p-4 text-center">
              <span className="text-destructive/80 text-[11px] italic">
                {loadErrorMessage}
              </span>
            </div>
          )}

          {sessions.length === 0 && !loadError ? (
            <div className="p-8 text-center">
              <span className="text-xs text-white/20">
                {t("chat.noHistory")}
              </span>
            </div>
          ) : (
            sessions.map((session) => {
              const isActive = session.id === activeSessionId
              return (
                <DropdownMenuItem
                  key={session.id}
                  className={`group relative mb-1.5 flex flex-col items-start gap-1 rounded-2xl border border-transparent px-4 py-3 transition-all focus:bg-white/[0.04] focus:text-white ${
                    isActive
                      ? "border-white/10 bg-white/[0.06]"
                      : "hover:bg-white/[0.02]"
                  }`}
                  onClick={() => onSwitchSession(session.id)}
                >
                  <div className="flex w-full items-center justify-between gap-2">
                    <span
                      className={`line-clamp-1 text-sm font-medium transition-colors ${isActive ? "text-indigo-400" : "text-white/90"}`}
                    >
                      {session.title || session.preview}
                    </span>

                    {session.agent_id && (
                      <span className="shrink-0 rounded-md border border-indigo-500/20 bg-indigo-500/10 px-1.5 py-0.5 text-[9px] font-bold tracking-wider text-indigo-400 uppercase">
                        {session.agent_id}
                      </span>
                    )}
                  </div>

                  {/* Secondary Info Line */}
                  <div className="flex items-center gap-2 text-[11px] text-white/40">
                    <span className="font-mono opacity-80">
                      {t("chat.messagesCount", {
                        count: session.message_count,
                      })}
                    </span>
                    <span className="text-[8px] opacity-30">|</span>
                    <span>{dayjs(session.updated).fromNow()}</span>
                  </div>

                  {/* Delete Button: Integrated more cleanly */}
                  <Button
                    variant="ghost"
                    size="icon"
                    className="absolute top-1/2 right-2 h-8 w-8 -translate-y-1/2 rounded-full bg-red-500/0 text-white/0 transition-all group-hover:text-white/20 hover:bg-red-500/10 hover:text-red-400 focus:text-red-400"
                    onClick={(e) => {
                      e.preventDefault()
                      e.stopPropagation()
                      onDeleteSession(session.id)
                    }}
                  >
                    <IconTrash className="h-3.5 w-3.5" />
                  </Button>

                  {/* Active Glow Indicator */}
                  {isActive && (
                    <div className="absolute top-1/2 left-1 h-6 w-[2px] -translate-y-1/2 rounded-full bg-indigo-500 shadow-[0_0_8px_rgba(99,102,241,0.8)]" />
                  )}
                </DropdownMenuItem>
              )
            })
          )}

          {hasMore && (
            <div ref={observerRef} className="flex justify-center py-4">
              <div className="h-1 w-12 animate-pulse rounded-full bg-white/10" />
            </div>
          )}
        </ScrollArea>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
