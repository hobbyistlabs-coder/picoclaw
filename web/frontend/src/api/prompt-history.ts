export interface PromptRevision {
  id: string
  kind: string
  name: string
  timestamp: string
  content: string
}

interface PromptHistoryResponse {
  revisions: PromptRevision[]
}

async function request<T>(path: string): Promise<T> {
  const res = await fetch(path)
  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`)
  }
  return res.json() as Promise<T>
}

export async function getWorkspaceFileHistory(name: string) {
  return request<PromptHistoryResponse>(`/api/workspace/bootstrap/${name}/history`)
}

export async function getPersonaPromptHistory(id: string) {
  return request<PromptHistoryResponse>(
    `/api/config/personas/${encodeURIComponent(id)}/system-prompt/history`,
  )
}
