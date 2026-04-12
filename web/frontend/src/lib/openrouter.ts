import type { ModelInfo } from "@/api/models"
import type { OpenRouterModel } from "@/api/openrouter"

export function isOpenRouterModel(model: Pick<ModelInfo, "model">) {
  return model.model.startsWith("openrouter/")
}

export function getOpenRouterCatalogId(model: Pick<ModelInfo, "model">) {
  return isOpenRouterModel(model) ? model.model.slice("openrouter/".length) : ""
}

export function mergeOpenRouterCatalog(
  models: ModelInfo[],
  catalog: OpenRouterModel[],
) {
  const byId = new Map(catalog.map((entry) => [entry.id, entry]))
  return models.map((model) => ({
    ...model,
    catalog: byId.get(getOpenRouterCatalogId(model)),
  }))
}

export function buildOpenRouterAlias(
  model: Pick<OpenRouterModel, "id">,
  existingNames: string[],
) {
  const base = `openrouter-${model.id}`
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
  let candidate = base
  let suffix = 2
  while (existingNames.includes(candidate)) {
    candidate = `${base}-${suffix}`
    suffix += 1
  }
  return candidate
}

export function parseOpenRouterPrice(rate?: string) {
  const value = Number(rate)
  return Number.isFinite(value) && value > 0 ? value * 1_000_000 : undefined
}

export function getCapabilityLabels(model?: OpenRouterModel) {
  if (!model) return []
  const labels = [
    ...model.architecture.input_modalities.filter((item) => item !== "text"),
    ...model.architecture.output_modalities.filter((item) => item !== "text"),
  ]
  if (model.supported_parameters.includes("tools")) labels.push("tools")
  if (
    model.supported_parameters.includes("structured_outputs") ||
    model.supported_parameters.includes("response_format")
  ) {
    labels.push("json")
  }
  if (model.supported_parameters.includes("reasoning")) labels.push("reasoning")
  return [...new Set(labels)]
}
