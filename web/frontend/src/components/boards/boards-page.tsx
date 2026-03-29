import { useMemo, useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

import {
  createCard,
  deleteCard,
  getBoard,
  getBoards,
  type BoardCard,
  updateBoardReview,
  updateCard,
} from "@/api/boards"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"

export function BoardsPage() {
  const queryClient = useQueryClient()
  const [drafts, setDrafts] = useState<Record<string, string>>({})
  const [descs, setDescs] = useState<Record<string, string>>({})
  const boardsQuery = useQuery({ queryKey: ["boards"], queryFn: getBoards })
  const boardID = boardsQuery.data?.[0]?.id ?? "default"
  const boardQuery = useQuery({
    queryKey: ["boards", boardID],
    queryFn: () => getBoard(boardID),
  })

  const invalidate = async () => {
    await queryClient.invalidateQueries({ queryKey: ["boards"] })
    await queryClient.invalidateQueries({ queryKey: ["boards", boardID] })
  }

  const addMutation = useMutation({
    mutationFn: ({
      title,
      description,
      columnID,
    }: {
      title: string
      description: string
      columnID: string
    }) => createCard(boardID, { title, description, column_id: columnID }),
    onSuccess: async (_, vars) => {
      setDrafts((v) => ({ ...v, [vars.columnID]: "" }))
      setDescs((v) => ({ ...v, [vars.columnID]: "" }))
      await invalidate()
    },
    onError: showError,
  })

  const moveMutation = useMutation({
    mutationFn: ({
      card,
      columnID,
    }: {
      card: BoardCard
      columnID: string
    }) => updateCard(boardID, card.id, { column_id: columnID }),
    onSuccess: invalidate,
    onError: showError,
  })

  const editMutation = useMutation({
    mutationFn: (card: BoardCard) =>
      updateCard(boardID, card.id, {
        title: card.title,
        description: card.description,
      }),
    onSuccess: invalidate,
    onError: showError,
  })

  const deleteMutation = useMutation({
    mutationFn: (cardID: string) => deleteCard(boardID, cardID),
    onSuccess: invalidate,
    onError: showError,
  })

  const reviewMutation = useMutation({
    mutationFn: (enabled: boolean) =>
      updateBoardReview(boardID, {
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
      updateBoardReview(boardID, { enabled: true, every_minutes: every }),
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
          <ReviewPanel
            enabled={boardQuery.data?.review?.enabled ?? false}
            every={boardQuery.data?.review?.every_minutes ?? 30}
            busy={reviewMutation.isPending || intervalMutation.isPending}
            onToggle={(enabled) => reviewMutation.mutate(enabled)}
            onSaveInterval={(every) => intervalMutation.mutate(every)}
          />
          <div className="grid gap-4 xl:grid-cols-3">
            {ordered.map((column, index) => (
              <Card key={column.id} className="border-white/10 bg-white/5">
                <CardHeader>
                  <CardTitle className="flex items-center justify-between text-sm uppercase tracking-[0.24em]">
                    <span>{column.name}</span>
                    <span>{column.cards.length}</span>
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  {column.cards.map((card) => (
                    <BoardCardView
                      key={card.id}
                      card={card}
                      canLeft={index > 0}
                      canRight={index < ordered.length - 1}
                      onLeft={() =>
                        moveMutation.mutate({
                          card,
                          columnID: ordered[index - 1].id,
                        })
                      }
                      onRight={() =>
                        moveMutation.mutate({
                          card,
                          columnID: ordered[index + 1].id,
                        })
                      }
                      onSave={editMutation.mutate}
                      onDelete={() => deleteMutation.mutate(card.id)}
                    />
                  ))}
                  <div className="space-y-2 rounded-2xl border border-dashed border-white/10 p-3">
                    <Input
                      value={drafts[column.id] ?? ""}
                      onChange={(e) =>
                        setDrafts((v) => ({ ...v, [column.id]: e.target.value }))
                      }
                      placeholder="Add a task"
                    />
                    <Textarea
                      value={descs[column.id] ?? ""}
                      onChange={(e) =>
                        setDescs((v) => ({ ...v, [column.id]: e.target.value }))
                      }
                      placeholder="Details"
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
  )
}

function BoardCardView({
  card,
  canLeft,
  canRight,
  onLeft,
  onRight,
  onSave,
  onDelete,
}: {
  card: BoardCard
  canLeft: boolean
  canRight: boolean
  onLeft: () => void
  onRight: () => void
  onSave: (card: BoardCard) => void
  onDelete: () => void
}) {
  const [title, setTitle] = useState(card.title)
  const [description, setDescription] = useState(card.description)
  return (
    <div className="space-y-2 rounded-2xl border border-white/10 bg-black/10 p-3">
      <Input value={title} onChange={(e) => setTitle(e.target.value)} />
      <Textarea
        value={description}
        onChange={(e) => setDescription(e.target.value)}
      />
      <div className="flex flex-wrap gap-2">
        <Button size="sm" variant="outline" disabled={!canLeft} onClick={onLeft}>
          Left
        </Button>
        <Button size="sm" variant="outline" disabled={!canRight} onClick={onRight}>
          Right
        </Button>
        <Button
          size="sm"
          variant="secondary"
          onClick={() => onSave({ ...card, title, description })}
        >
          Save
        </Button>
        <Button size="sm" variant="destructive" onClick={onDelete}>
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
        <CardTitle className="text-sm uppercase tracking-[0.24em]">
          Board Review Heartbeat
        </CardTitle>
      </CardHeader>
      <CardContent className="flex flex-wrap items-center gap-3">
        <Button disabled={busy} onClick={() => onToggle(!enabled)}>
          {enabled ? "Disable review" : "Enable review"}
        </Button>
        <Input
          className="w-32"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          placeholder="30"
        />
        <Button
          variant="outline"
          disabled={busy}
          onClick={() => onSaveInterval(Number.parseInt(value, 10) || 30)}
        >
          Save minutes
        </Button>
      </CardContent>
    </Card>
  )
}

function showError(error: unknown) {
  toast.error(error instanceof Error ? error.message : "Request failed")
}
