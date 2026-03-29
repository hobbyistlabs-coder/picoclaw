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
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
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
  const boardsQuery = useQuery({ queryKey: ["boards"], queryFn: getBoards })
  const boardID = useMemo(() => {
    if (!boardsQuery.data?.length) {
      return undefined
    }
    return boardsQuery.data.some((board) => board.id === selectedBoardID)
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
        columns: BOARD_TEMPLATES[boardTemplate],
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
    onSuccess: invalidate,
    onError: showError,
  })

  const editMutation = useMutation({
    mutationFn: (card: BoardCard) =>
      updateCard(requireBoardID(boardID), card.id, {
        title: card.title,
        description: card.description,
      }),
    onSuccess: invalidate,
    onError: showError,
  })

  const deleteMutation = useMutation({
    mutationFn: (cardID: string) => deleteCard(requireBoardID(boardID), cardID),
    onSuccess: invalidate,
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
    onSuccess: invalidate,
    onError: showError,
  })

  const intervalMutation = useMutation({
    mutationFn: (every: number) =>
      updateBoardReview(requireBoardID(boardID), {
        enabled: true,
        every_minutes: every,
      }),
    onSuccess: invalidate,
    onError: showError,
  })

  const ordered = useMemo(
    () => boardQuery.data?.columns ?? [],
    [boardQuery.data?.columns],
  )

  return (
    <div className="flex h-full flex-col">
      <PageHeader title="Boards" />
      <div className="flex-1 overflow-auto px-6 py-4">
        <div className="mx-auto flex max-w-7xl flex-col gap-6">
          <BoardSelector
            board={boardQuery.data}
            boards={boardsQuery.data ?? []}
            boardDescription={boardDescription}
            boardName={boardName}
            boardTemplate={boardTemplate}
            onBoardDescriptionChange={setBoardDescription}
            onBoardNameChange={setBoardName}
            onBoardTemplateChange={(value) =>
              setBoardTemplate(value as keyof typeof BOARD_TEMPLATES)
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
          <Card className="border-white/10 bg-white/5">
            <CardHeader>
              <CardTitle className="text-sm tracking-[0.24em] uppercase">
                Board Shape
              </CardTitle>
            </CardHeader>
            <CardContent className="flex flex-wrap items-center gap-3">
              <Input
                className="max-w-xs"
                onChange={(e) => setColumnName(e.target.value)}
                placeholder="Add a new column"
                value={columnName}
              />
              <Button
                disabled={!boardID || addColumnMutation.isPending}
                onClick={() => addColumnMutation.mutate()}
              >
                Add column
              </Button>
              <p className="text-muted-foreground text-sm">
                Boards are no longer limited to the default three lanes.
              </p>
            </CardContent>
          </Card>
          <div className="overflow-x-auto pb-2">
            <div className="grid auto-cols-[minmax(19rem,22rem)] grid-flow-col gap-4">
              {ordered.map((column) => (
                <Card key={column.id} className="border-white/10 bg-white/5">
                  <CardHeader>
                    <CardTitle className="flex items-center justify-between text-sm tracking-[0.24em] uppercase">
                      <span>{column.name}</span>
                      <span>{column.cards.length}</span>
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-3">
                    {column.cards.map((card) => (
                      <BoardCardView
                        key={[
                          card.id,
                          card.column_id,
                          card.title,
                          card.description,
                        ].join(":")}
                        card={card}
                        columns={ordered}
                        onDelete={() => deleteMutation.mutate(card.id)}
                        onMove={(columnID) =>
                          moveMutation.mutate({ cardID: card.id, columnID })
                        }
                        onRun={() => runMutation.mutate(card.id)}
                        onSave={editMutation.mutate}
                        runPending={runMutation.isPending}
                      />
                    ))}
                    <div className="space-y-2 rounded-2xl border border-dashed border-white/10 p-3">
                      <Input
                        onChange={(e) =>
                          setDrafts((v) => ({
                            ...v,
                            [column.id]: e.target.value,
                          }))
                        }
                        placeholder="Add a task"
                        value={drafts[column.id] ?? ""}
                      />
                      <Textarea
                        onChange={(e) =>
                          setDescs((v) => ({
                            ...v,
                            [column.id]: e.target.value,
                          }))
                        }
                        placeholder="Details"
                        value={descs[column.id] ?? ""}
                      />
                      <Button
                        className="w-full"
                        onClick={() =>
                          addMutation.mutate({
                            title: drafts[column.id] ?? "",
                            description: descs[column.id] ?? "",
                            columnID: column.id,
                          })
                        }
                      >
                        Add card
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

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
    <Card className="border-white/10 bg-white/5">
      <CardHeader>
        <CardTitle className="text-sm tracking-[0.24em] uppercase">
          Board Workspace
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex flex-wrap items-center gap-3">
          <Select onValueChange={onSelectedBoardChange} value={selectedBoardID}>
            <SelectTrigger className="min-w-64">
              <SelectValue placeholder="Select a board" />
            </SelectTrigger>
            <SelectContent>
              {boards.map((item) => (
                <SelectItem key={item.id} value={item.id}>
                  {item.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {board?.description ? (
            <p className="text-muted-foreground text-sm">{board.description}</p>
          ) : null}
        </div>
        <div className="grid gap-3 md:grid-cols-[1.3fr_1.4fr_0.8fr_auto]">
          <Input
            onChange={(e) => onBoardNameChange(e.target.value)}
            placeholder="New board name"
            value={boardName}
          />
          <Input
            onChange={(e) => onBoardDescriptionChange(e.target.value)}
            placeholder="What this board tracks"
            value={boardDescription}
          />
          <Select onValueChange={onBoardTemplateChange} value={boardTemplate}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="research">Research template</SelectItem>
              <SelectItem value="default">Default template</SelectItem>
            </SelectContent>
          </Select>
          <Button disabled={!boardName.trim()} onClick={onCreateBoard}>
            Create board
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

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

  return (
    <div className="space-y-2 rounded-2xl border border-white/10 bg-black/10 p-3">
      <Input onChange={(e) => setTitle(e.target.value)} value={title} />
      <Textarea
        onChange={(e) => setDescription(e.target.value)}
        value={description}
      />
      <div className="flex gap-2">
        <Select onValueChange={setColumnID} value={columnID}>
          <SelectTrigger className="flex-1">
            <SelectValue placeholder="Move to" />
          </SelectTrigger>
          <SelectContent>
            {columns.map((column) => (
              <SelectItem key={column.id} value={column.id}>
                {column.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Button
          disabled={columnID === card.column_id}
          onClick={() => onMove(columnID)}
          size="sm"
          variant="outline"
        >
          Move
        </Button>
      </div>
      <div className="flex flex-wrap gap-2">
        <Button
          onClick={() => onSave({ ...card, title, description })}
          size="sm"
          variant="secondary"
        >
          Save
        </Button>
        <Button
          disabled={runPending}
          onClick={onRun}
          size="sm"
          variant="outline"
        >
          Run agent
        </Button>
        <Button onClick={onDelete} size="sm" variant="destructive">
          Delete
        </Button>
      </div>
    </div>
  )
}

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
    <Card className="border-white/10 bg-white/5">
      <CardHeader>
        <CardTitle className="text-sm tracking-[0.24em] uppercase">
          Board Review Heartbeat
        </CardTitle>
      </CardHeader>
      <CardContent className="flex flex-wrap items-center gap-3">
        <Button disabled={busy} onClick={() => onToggle(!enabled)}>
          {enabled ? "Disable review" : "Enable review"}
        </Button>
        <Input
          className="w-32"
          onChange={(e) => setValue(e.target.value)}
          placeholder="30"
          value={value}
        />
        <Button
          disabled={busy}
          onClick={() => onSaveInterval(Number.parseInt(value, 10) || 30)}
          variant="outline"
        >
          Save minutes
        </Button>
      </CardContent>
    </Card>
  )
}

function requireBoardID(boardID?: string) {
  if (!boardID) {
    throw new Error("board is not selected")
  }
  return boardID
}

function showError(error: unknown) {
  toast.error(error instanceof Error ? error.message : "Request failed")
}
