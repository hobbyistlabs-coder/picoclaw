export interface OpenRouterModel {
  id: string
  canonical_slug: string
  name: string
  created: number
  description: string
  context_length: number
  architecture: {
    input_modalities: string[]
    output_modalities: string[]
    tokenizer: string
    instruct_type: string | null
  }
  pricing: Record<string, string>
  top_provider: {
    context_length: number
    max_completion_tokens: number
    is_moderated: boolean
  }
  supported_parameters: string[]
  expiration_date: string | null
}

interface OpenRouterModelsResponse {
  data: OpenRouterModel[]
  total: number
}

const BASE_URL = ""

export async function getOpenRouterModels(params?: {
  outputModalities?: string
  supportedParameters?: string
}): Promise<OpenRouterModelsResponse> {
  const query = new URLSearchParams()
  if (params?.outputModalities) {
    query.set("output_modalities", params.outputModalities)
  }
  if (params?.supportedParameters) {
    query.set("supported_parameters", params.supportedParameters)
  }

  const suffix = query.toString()
  const res = await fetch(
    `${BASE_URL}/api/models/openrouter/catalog${suffix ? `?${suffix}` : ""}`,
  )
  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`)
  }
  return res.json() as Promise<OpenRouterModelsResponse>
}
