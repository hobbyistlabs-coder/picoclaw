import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useMemo, useState } from "react"
import { toast } from "sonner"

import {
  type Board,
  type BoardCard,
  createBoard,
  createBoardColumn,
  createCard,
  deleteCard,
  getBoard,
  getBoards,
  runCardAgent,
  updateBoardReview,
  updateCard,
} from "@/api/boards"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"

const BOARD_TEMPLATES = {
  default: [
    { key: "todo", name: "Todo" },
    { key: "in_progress", name: "In Progress" },
    { key: "done", name: "Done" },
  ],
  research: [
    { key: "research", name: "Research" },
    { key: "viability", name: "Viability" },
    { key: "implementation", name: "Implementation" },
    { key: "todo", name: "Todo" },
    { key: "not_doing", name: "Decided Not To Do" },
    { key: "done", name: "Done" },
  ],
} as const

// Column accent colors — cycles through indigo, violet, sky, emerald, amber, rose
const COLUMN_ACCENTS = [
  {
    dot: "bg-indigo-500/60",
    glow: "shadow-[0_0_8px_rgba(99,102,241,0.5)]",
    ring: "focus-within:border-indigo-500/30",
  },
  {
    dot: "bg-violet-500/60",
    glow: "shadow-[0_0_8px_rgba(139,92,246,0.5)]",
    ring: "focus-within:border-violet-500/30",
  },
  {
    dot: "bg-sky-500/60",
    glow: "shadow-[0_0_8px_rgba(14,165,233,0.5)]",
    ring: "focus-within:border-sky-500/30",
  },
  {
    dot: "bg-emerald-500/60",
    glow: "shadow-[0_0_8px_rgba(16,185,129,0.5)]",
    ring: "focus-within:border-emerald-500/30",
  },
  {
    dot: "bg-amber-500/60",
    glow: "shadow-[0_0_8px_rgba(245,158,11,0.5)]",
    ring: "focus-within:border-amber-500/30",
  },
  {
    dot: "bg-rose-500/60",
    glow: "shadow-[0_0_8px_rgba(244,63,94,0.5)]",
    ring: "focus-within:border-rose-500/30",
  },
]

export function BoardsPage() {
  const queryClient = useQueryClient()
  const [drafts, setDrafts] = useState<Record<string, string>>({})
  const [descs, setDescs] = useState<Record<string, string>>({})
  const [selectedBoardID, setSelectedBoardID] = useState<string>()
  const [boardName, setBoardName] = useState("")
  const [boardDescription, setBoardDescription] = useState("")
  const [boardTemplate, setBoardTemplate] =
    useState<keyof typeof BOARD_TEMPLATES>("research")
  const [columnName, setColumnName] = useState("")
  // Mobile: which column index is active
  const [mobileColIndex, setMobileColIndex] = useState(0)

  const boardsQuery = useQuery({ queryKey: ["boards"], queryFn: getBoards })
  const boardID = useMemo(() => {
    if (!boardsQuery.data?.length) return undefined
    return boardsQuery.data.some((b) => b.id === selectedBoardID)
      ? selectedBoardID
      : boardsQuery.data[0].id
  }, [boardsQuery.data, selectedBoardID])

  const boardQuery = useQuery({
    queryKey: ["boards", boardID],
    queryFn: () => getBoard(boardID ?? "default"),
    enabled: Boolean(boardID),
  })

  const invalidate = async (nextBoardID = boardID) => {
    await queryClient.invalidateQueries({ queryKey: ["boards"] })
    await queryClient.invalidateQueries({ queryKey: ["boards", nextBoardID] })
  }

  const refreshBoardSoon = () => {
    window.setTimeout(() => void invalidate(), 1500)
    window.setTimeout(() => void invalidate(), 5000)
    window.setTimeout(() => void invalidate(), 12000)
  }

  const createBoardMutation = useMutation({
    mutationFn: () =>
      createBoard({
        name: boardName,
        description: boardDescription,
        columns: [...BOARD_TEMPLATES[boardTemplate]],
      }),
    onSuccess: async (board) => {
      setBoardName("")
      setBoardDescription("")
      setSelectedBoardID(board.id)
      await invalidate(board.id)
    },
    onError: showError,
  })

  const addColumnMutation = useMutation({
    mutationFn: () =>
      createBoardColumn(requireBoardID(boardID), { name: columnName }),
    onSuccess: async () => {
      setColumnName("")
      await invalidate()
    },
    onError: showError,
  })

  const addMutation = useMutation({
    mutationFn: ({
      title,
      description,
      columnID,
    }: {
      title: string
      description: string
      columnID: string
    }) =>
      createCard(requireBoardID(boardID), {
        title,
        description,
        column_id: columnID,
      }),
    onSuccess: async (_, vars) => {
      setDrafts((v) => ({ ...v, [vars.columnID]: "" }))
      setDescs((v) => ({ ...v, [vars.columnID]: "" }))
      await invalidate()
    },
    onError: showError,
  })

  const moveMutation = useMutation({
    mutationFn: ({ cardID, columnID }: { cardID: string; columnID: string }) =>
      updateCard(requireBoardID(boardID), cardID, { column_id: columnID }),
    onSuccess: () => invalidate(),
    onError: showError,
  })

  const editMutation = useMutation({
    mutationFn: (card: BoardCard) =>
      updateCard(requireBoardID(boardID), card.id, {
        title: card.title,
        description: card.description,
      }),
    onSuccess: () => invalidate(),
    onError: showError,
  })

  const deleteMutation = useMutation({
    mutationFn: (cardID: string) => deleteCard(requireBoardID(boardID), cardID),
    onSuccess: () => invalidate(),
    onError: showError,
  })

  const runMutation = useMutation({
    mutationFn: (cardID: string) =>
      runCardAgent(requireBoardID(boardID), cardID),
    onSuccess: () => {
      toast.success("Agent run queued")
      refreshBoardSoon()
    },
    onError: showError,
  })

  const reviewMutation = useMutation({
    mutationFn: (enabled: boolean) =>
      updateBoardReview(requireBoardID(boardID), {
        enabled,
        every_minutes: enabled
          ? Math.max(boardQuery.data?.review?.every_minutes ?? 30, 5)
          : 0,
      }),
    onSuccess: () => invalidate(),
    onError: showError,
  })

  const intervalMutation = useMutation({
    mutationFn: (every: number) =>
      updateBoardReview(requireBoardID(boardID), {
        enabled: true,
        every_minutes: every,
      }),
    onSuccess: () => invalidate(),
    onError: showError,
  })

  const ordered = useMemo(
    () => boardQuery.data?.columns ?? [],
    [boardQuery.data?.columns],
  )

  // Clamp mobile index when columns change
  const safeMobileColIndex = Math.min(
    mobileColIndex,
    Math.max(0, ordered.length - 1),
  )

  return (
    <div className="flex h-full flex-col">
      <PageHeader title="Boards" />

      <div className="flex flex-1 flex-col overflow-hidden px-3 py-3 sm:px-4 sm:py-4 md:px-6 md:py-4 lg:px-8">
        <div className="mx-auto flex w-full max-w-[1600px] flex-1 flex-col gap-4 overflow-hidden">
          {/* ── Top Control Bar ── */}
          <div className="flex shrink-0 flex-col gap-3 rounded-2xl border border-white/[0.06] bg-white/[0.02] p-3 shadow-inner backdrop-blur-xl sm:rounded-[1.5rem] sm:p-4">
            {/* Row 1: Board selector + Review panel */}
            <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center sm:justify-between">
              <BoardSelector
                board={boardQuery.data}
                boards={boardsQuery.data ?? []}
                boardDescription={boardDescription}
                boardName={boardName}
                boardTemplate={boardTemplate}
                onBoardDescriptionChange={setBoardDescription}
                onBoardNameChange={setBoardName}
                onBoardTemplateChange={(v) =>
                  setBoardTemplate(v as keyof typeof BOARD_TEMPLATES)
                }
                onCreateBoard={() => createBoardMutation.mutate()}
                onSelectedBoardChange={setSelectedBoardID}
                selectedBoardID={boardID}
              />

              <ReviewPanel
                enabled={boardQuery.data?.review?.enabled ?? false}
                every={boardQuery.data?.review?.every_minutes ?? 30}
                busy={reviewMutation.isPending || intervalMutation.isPending}
                onSaveInterval={(every) => intervalMutation.mutate(every)}
                onToggle={(enabled) => reviewMutation.mutate(enabled)}
              />
            </div>

            {/* Row 2: Add column */}
            <div className="flex items-center gap-2 self-start rounded-full border border-white/[0.03] bg-black/20 p-1">
              <Input
                className="h-8 w-40 border-none bg-transparent px-3 text-xs text-white placeholder:text-white/30 focus-visible:ring-0 focus-visible:ring-offset-0 sm:w-48"
                onChange={(e) => setColumnName(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter" && columnName && boardID) {
                    addColumnMutation.mutate()
                  }
                }}
                placeholder="New column name..."
                value={columnName}
              />
              <Button
                size="sm"
                className="h-7 rounded-full bg-indigo-500/10 px-4 text-[10px] font-bold tracking-wider text-indigo-400 uppercase hover:bg-indigo-500/20 disabled:opacity-30"
                disabled={
                  !boardID || addColumnMutation.isPending || !columnName
                }
                onClick={() => {
                  addColumnMutation.mutate()
                  setColumnName("")
                }}
              >
                Add Column
              </Button>
            </div>
          </div>

          {/* ── Mobile Column Navigator ── */}
          {ordered.length > 0 && (
            <div className="no-scrollbar flex shrink-0 items-center gap-2 overflow-x-auto pb-0.5 sm:hidden">
              {ordered.map((col, i) => {
                const accent = COLUMN_ACCENTS[i % COLUMN_ACCENTS.length]
                const isActive = i === safeMobileColIndex
                return (
                  <button
                    key={col.id}
                    onClick={() => setMobileColIndex(i)}
                    className={[
                      "flex shrink-0 items-center gap-1.5 rounded-full border px-3 py-1.5 text-[10px] font-bold tracking-widest uppercase transition-all duration-200",
                      isActive
                        ? "border-white/20 bg-white/10 text-white"
                        : "border-white/[0.04] bg-white/[0.02] text-white/40 hover:text-white/70",
                    ].join(" ")}
                  >
                    <span
                      className={`h-1.5 w-1.5 rounded-full ${accent.dot}`}
                    />
                    <span className="max-w-[100px] truncate">{col.name}</span>
                    <span className="rounded-full bg-white/10 px-1.5 py-0.5 font-mono text-[9px] text-white/50">
                      {col.cards.length}
                    </span>
                  </button>
                )
              })}
            </div>
          )}

          {/* ── Kanban Area ── */}
          <div className="flex-1 overflow-hidden">
            {/* Desktop: horizontal scroll */}
            <div className="hidden h-full sm:block">
              <div className="no-scrollbar h-full overflow-x-auto overflow-y-hidden pb-4">
                <div
                  className="flex h-full gap-3 px-0.5"
                  style={{ minWidth: `${ordered.length * 300}px` }}
                >
                  {ordered.map((column, i) => (
                    <KanbanColumn
                      key={column.id}
                      column={column}
                      accentIndex={i}
                      columns={ordered}
                      draft={drafts[column.id] ?? ""}
                      desc={descs[column.id] ?? ""}
                      onDraftChange={(v) =>
                        setDrafts((p) => ({ ...p, [column.id]: v }))
                      }
                      onDescChange={(v) =>
                        setDescs((p) => ({ ...p, [column.id]: v }))
                      }
                      onAddCard={() =>
                        addMutation.mutate({
                          title: drafts[column.id] ?? "",
                          description: descs[column.id] ?? "",
                          columnID: column.id,
                        })
                      }
                      onDelete={(cardID) => deleteMutation.mutate(cardID)}
                      onMove={(cardID, columnID) =>
                        moveMutation.mutate({ cardID, columnID })
                      }
                      onRun={(cardID) => runMutation.mutate(cardID)}
                      onSave={editMutation.mutate}
                      runPending={runMutation.isPending}
                    />
                  ))}
                </div>
              </div>
            </div>

            {/* Mobile: single active column */}
            <div className="flex h-full flex-col sm:hidden">
              {ordered[safeMobileColIndex] &&
                (() => {
                  const column = ordered[safeMobileColIndex]
                  return (
                    <KanbanColumn
                      key={column.id}
                      column={column}
                      accentIndex={safeMobileColIndex}
                      columns={ordered}
                      draft={drafts[column.id] ?? ""}
                      desc={descs[column.id] ?? ""}
                      onDraftChange={(v) =>
                        setDrafts((p) => ({ ...p, [column.id]: v }))
                      }
                      onDescChange={(v) =>
                        setDescs((p) => ({ ...p, [column.id]: v }))
                      }
                      onAddCard={() =>
                        addMutation.mutate({
                          title: drafts[column.id] ?? "",
                          description: descs[column.id] ?? "",
                          columnID: column.id,
                        })
                      }
                      onDelete={(cardID) => deleteMutation.mutate(cardID)}
                      onMove={(cardID, columnID) =>
                        moveMutation.mutate({ cardID, columnID })
                      }
                      onRun={(cardID) => runMutation.mutate(cardID)}
                      onSave={editMutation.mutate}
                      runPending={runMutation.isPending}
                      isMobile
                    />
                  )
                })()}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

// ── KanbanColumn ────────────────────────────────────────────────────────────

function KanbanColumn({
  column,
  accentIndex,
  columns,
  draft,
  desc,
  isMobile = false,
  onDraftChange,
  onDescChange,
  onAddCard,
  onDelete,
  onMove,
  onRun,
  onSave,
  runPending,
}: {
  column: Board["columns"][number]
  accentIndex: number
  columns: Board["columns"]
  draft: string
  desc: string
  isMobile?: boolean
  onDraftChange: (v: string) => void
  onDescChange: (v: string) => void
  onAddCard: () => void
  onDelete: (cardID: string) => void
  onMove: (cardID: string, columnID: string) => void
  onRun: (cardID: string) => void
  onSave: (card: BoardCard) => void
  runPending: boolean
}) {
  const accent = COLUMN_ACCENTS[accentIndex % COLUMN_ACCENTS.length]

  return (
    <div
      className={[
        "flex flex-col rounded-2xl border border-white/[0.05] bg-white/[0.01] backdrop-blur-sm transition-colors hover:bg-white/[0.02]",
        isMobile ? "h-full" : "h-full w-[296px] shrink-0",
      ].join(" ")}
    >
      {/* Column Header */}
      <div className="flex shrink-0 items-center justify-between px-4 py-3">
        <div className="flex items-center gap-2">
          <span
            className={`h-1.5 w-1.5 rounded-full ${accent.dot} ${accent.glow}`}
          />
          <span className="text-[10px] font-bold tracking-[0.18em] text-white/60 uppercase">
            {column.name}
          </span>
        </div>
        <span className="flex h-5 items-center justify-center rounded-full border border-white/[0.06] bg-white/[0.05] px-2 font-mono text-[10px] text-white/40">
          {column.cards.length}
        </span>
      </div>

      {/* Cards */}
      <div className="no-scrollbar flex-1 space-y-2 overflow-y-auto px-2 pb-2">
        {column.cards.map((card) => (
          <BoardCardView
            key={[card.id, card.column_id, card.title, card.description].join(
              ":",
            )}
            card={card}
            columns={columns}
            onDelete={() => onDelete(card.id)}
            onMove={(columnID) => onMove(card.id, columnID)}
            onRun={() => onRun(card.id)}
            onSave={onSave}
            runPending={runPending}
          />
        ))}

        {/* Draft card */}
        <div
          className={[
            "mt-3 flex flex-col gap-1 rounded-xl border bg-black/20 p-2 transition-colors focus-within:bg-black/40",
            draft ? "border-white/10" : "border-white/[0.03]",
          ].join(" ")}
        >
          <Input
            className="h-8 border-none bg-transparent px-2 text-sm font-medium text-white placeholder:text-white/25 focus-visible:ring-0 focus-visible:ring-offset-0"
            onChange={(e) => onDraftChange(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter" && !e.shiftKey && draft) {
                e.preventDefault()
                onAddCard()
              }
            }}
            placeholder="New task..."
            value={draft}
          />
          {draft && (
            <Textarea
              className="no-scrollbar min-h-[52px] resize-none border-none bg-transparent px-2 py-1 text-xs leading-relaxed text-white/55 placeholder:text-white/20 focus-visible:ring-0 focus-visible:ring-offset-0"
              onChange={(e) => onDescChange(e.target.value)}
              placeholder="Details (optional)..."
              value={desc}
            />
          )}
          <Button
            size="sm"
            className="mt-0.5 h-7 w-full rounded-lg bg-white/[0.04] text-[11px] font-semibold tracking-wide text-white/45 transition-all hover:bg-white/[0.09] hover:text-white/80 disabled:opacity-20"
            disabled={!draft}
            onClick={onAddCard}
          >
            + Add Card
          </Button>
        </div>
      </div>
    </div>
  )
}

// ── BoardSelector ────────────────────────────────────────────────────────────

function BoardSelector({
  board,
  boards,
  boardDescription,
  boardName,
  boardTemplate,
  onBoardDescriptionChange,
  onBoardNameChange,
  onBoardTemplateChange,
  onCreateBoard,
  onSelectedBoardChange,
  selectedBoardID,
}: {
  board?: Board
  boards: Board[]
  boardDescription: string
  boardName: string
  boardTemplate: keyof typeof BOARD_TEMPLATES
  onBoardDescriptionChange: (value: string) => void
  onBoardNameChange: (value: string) => void
  onBoardTemplateChange: (value: string) => void
  onCreateBoard: () => void
  onSelectedBoardChange: (value: string) => void
  selectedBoardID?: string
}) {
  return (
    <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-center sm:gap-3">
      {/* Board picker */}
      <div className="flex items-center gap-2 rounded-full border border-white/[0.06] bg-white/[0.02] p-1 backdrop-blur-md">
        <Select onValueChange={onSelectedBoardChange} value={selectedBoardID}>
          <SelectTrigger className="h-8 min-w-[160px] border-none bg-transparent text-sm font-semibold text-white/90 shadow-none focus:ring-0 sm:min-w-[200px]">
            <SelectValue placeholder="Select a board" />
          </SelectTrigger>
          <SelectContent className="border-white/10 bg-[#0c0c0e]/95 text-white backdrop-blur-xl">
            {boards.map((item) => (
              <SelectItem
                key={item.id}
                value={item.id}
                className="cursor-pointer focus:bg-white/10 focus:text-white"
              >
                {item.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {board?.description && (
        <span className="hidden max-w-[180px] truncate text-[11px] font-medium text-white/35 xl:block">
          {board.description}
        </span>
      )}

      {/* Create board inline form */}
      <div className="flex flex-wrap items-center gap-1 rounded-full border border-white/[0.04] bg-black/20 p-1">
        <Input
          className="h-8 w-28 border-none bg-transparent px-3 text-xs text-white placeholder:text-white/30 focus-visible:ring-1 focus-visible:ring-indigo-500/50 sm:w-36"
          onChange={(e) => onBoardNameChange(e.target.value)}
          placeholder="New board..."
          value={boardName}
        />
        <div className="h-4 w-px bg-white/10" />
        <Input
          className="h-8 w-28 border-none bg-transparent px-3 text-xs text-white placeholder:text-white/30 focus-visible:ring-1 focus-visible:ring-indigo-500/50 sm:w-40"
          onChange={(e) => onBoardDescriptionChange(e.target.value)}
          placeholder="Purpose..."
          value={boardDescription}
        />
        <div className="h-4 w-px bg-white/10" />
        <Select onValueChange={onBoardTemplateChange} value={boardTemplate}>
          <SelectTrigger className="h-8 w-28 border-none bg-transparent text-xs text-white/55 shadow-none hover:text-white/90 focus:ring-0">
            <SelectValue />
          </SelectTrigger>
          <SelectContent className="border-white/10 bg-[#0c0c0e]/95 backdrop-blur-xl">
            <SelectItem value="research">Research</SelectItem>
            <SelectItem value="default">Default</SelectItem>
          </SelectContent>
        </Select>
        <Button
          size="sm"
          disabled={!boardName.trim()}
          onClick={onCreateBoard}
          className="h-7 rounded-full bg-indigo-500/10 px-4 text-[10px] font-bold tracking-wider text-indigo-400 uppercase hover:bg-indigo-500/20 disabled:opacity-30"
        >
          Create
        </Button>
      </div>
    </div>
  )
}

// ── BoardCardView ────────────────────────────────────────────────────────────

function BoardCardView({
  card,
  columns,
  onDelete,
  onMove,
  onRun,
  onSave,
  runPending,
}: {
  card: BoardCard
  columns: Board["columns"]
  onDelete: () => void
  onMove: (columnID: string) => void
  onRun: () => void
  onSave: (card: BoardCard) => void
  runPending: boolean
}) {
  const [columnID, setColumnID] = useState(card.column_id)
  const [title, setTitle] = useState(card.title)
  const [description, setDescription] = useState(card.description)

  const hasChanges = title !== card.title || description !== card.description
  const needsMove = columnID !== card.column_id

  return (
    <div
      className={[
        "group relative flex flex-col gap-1.5 rounded-[1.125rem] border border-white/[0.05] bg-[#0c0c0e]/60 p-3 shadow-md backdrop-blur-xl",
        "transition-all duration-300",
        "hover:border-white/10 hover:bg-[#0c0c0e]/80 hover:shadow-lg",
        "focus-within:border-indigo-500/25 focus-within:bg-[#0c0c0e]/90 focus-within:shadow-[0_4px_20px_rgba(99,102,241,0.06)]",
      ].join(" ")}
    >
      <Input
        className="h-7 rounded-md border-none bg-transparent px-1.5 text-sm font-semibold text-white/90 placeholder:text-white/25 focus-visible:ring-1 focus-visible:ring-indigo-500/40"
        onChange={(e) => setTitle(e.target.value)}
        value={title}
        placeholder="Card title..."
      />
      <Textarea
        className="no-scrollbar min-h-[46px] resize-none rounded-md border-none bg-transparent px-1.5 py-0.5 text-xs leading-relaxed text-white/55 placeholder:text-white/20 focus-visible:ring-1 focus-visible:ring-indigo-500/40"
        onChange={(e) => setDescription(e.target.value)}
        value={description}
        placeholder="Details..."
      />

      {/* Card footer */}
      <div
        className={[
          "mt-1.5 flex flex-wrap items-center justify-between gap-2 border-t border-white/[0.04] pt-2 transition-opacity duration-300",
          hasChanges || needsMove
            ? "opacity-100"
            : "opacity-35 group-hover:opacity-100",
        ].join(" ")}
      >
        {/* Column mover */}
        <div className="flex items-center gap-1">
          <Select onValueChange={setColumnID} value={columnID}>
            <SelectTrigger className="h-6 max-w-[96px] border-none bg-transparent px-1 text-[10px] font-medium text-white/40 shadow-none hover:text-white/80 focus:ring-0">
              <span className="truncate">
                <SelectValue placeholder="Move" />
              </span>
            </SelectTrigger>
            <SelectContent className="border-white/10 bg-[#0c0c0e]/95 text-white backdrop-blur-xl">
              {columns.map((col) => (
                <SelectItem
                  key={col.id}
                  value={col.id}
                  className="text-xs focus:bg-white/10"
                >
                  {col.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          {needsMove && (
            <Button
              size="sm"
              onClick={() => onMove(columnID)}
              className="h-5 rounded-full bg-indigo-500/20 px-2 text-[9px] tracking-widest text-indigo-300 uppercase hover:bg-indigo-500/35"
            >
              Move →
            </Button>
          )}
        </div>

        {/* Action buttons */}
        <div className="flex items-center gap-1">
          {hasChanges && (
            <Button
              size="sm"
              onClick={() => onSave({ ...card, title, description })}
              className="h-6 rounded-full bg-emerald-500/15 px-3 text-[10px] font-bold text-emerald-400 hover:bg-emerald-500/25"
            >
              Save
            </Button>
          )}

          <Button
            disabled={runPending}
            onClick={onRun}
            size="icon"
            className="h-6 w-6 rounded-full bg-white/[0.04] text-white/40 transition-colors hover:bg-indigo-500/20 hover:text-indigo-400 disabled:opacity-30"
            title="Run agent"
          >
            <span className="text-[10px]">▶</span>
          </Button>

          <Button
            onClick={onDelete}
            size="icon"
            className="h-6 w-6 rounded-full bg-transparent text-white/20 transition-colors hover:bg-red-500/20 hover:text-red-400"
            title="Delete card"
          >
            <span className="text-[10px]">✕</span>
          </Button>
        </div>
      </div>
    </div>
  )
}

// ── ReviewPanel ──────────────────────────────────────────────────────────────

function ReviewPanel({
  enabled,
  every,
  busy,
  onToggle,
  onSaveInterval,
}: {
  enabled: boolean
  every: number
  busy: boolean
  onToggle: (enabled: boolean) => void
  onSaveInterval: (every: number) => void
}) {
  const [value, setValue] = useState(String(every))

  return (
    <div className="flex flex-wrap items-center gap-2 rounded-full border border-white/[0.06] bg-white/[0.02] p-1 backdrop-blur-md">
      {/* Status dot + label */}
      <div className="flex items-center gap-2 pr-1 pl-2">
        <div className="relative flex h-2 w-2 items-center justify-center">
          {enabled && (
            <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-50" />
          )}
          <span
            className={`relative inline-flex h-1.5 w-1.5 rounded-full transition-all duration-500 ${
              enabled
                ? "bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.8)]"
                : "bg-white/20"
            }`}
          />
        </div>
        <span className="text-[10px] font-bold tracking-[0.18em] text-white/45 uppercase">
          Heartbeat
        </span>
      </div>

      <Button
        size="sm"
        disabled={busy}
        onClick={() => onToggle(!enabled)}
        className={`h-7 rounded-full px-3 text-[10px] font-bold tracking-wider uppercase transition-all ${
          enabled
            ? "bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500/20"
            : "bg-white/[0.04] text-white/35 hover:bg-white/10 hover:text-white/80"
        }`}
      >
        {enabled ? "Active" : "Paused"}
      </Button>

      <div className="mx-0.5 h-4 w-px bg-white/10" />

      {/* Interval */}
      <div className="flex items-center gap-1 rounded-full bg-black/20 p-1">
        <Input
          className="h-6 w-10 border-none bg-transparent px-1 text-center text-xs text-white focus-visible:ring-1 focus-visible:ring-indigo-500/50"
          onChange={(e) => setValue(e.target.value)}
          placeholder="30"
          value={value}
        />
        <span className="mr-0.5 text-[10px] text-white/30">min</span>
        <Button
          size="sm"
          disabled={busy || Number.parseInt(value, 10) === every}
          onClick={() => onSaveInterval(Number.parseInt(value, 10) || 30)}
          className="h-6 rounded-full bg-indigo-500/10 px-3 text-[10px] font-bold text-indigo-400 transition-all hover:bg-indigo-500/20 disabled:opacity-0"
        >
          Save
        </Button>
      </div>
    </div>
  )
}

// ── Helpers ──────────────────────────────────────────────────────────────────

function requireBoardID(boardID?: string) {
  if (!boardID) throw new Error("board is not selected")
  return boardID
}

function showError(error: unknown) {
  toast.error(error instanceof Error ? error.message : "Request failed")
}
