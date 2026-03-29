import {
  IconLanguage,
  IconLoader2,
  IconMenu2,
  IconMoon,
  IconPlayerPlay,
  IconPower,
  IconSun,
} from "@tabler/icons-react"
import { Link } from "@tanstack/react-router"
import * as React from "react"
import { useTranslation } from "react-i18next"

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog.tsx"
import { Button } from "@/components/ui/button.tsx"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu.tsx"
import { Separator } from "@/components/ui/separator.tsx"
import { SidebarTrigger } from "@/components/ui/sidebar"
import { useGateway } from "@/hooks/use-gateway.ts"
import { useTheme } from "@/hooks/use-theme.ts"

export function AppHeader() {
  const { i18n, t } = useTranslation()
  const { theme, toggleTheme } = useTheme()
  const {
    state: gwState,
    loading: gwLoading,
    canStart,
    start,
    stop,
  } = useGateway()

  const isRunning = gwState === "running"
  const isStarting = gwState === "starting"
  const isStopped = gwState === "stopped" || gwState === "unknown"
  const gatewayTone = isRunning
    ? "border-emerald-400/20 bg-emerald-400/10 text-emerald-100"
    : isStarting
      ? "border-sky-400/20 bg-sky-400/10 text-sky-100"
      : "border-amber-400/20 bg-amber-400/10 text-amber-100"
  const gatewayLabel = isRunning
    ? t("header.gateway.label.running")
    : isStarting
      ? t("header.gateway.label.starting")
      : t("header.gateway.label.offline")
  const showNotConnectedHint =
    canStart && (gwState === "stopped" || gwState === "error")

  const [showStopDialog, setShowStopDialog] = React.useState(false)

  const handleGatewayToggle = () => {
    if (gwLoading || (!isRunning && !canStart)) return
    if (isRunning) {
      setShowStopDialog(true)
    } else {
      start()
    }
  }

  const confirmStop = () => {
    setShowStopDialog(false)
    stop()
  }

  return (
    <header className="bg-background/95 supports-backdrop-filter:bg-background/60 border-b-border/50 sticky top-0 z-50 flex h-14 shrink-0 items-center justify-between border-b px-4 backdrop-blur">
      <div className="flex items-center gap-2">
        <SidebarTrigger className="text-muted-foreground hover:bg-accent hover:text-foreground flex h-9 w-9 items-center justify-center rounded-lg sm:hidden [&>svg]:size-5">
          <IconMenu2 />
        </SidebarTrigger>
        <div className="flex shrink-0 items-center">
          <Link to="/">
            <img
              className="h-10 w-auto"
              src="/jane-wordmark.svg"
              alt="JANE-ai"
            />
          </Link>
        </div>
      </div>

      {/* Center prominent connection status */}
      <div className="pointer-events-none absolute left-1/2 hidden h-full -translate-x-1/2 items-center justify-center lg:flex">
        <div
          className={`flex items-center gap-2 rounded-full border px-4 py-1.5 text-xs shadow-sm backdrop-blur-md ${gatewayTone}`}
        >
          <span className="relative flex size-2 shrink-0 items-center justify-center rounded-full bg-current/50">
            {!isRunning ? (
              <span className="absolute inline-flex size-full animate-ping rounded-full bg-current opacity-70"></span>
            ) : null}
          </span>
          {gatewayLabel}
          {showNotConnectedHint ? (
            <span className="text-current/65">{t("chat.notConnected")}</span>
          ) : null}
        </div>
      </div>

      <AlertDialog open={showStopDialog} onOpenChange={setShowStopDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {t("header.gateway.stopDialog.title")}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {t("header.gateway.stopDialog.description")}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t("common.cancel")}</AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmStop}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {t("header.gateway.stopDialog.confirm")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <div className="text-muted-foreground flex items-center gap-1 text-sm font-medium md:gap-2">
        {/* Gateway Start/Stop */}
        <Button
          variant={isStarting ? "secondary" : "default"}
          size="sm"
          className={`h-8 gap-2 px-3 ${
            isRunning
              ? "bg-destructive/10 text-destructive hover:bg-destructive/20"
              : isStopped
                ? "bg-green-500 text-white hover:bg-green-600"
                : ""
          }`}
          onClick={handleGatewayToggle}
          disabled={gwLoading || isStarting || (!isRunning && !canStart)}
        >
          {gwLoading || isStarting ? (
            <IconLoader2 className="h-4 w-4 animate-spin opacity-70" />
          ) : isRunning ? (
            <IconPower className="h-4 w-4 opacity-80" />
          ) : (
            <IconPlayerPlay className="h-4 w-4 opacity-80" />
          )}
          <span className="text-xs font-semibold">
            {isRunning
              ? t("header.gateway.action.stop")
              : isStarting
                ? t("header.gateway.status.starting")
                : t("header.gateway.action.start")}
          </span>
        </Button>

        <Separator
          className="mx-4 my-2 hidden md:block"
          orientation="vertical"
        />

        {/* Language Switcher */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="size-8">
              <IconLanguage className="size-4.5" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => i18n.changeLanguage("en")}>
              English
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => i18n.changeLanguage("zh")}>
              简体中文
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>

        {/* Theme Toggle */}
        <Button
          variant="ghost"
          size="icon"
          className="size-8"
          onClick={toggleTheme}
        >
          {theme === "dark" ? (
            <IconSun className="size-4.5" />
          ) : (
            <IconMoon className="size-4.5" />
          )}
        </Button>
      </div>
    </header>
  )
}
