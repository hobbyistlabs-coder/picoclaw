import {
  IconChevronDown,
  IconChevronUp,
  IconDownload,
  IconPlus,
} from "@tabler/icons-react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import { AssistantMessage } from "@/components/chat/assistant-message"
import { ChatComposer } from "@/components/chat/chat-composer"
import { ChatEmptyState } from "@/components/chat/chat-empty-state"
import { ChatMetricsPills } from "@/components/chat/chat-metrics-pills"
import { ModelSelector } from "@/components/chat/model-selector"
import { PersonaSelector } from "@/components/chat/persona-selector"
import { SessionHistoryMenu } from "@/components/chat/session-history-menu"
import { TypingIndicator } from "@/components/chat/typing-indicator"
import { UserMessage } from "@/components/chat/user-message"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { useChatModels } from "@/hooks/use-chat-models"
import { useChatPersonas } from "@/hooks/use-chat-personas"
import { useGateway } from "@/hooks/use-gateway"
import { usePicoChat } from "@/hooks/use-pico-chat"
import { useSessionHistory } from "@/hooks/use-session-history"
import {
  buildConversationMarkdown,
  downloadConversationMarkdown,
} from "@/lib/chat-export"
import { getChatCostDebugInfo, hasChatMetrics } from "@/lib/chat-metrics"
import { getUserDisplayName } from "@/lib/user-profile"

const CHAT_PERSONA_STORAGE_KEY = "jane-ai:selected-persona-id"
const CHAT_HEADER_COLLAPSED_KEY = "jane-ai:chat-header-collapsed"

function readStoredPersonaId() {
  return localStorage.getItem(CHAT_PERSONA_STORAGE_KEY)?.trim() || "__auto__"
}

function writeStoredPersonaId(personaId: string) {
  localStorage.setItem(CHAT_PERSONA_STORAGE_KEY, personaId || "__auto__")
}

function readStoredHeaderCollapsed() {
  return localStorage.getItem(CHAT_HEADER_COLLAPSED_KEY) === "true"
}

function writeStoredHeaderCollapsed(collapsed: boolean) {
  localStorage.setItem(CHAT_HEADER_COLLAPSED_KEY, String(collapsed))
}

export function ChatPage() {
  const { t } = useTranslation()
  const scrollRef = useRef<HTMLDivElement>(null)
  const [isAtBottom, setIsAtBottom] = useState(true)
  const [input, setInput] = useState("")
  const [selectedPersonaId, setSelectedPersonaId] =
    useState(readStoredPersonaId)

  const [isHeaderCollapsed, setIsHeaderCollapsed] = useState(
    readStoredHeaderCollapsed,
  )

  const userDisplayName = getUserDisplayName()
  const personas = useChatPersonas()

  const {
    messages,
    sessionMetrics,
    isTyping,
    activeSessionId,
    sendMessage,
    switchSession,
    newChat,
  } = usePicoChat()

  const { state: gwState } = useGateway()
  const isConnected = gwState === "running"

  const {
    defaultModelName,
    defaultModel,
    hasAnyModels,
    hasConfiguredModels,
    apiKeyModels,
    oauthModels,
    localModels,
    handleSetDefault,
  } = useChatModels({ isConnected })

  const {
    sessions,
    hasMore,
    loadError,
    loadErrorMessage,
    observerRef,
    loadSessions,
    handleDeleteSession,
  } = useSessionHistory({
    activeSessionId,
    onDeletedActiveSession: newChat,
  })

  const handleScroll = (e: React.UIEvent<HTMLDivElement>) => {
    const { scrollTop, scrollHeight, clientHeight } = e.currentTarget
    setIsAtBottom(scrollHeight - scrollTop <= clientHeight + 10)
  }

  useEffect(() => {
    if (isAtBottom && scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [messages, isTyping, isAtBottom])

  useEffect(() => {
    const debug = getChatCostDebugInfo(sessionMetrics, defaultModel)
    console.debug("[chat-cost-debug]", {
      sessionId: activeSessionId,
      selectedModelName: defaultModelName,
      selectedModelRuntimeId: defaultModel?.model ?? null,
      pricing: defaultModel
        ? {
            price_per_m_token: defaultModel.price_per_m_token ?? null,
            input_price_per_m_token:
              defaultModel.input_price_per_m_token ?? null,
            output_price_per_m_token:
              defaultModel.output_price_per_m_token ?? null,
          }
        : null,
      metrics: sessionMetrics
        ? {
            prompt_tokens: sessionMetrics.prompt_tokens,
            completion_tokens: sessionMetrics.completion_tokens,
            total_tokens: sessionMetrics.total_tokens,
            estimated_cost_usd: sessionMetrics.estimated_cost_usd ?? null,
            has_estimated_cost: sessionMetrics.has_estimated_cost ?? false,
          }
        : null,
      debug,
      failureChecks: {
        missingTokenCounts: !sessionMetrics || sessionMetrics.total_tokens <= 0,
        missingSelectedModelPricing:
          !defaultModel ||
          ((defaultModel.input_price_per_m_token ?? 0) <= 0 &&
            (defaultModel.output_price_per_m_token ?? 0) <= 0 &&
            (defaultModel.price_per_m_token ?? 0) <= 0),
        potentialModelMismatch: Boolean(
          sessionMetrics &&
          sessionMetrics.total_tokens > 0 &&
          defaultModelName &&
          !defaultModel,
        ),
      },
    })
  }, [activeSessionId, defaultModel, defaultModelName, sessionMetrics])

  const effectivePersonaId =
    selectedPersonaId === "__auto__" ||
    personas.some((p) => p.id === selectedPersonaId)
      ? selectedPersonaId
      : "__auto__"

  const handleSend = () => {
    if (!input.trim() || !isConnected || !defaultModelName) return
    sendMessage(input.trim(), effectivePersonaId)
    setInput("")
  }

  useEffect(() => {
    writeStoredPersonaId(effectivePersonaId)
  }, [effectivePersonaId])

  useEffect(() => {
    writeStoredHeaderCollapsed(isHeaderCollapsed)
  }, [isHeaderCollapsed])

  const selectedPersonaLabel =
    effectivePersonaId === "__auto__"
      ? t("chat.persona.auto")
      : personas.find((persona) => persona.id === effectivePersonaId)?.label ||
        t("chat.persona.auto")

  const statusTone = isConnected
    ? "border-primary/30 bg-primary/10 text-primary"
    : "border-border bg-secondary/60 text-secondary-foreground"
  const statusLabel = isConnected
    ? t("chat.status.connected")
    : t("chat.status.disconnected")
  const statusDescription = !hasConfiguredModels
    ? t("chat.empty.noConfiguredModelDescription")
    : !defaultModelName
      ? t("chat.empty.noSelectedModelDescription")
      : isConnected
        ? t("chat.status.connectedDescription", { model: defaultModelName })
        : t("chat.empty.notRunningDescription")

  return (
    <div className="relative flex h-full flex-col bg-transparent">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top,_rgba(255,157,66,0.08),transparent_24%),radial-gradient(circle_at_left,_rgba(94,234,212,0.08),transparent_18%)]" />

      {/* Drawer Pull Toggle Button */}
      <div className="absolute top-0 left-1/2 z-50 -translate-x-1/2">
        <button
          onClick={() => setIsHeaderCollapsed((prev) => !prev)}
          className="group flex h-5 w-16 items-center justify-center rounded-b-xl border border-t-0 border-white/10 bg-[#0c0c0e]/80 text-white/30 backdrop-blur-xl transition-all hover:bg-white/10 hover:text-white"
          aria-label="Toggle header"
        >
          {isHeaderCollapsed ? (
            <IconChevronDown className="size-3.5 transition-transform group-hover:translate-y-0.5" />
          ) : (
            <IconChevronUp className="size-3.5 transition-transform group-hover:-translate-y-0.5" />
          )}
        </button>
      </div>

      <div className="flex flex-col">
        {!isHeaderCollapsed ? (
          /* --- EXPANDED HEADER --- */
          <div className="animate-in fade-in slide-in-from-top-2 flex shrink-0 flex-col duration-200">
            <PageHeader
              title={t("navigation.chat")}
              titleExtra={
                hasAnyModels && (
                  <div className="flex items-center gap-1.5 rounded-full border border-white/[0.05] bg-white/[0.02] p-1 shadow-inner backdrop-blur-xl">
                    <PersonaSelector
                      personas={personas}
                      value={selectedPersonaId}
                      onValueChange={setSelectedPersonaId}
                    />
                    <div className="h-3 w-[1px] bg-white/10" />
                    <ModelSelector
                      defaultModelName={defaultModelName}
                      apiKeyModels={apiKeyModels}
                      oauthModels={oauthModels}
                      localModels={localModels}
                      onValueChange={handleSetDefault}
                    />
                  </div>
                )
              }
            >
              <div className="flex items-center gap-1.5">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={newChat}
                  className="group h-9 gap-2 rounded-full border-indigo-500/20 bg-indigo-500/10 px-4 text-[11px] font-bold tracking-wider text-indigo-400 uppercase backdrop-blur-md transition-all hover:scale-[1.02] hover:border-indigo-500/40 hover:bg-indigo-500/20"
                >
                  <IconPlus className="size-3.5 transition-transform group-hover:rotate-90" />
                  <span className="hidden sm:inline">{t("chat.newChat")}</span>
                </Button>

                <div className="flex items-center gap-1 rounded-full border border-white/[0.05] bg-white/[0.03] p-1">
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7 rounded-full text-white/40 transition-all hover:bg-white/10 hover:text-white disabled:opacity-20"
                    onClick={() =>
                      downloadConversationMarkdown(
                        buildConversationMarkdown(messages, userDisplayName),
                        activeSessionId,
                      )
                    }
                    disabled={messages.length === 0}
                    title={t("chat.export")}
                  >
                    <IconDownload className="size-3.5" />
                  </Button>

                  <div className="h-3 w-[1px] bg-white/5" />

                  <SessionHistoryMenu
                    sessions={sessions}
                    activeSessionId={activeSessionId}
                    hasMore={hasMore}
                    loadError={loadError}
                    loadErrorMessage={loadErrorMessage}
                    observerRef={observerRef}
                    onOpenChange={(open) => open && void loadSessions(true)}
                    onSwitchSession={switchSession}
                    onDeleteSession={handleDeleteSession}
                  />
                </div>

                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => setIsHeaderCollapsed(true)}
                  className="h-8 w-8 rounded-full text-white/20 hover:bg-white/5 hover:text-white"
                >
                  <IconChevronUp className="size-4" />
                </Button>
              </div>
            </PageHeader>

            <div className="px-4 pb-4 md:px-8 lg:px-24 xl:px-48">
              <div className="group relative mx-auto w-full max-w-[1120px]">
                <div className="absolute -inset-[1px] rounded-2xl bg-gradient-to-r from-indigo-500/20 to-purple-500/20 opacity-75 blur transition duration-1000 group-hover:opacity-100 group-hover:duration-200" />
                <div
                  className={`relative flex flex-col gap-3 rounded-2xl border border-white/10 bg-[#0c0c0e]/80 px-5 py-3 shadow-2xl backdrop-blur-xl md:flex-row md:items-center md:justify-between ${statusTone}`}
                >
                  <div className="min-w-0">
                    <div className="flex flex-col gap-1 sm:flex-row sm:items-center sm:gap-3">
                      <div className="flex flex-shrink-0 items-center gap-2">
                        <span className="relative flex h-2 w-2">
                          <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-indigo-400 opacity-75" />
                          <span className="relative inline-flex h-2 w-2 rounded-full bg-indigo-500" />
                        </span>
                        <p className="text-[10px] font-bold tracking-[0.2em] text-white/50 uppercase">
                          {statusLabel}
                        </p>
                      </div>
                      <div className="hidden h-3 w-px bg-white/10 sm:block" />
                      <p className="truncate text-sm font-medium text-slate-300 antialiased">
                        {statusDescription}
                      </p>
                    </div>
                  </div>
                  <div className="flex flex-shrink-0 flex-wrap gap-2 text-[11px] font-medium tracking-tight">
                    {[
                      t("chat.status.sessions", { count: sessions.length }),
                      defaultModelName || t("chat.noModel"),
                      t("chat.persona.label", {
                        persona: selectedPersonaLabel,
                      }),
                    ].map((stat, i) => (
                      <span
                        key={i}
                        className="flex cursor-default items-center justify-center rounded-full border border-white/[0.08] bg-white/[0.03] px-3 py-1 text-slate-400 transition-all hover:bg-white/[0.07] hover:text-white"
                      >
                        {stat}
                      </span>
                    ))}
                  </div>
                </div>
              </div>
            </div>

            {hasChatMetrics(sessionMetrics) && (
              <div className="sticky top-0 z-10 border-b border-white/[0.05] bg-[#0c0c0e]/60 px-4 py-2 backdrop-blur-md md:px-8 lg:px-24 xl:px-48">
                <div className="mx-auto flex max-w-[1120px] items-center gap-4">
                  <div className="flex shrink-0 items-center gap-2">
                    <div className="h-1 w-1 rounded-full bg-indigo-500/50 shadow-[0_0_8px_rgba(99,102,241,0.6)]" />
                    <span className="text-[10px] font-bold tracking-[0.25em] whitespace-nowrap text-white/40 uppercase">
                      {t("chat.metrics.session")}
                    </span>
                  </div>
                  <div className="hidden h-4 w-[1px] bg-white/10 sm:block" />
                  <div className="no-scrollbar flex-1 overflow-x-auto">
                    <div className="flex origin-left scale-[0.9] items-center">
                      <ChatMetricsPills
                        metrics={sessionMetrics}
                        pricing={defaultModel}
                      />
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>
        ) : (
          /* --- COLLAPSED MINI-HEADER --- */
          <div className="animate-in fade-in slide-in-from-top-1 sticky top-0 z-50 flex items-center justify-between border-b border-white/[0.05] bg-[#0c0c0e]/80 px-4 py-2 backdrop-blur-xl duration-300 md:px-8">
            <div className="flex items-center gap-3">
              <h2 className="text-[11px] font-bold tracking-widest text-white/40 uppercase">
                {t("navigation.chat")}
              </h2>
              <div className="h-3 w-px bg-white/10" />
              <p className="hidden max-w-[200px] truncate text-[11px] font-medium text-slate-400 antialiased lg:block">
                {statusDescription}
              </p>
            </div>

            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={newChat}
                className="group h-7 gap-1.5 rounded-full border-indigo-500/20 bg-indigo-500/5 px-3 text-[10px] font-bold tracking-tight text-indigo-400 uppercase transition-all hover:bg-indigo-500/15"
              >
                <IconPlus className="size-3 transition-transform group-hover:rotate-90" />
                <span>{t("chat.newChat")}</span>
              </Button>

              <div className="flex items-center gap-0.5 rounded-full border border-white/[0.05] bg-white/[0.02] p-0.5">
                <SessionHistoryMenu
                  sessions={sessions}
                  activeSessionId={activeSessionId}
                  hasMore={hasMore}
                  loadError={loadError}
                  loadErrorMessage={loadErrorMessage}
                  observerRef={observerRef}
                  onOpenChange={(open) => open && void loadSessions(true)}
                  onSwitchSession={switchSession}
                  onDeleteSession={handleDeleteSession}
                />
                <div className="mx-0.5 h-3 w-[1px] bg-white/5" />
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6 rounded-full text-white/30 hover:text-white"
                  onClick={() =>
                    downloadConversationMarkdown(
                      buildConversationMarkdown(messages, userDisplayName),
                      activeSessionId,
                    )
                  }
                  disabled={messages.length === 0}
                >
                  <IconDownload className="size-3" />
                </Button>
              </div>

              <Button
                variant="ghost"
                size="icon"
                onClick={() => setIsHeaderCollapsed(false)}
                className="h-8 w-8 rounded-full text-indigo-400/50 hover:bg-indigo-500/10 hover:text-indigo-400"
              >
                <IconChevronDown className="size-4" />
              </Button>
            </div>
          </div>
        )}
      </div>

      <div
        ref={scrollRef}
        onScroll={handleScroll}
        className="min-h-0 flex-1 overflow-y-auto px-4 py-6 md:px-8 lg:px-24 xl:px-48"
      >
        <div className="mx-auto flex w-full max-w-[1120px] flex-col gap-8 pb-8">
          {messages.length === 0 && !isTyping && (
            <ChatEmptyState
              hasConfiguredModels={hasConfiguredModels}
              defaultModelName={defaultModelName}
              isConnected={isConnected}
              onPromptSelect={setInput}
            />
          )}

          {messages.map((msg) => (
            <div key={msg.id} className="flex w-full">
              {msg.role === "assistant" ? (
                <AssistantMessage
                  content={msg.content}
                  metrics={msg.metrics}
                  pricing={defaultModel}
                  timestamp={msg.timestamp}
                  reasoningContent={msg.reasoningContent}
                  toolEvents={msg.toolEvents}
                  pending={msg.pending}
                />
              ) : (
                <UserMessage
                  content={msg.content}
                  displayName={userDisplayName}
                />
              )}
            </div>
          ))}

          {isTyping && <TypingIndicator />}
        </div>
      </div>

      <ChatComposer
        input={input}
        onInputChange={setInput}
        onSend={handleSend}
        isConnected={isConnected}
        hasDefaultModel={Boolean(defaultModelName)}
      />
    </div>
  )
}
