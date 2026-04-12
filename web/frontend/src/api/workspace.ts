export interface WorkspaceBootstrapFile {
  name: string
  path: string
  content: string
  exists: boolean
}

export interface WorkspaceBootstrapResponse {
  files: WorkspaceBootstrapFile[]
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(path, options)
  if (!res.ok) {
    let message = `API error: ${res.status} ${res.statusText}`
    try {
      const body = (await res.text()).trim()
      if (body) {
        message = body
      }
    } catch {
      // Keep fallback message.
    }
    throw new Error(message)
  }
  return res.json() as Promise<T>
}

export async function getWorkspaceBootstrapFiles() {
  return request<WorkspaceBootstrapResponse>("/api/workspace/bootstrap")
}

export async function updateWorkspaceBootstrapFile(
  name: string,
  content: string,
) {
  return request<WorkspaceBootstrapFile>(`/api/workspace/bootstrap/${name}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ content }),
  })
}
