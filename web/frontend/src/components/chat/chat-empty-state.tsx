import {
  IconArrowRight,
  IconPlugConnectedX,
  IconRobotOff,
  IconSparkles,
  IconStar,
} from "@tabler/icons-react"
import { Link } from "@tanstack/react-router"
import React from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"

interface StatusCardProps {
  icon: React.ElementType
  title: string
  description: string
  action?: React.ReactNode
}

// 1. COMPACTED: StatusCard padding and icon sizes reduced significantly
const StatusCard = ({
  icon: Icon,
  title,
  description,
  action,
}: StatusCardProps) => (
  <div className="group relative flex w-full max-w-lg flex-col items-center justify-center overflow-hidden rounded-[2.5rem] border border-white/5 bg-[#0c0c0e]/40 px-6 py-12 shadow-2xl backdrop-blur-xl transition-all hover:border-white/10">
    {/* Background Glow */}
    <div className="absolute inset-0 bg-[radial-gradient(circle_at_center,_rgba(99,102,241,0.08),transparent_70%)] transition-opacity group-hover:opacity-100" />

    <div className="relative mb-5 flex h-14 w-14 items-center justify-center rounded-[1.25rem] border border-white/10 bg-gradient-to-br from-white/10 to-transparent text-indigo-400 shadow-2xl backdrop-blur-md">
      <Icon className="h-7 w-7 stroke-[1.5]" />
      <div className="absolute inset-0 rounded-[1.25rem] bg-[linear-gradient(120deg,rgba(255,255,255,0.1)_0%,transparent_50%)]" />
    </div>

    <h3 className="relative mb-2 text-xl font-bold tracking-tight text-white/90">
      {title}
    </h3>
    <p className="relative mb-6 max-w-sm text-center text-xs leading-relaxed font-medium text-slate-400 antialiased">
      {description}
    </p>
    <div className="relative">{action}</div>
  </div>
)

// 2. FIXED: Moved outside the main component to prevent render-cycle destruction
const StatusWrapper = ({ children }: { children: React.ReactNode }) => (
  <div className="flex h-full items-center justify-center p-4">
    <div className="relative overflow-hidden rounded-[2rem] border border-white/[0.05] bg-[#0c0c0e]/60 px-6 py-6 shadow-2xl backdrop-blur-xl transition-all">
      <div className="absolute -top-24 -right-24 h-48 w-48 rounded-full bg-indigo-500/10 blur-[50px]" />
      {children}
    </div>
  </div>
)

interface ChatEmptyStateProps {
  hasConfiguredModels: boolean
  defaultModelName: string | null | undefined
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
      <StatusWrapper>
        <StatusCard
          icon={IconRobotOff}
          title={t("chat.empty.noConfiguredModel")}
          description={t("chat.empty.noConfiguredModelDescription")}
          action={
            <Button
              asChild
              variant="outline"
              className="group mt-2 h-9 gap-2 rounded-full border-indigo-500/20 bg-indigo-500/10 px-6 text-[10px] font-extrabold tracking-[0.2em] text-indigo-400 uppercase backdrop-blur-md transition-all hover:border-indigo-500/40 hover:bg-indigo-500/20"
            >
              <Link to="/models" className="flex items-center">
                {t("chat.empty.goToModels")}
                <IconArrowRight className="size-3.5 transition-transform group-hover:translate-x-1" />
              </Link>
            </Button>
          }
        />
      </StatusWrapper>
    )
  }

  if (!defaultModelName) {
    return (
      <StatusWrapper>
        <StatusCard
          icon={IconStar}
          title={t("chat.empty.noSelectedModel")}
          description={t("chat.empty.noSelectedModelDescription")}
        />
      </StatusWrapper>
    )
  }

  if (!isConnected) {
    return (
      <StatusWrapper>
        <StatusCard
          icon={IconPlugConnectedX}
          title={t("chat.empty.notRunning")}
          description={t("chat.empty.notRunningDescription")}
        />
      </StatusWrapper>
    )
  }

  return (
    <div className="flex h-full items-center justify-center p-4">
      {/* 3. COMPACTED: Dropped vertical padding from py-10/14 to py-8/10 */}
      <div className="group relative mx-auto w-full max-w-3xl overflow-hidden rounded-[2rem] border border-white/[0.04] bg-[#0c0c0e]/40 px-6 py-8 shadow-2xl backdrop-blur-2xl md:px-10 md:py-10">
        {/* Holographic Background Effects */}
        <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_50%_0%,_rgba(99,102,241,0.08),_transparent_50%),radial-gradient(circle_at_50%_100%,_rgba(244,63,94,0.03),_transparent_50%)] transition-opacity duration-1000 group-hover:opacity-100" />

        <div className="pointer-events-none absolute inset-0 [mask-image:radial-gradient(ellipse_at_center,black_20%,transparent_70%)] opacity-[0.02]">
          <div className="absolute inset-0 [background-image:linear-gradient(to_right,#fff_1px,transparent_1px),linear-gradient(to_bottom,#fff_1px,transparent_1px)] [background-size:32px_32px]" />
        </div>

        <div className="relative flex flex-col items-center justify-center text-center">
          {/* Avatar / Logo Area: Scaled down even further and margin cut down */}
          <div className="relative mb-4">
            <div className="absolute -inset-2 animate-pulse rounded-full bg-indigo-500/20 blur-lg" />
            <div className="relative flex h-12 w-12 items-center justify-center rounded-2xl border border-white/10 bg-[#0c0c0e]/80 p-2 shadow-xl ring-1 ring-white/5 backdrop-blur-xl transition-transform duration-700 ease-out group-hover:scale-105 group-hover:border-indigo-500/30">
              <img
                className="size-6 object-contain drop-shadow-[0_0_8px_rgba(99,102,241,0.6)]"
                src="/jane-mark.svg"
                alt="Jane AI"
              />
            </div>
          </div>

          {/* Tighter spacing between text elements */}
          <div className="space-y-2.5">
            {/* Telemetry Indicator */}
            <div className="inline-flex items-center gap-2 rounded-full border border-white/[0.05] bg-white/[0.02] px-2.5 py-0.5 shadow-inner backdrop-blur-md">
              <div className="relative flex h-1 w-1">
                <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75"></span>
                <span className="relative inline-flex h-1 w-1 rounded-full bg-emerald-500"></span>
              </div>
              <span className="text-[8px] font-bold tracking-[0.3em] text-white/50 uppercase">
                {t("chat.empty.systemReady") || "System Online"}
              </span>
            </div>

            {/* Typography: Slightly smaller again to fit the tighter box */}
            <h3 className="bg-gradient-to-br from-white via-white/90 to-white/40 bg-clip-text text-2xl font-black tracking-tight text-transparent sm:text-3xl">
              {t("chat.welcome")}
            </h3>

            <p className="mx-auto max-w-[26rem] text-[11px] leading-relaxed font-medium text-white/40 antialiased sm:text-xs">
              {t("chat.welcomeDesc")}
            </p>
          </div>

          {/* Quick-Action Pills: Pulled tight against the text */}
          <div className="mt-6 flex flex-wrap items-center justify-center gap-2">
            {starterPrompts.map((prompt) => (
              <Button
                key={prompt}
                variant="outline"
                className="group/btn relative h-8 gap-1.5 rounded-full border-white/[0.05] bg-white/[0.02] px-3.5 text-[10px] font-semibold text-white/60 backdrop-blur-md transition-all duration-300 hover:-translate-y-0.5 hover:border-indigo-500/30 hover:bg-indigo-500/10 hover:text-white hover:shadow-[0_10px_20px_-10px_rgba(99,102,241,0.4)]"
                onClick={() => onPromptSelect(prompt)}
              >
                <IconSparkles className="size-3 text-indigo-400/70 transition-all duration-300 group-hover/btn:rotate-12 group-hover/btn:text-indigo-400" />
                {prompt}
              </Button>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
