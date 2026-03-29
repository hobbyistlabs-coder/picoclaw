import {
  IconPlugConnectedX,
  IconRobotOff,
  IconSparkles,
  IconStar,
} from "@tabler/icons-react"
import { Link } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"

interface ChatEmptyStateProps {
  hasConfiguredModels: boolean
  defaultModelName: string
  isConnected: boolean
  onPromptSelect: (prompt: string) => void
}

export function ChatEmptyState({
  hasConfiguredModels,
  defaultModelName,
  isConnected,
  onPromptSelect,
}: ChatEmptyStateProps) {
  const { t } = useTranslation()
  const starterPrompts = [
    t("chat.starters.plan"),
    t("chat.starters.review"),
    t("chat.starters.route"),
  ]

  if (!hasConfiguredModels) {
    return (
      <div className="flex flex-col items-center justify-center py-20 opacity-70">
        <div className="mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-amber-500/10 text-amber-500">
          <IconRobotOff className="h-8 w-8" />
        </div>
        <h3 className="mb-2 text-xl font-medium">
          {t("chat.empty.noConfiguredModel")}
        </h3>
        <p className="text-muted-foreground mb-4 text-center text-sm">
          {t("chat.empty.noConfiguredModelDescription")}
        </p>
        <Button asChild variant="secondary" size="sm" className="px-4">
          <Link to="/models">{t("chat.empty.goToModels")}</Link>
        </Button>
      </div>
    )
  }

  if (!defaultModelName) {
    return (
      <div className="flex flex-col items-center justify-center py-20 opacity-70">
        <div className="mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-amber-500/10 text-amber-500">
          <IconStar className="h-8 w-8" />
        </div>
        <h3 className="mb-2 text-xl font-medium">
          {t("chat.empty.noSelectedModel")}
        </h3>
        <p className="text-muted-foreground mb-4 text-center text-sm">
          {t("chat.empty.noSelectedModelDescription")}
        </p>
      </div>
    )
  }

  if (!isConnected) {
    return (
      <div className="flex flex-col items-center justify-center py-20 opacity-70">
        <div className="mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-amber-500/10 text-amber-500">
          <IconPlugConnectedX className="h-8 w-8" />
        </div>
        <h3 className="mb-2 text-xl font-medium">
          {t("chat.empty.notRunning")}
        </h3>
        <p className="text-muted-foreground mb-4 text-center text-sm">
          {t("chat.empty.notRunningDescription")}
        </p>
      </div>
    )
  }

  return (
    <div className="relative overflow-hidden rounded-[2rem] border border-white/10 bg-white/6 px-6 py-14 shadow-2xl shadow-black/25">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_left,_rgba(116,227,213,0.18),_transparent_30%),radial-gradient(circle_at_bottom_right,_rgba(242,196,109,0.14),_transparent_28%)]" />
      <div className="relative flex flex-col items-center justify-center">
        <img className="mb-6 size-20" src="/jane-mark.svg" alt="" />
        <p className="mb-3 font-mono text-[11px] tracking-[0.38em] text-[#74e3d5] uppercase">
          Strategic Mesh Intelligence
        </p>
        <h3 className="mb-3 font-serif text-3xl font-semibold tracking-[0.08em] text-white">
          {t("chat.welcome")}
        </h3>
        <p className="max-w-2xl text-center text-sm leading-7 text-white/70">
          {t("chat.welcomeDesc")}
        </p>
        <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
          {starterPrompts.map((prompt) => (
            <Button
              key={prompt}
              variant="outline"
              className="border-white/14 bg-black/15 text-white/82 hover:border-[#74e3d5]/40 hover:bg-[#74e3d5]/12 hover:text-white"
              onClick={() => onPromptSelect(prompt)}
            >
              <IconSparkles className="size-4" />
              {prompt}
            </Button>
          ))}
        </div>
      </div>
    </div>
  )
}
