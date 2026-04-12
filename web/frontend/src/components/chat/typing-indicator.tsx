import { IconSparkles } from "@tabler/icons-react"
import { useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

export function TypingIndicator() {
  const { t } = useTranslation()

  // Memoize the array so it isn't recreated on every single render cycle
  const thinkingSteps = useMemo(
    () => [
      t("chat.thinking.step1"),
      t("chat.thinking.step2"),
      t("chat.thinking.step3"),
      t("chat.thinking.step4"),
    ],
    [t],
  )

  const [stepIndex, setStepIndex] = useState(0)

  useEffect(() => {
    const interval = setInterval(() => {
      setStepIndex((prev) => (prev + 1) % thinkingSteps.length)
    }, 3000)

    return () => clearInterval(interval)
  }, [thinkingSteps.length])

  return (
    <div className="flex w-full flex-col gap-2">
      {/* Sender Header */}
      <div className="text-muted-foreground/80 flex items-center gap-1.5 px-1 text-xs font-medium">
        <IconSparkles className="size-3.5 text-violet-400" />
        <span>Jane AI</span>
      </div>

      {/* Main Indicator Bubble */}
      <div className="border-border/50 bg-card/60 inline-flex w-fit max-w-sm flex-col gap-3 rounded-2xl rounded-tl-sm border px-5 py-4 shadow-sm backdrop-blur-md">
        {/* Top Row: Bouncing Dots & Status Text */}
        <div className="flex items-center gap-3">
          <div className="flex flex-shrink-0 items-center gap-1">
            <span className="size-1.5 animate-bounce rounded-full bg-violet-400/80 [animation-delay:-0.3s]" />
            <span className="size-1.5 animate-bounce rounded-full bg-violet-400/80 [animation-delay:-0.15s]" />
            <span className="size-1.5 animate-bounce rounded-full bg-violet-400/80" />
          </div>

          <div className="relative h-4 overflow-hidden">
            <p
              key={stepIndex}
              className="animate-in fade-in slide-in-from-bottom-2 text-muted-foreground fill-mode-both text-xs font-medium duration-300 ease-out"
              role="status"
              aria-live="polite"
            >
              {thinkingSteps[stepIndex]}
            </p>
          </div>
        </div>

        {/* Bottom Row: Shimmer / Progress Bar */}
        <div className="bg-muted/50 relative h-1 w-36 overflow-hidden rounded-full">
          {/* Note: Kept your custom shimmer class, but added a generic pulse fallback just in case */}
          <div className="fallback:animate-pulse absolute inset-0 animate-[shimmer_2s_infinite] rounded-full bg-gradient-to-r from-violet-500/0 via-violet-400/60 to-violet-500/0 bg-[length:200%_100%]" />
        </div>
      </div>
    </div>
  )
}
