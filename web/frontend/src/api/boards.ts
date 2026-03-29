export interface BoardCard {
  id: string
  board_id: string
  column_id: string
  title: string
  description: string
}

export interface BoardColumn {
  id: string
  key: string
  name: string
  cards: BoardCard[]
}

export interface BoardReview {
  enabled: boolean
  every_minutes: number
  channel: string
  chat_id: string
}

export interface Board {
  id: string
  name: string
  description: string
  columns: BoardColumn[]
  review?: BoardReview | null
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(path, options)
  if (!res.ok) {
    throw new Error(await res.text())
  }
  return res.json() as Promise<T>
}

export function getBoards() {
  return request<Board[]>("/api/boards")
}

export function getBoard(id: string) {
  return request<Board>(`/api/boards/${encodeURIComponent(id)}`)
}

export function createCard(
  boardID: string,
  payload: { title: string; description: string; column_id?: string },
) {
  return request<BoardCard>(`/api/boards/${encodeURIComponent(boardID)}/cards`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  })
}

export function updateCard(
  boardID: string,
  cardID: string,
  payload: { title?: string; description?: string; column_id?: string },
) {
  return request<BoardCard>(
    `/api/boards/${encodeURIComponent(boardID)}/cards/${encodeURIComponent(cardID)}`,
    {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    },
  )
}

export function deleteCard(boardID: string, cardID: string) {
  return request<{ status: string }>(
    `/api/boards/${encodeURIComponent(boardID)}/cards/${encodeURIComponent(cardID)}`,
    { method: "DELETE" },
  )
}

export function updateBoardReview(
  boardID: string,
  payload: { enabled: boolean; every_minutes: number; channel?: string; chat_id?: string },
) {
  return request<BoardReview>(
    `/api/boards/${encodeURIComponent(boardID)}/review`,
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    },
  )
}
