import { IconLoader2, IconPlus, IconStar } from "@tabler/icons-react"
import { useQuery } from "@tanstack/react-query"
import { useCallback, useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

import { type ModelInfo, getModels, setDefaultModel } from "@/api/models"
import { type OpenRouterModel, getOpenRouterModels } from "@/api/openrouter"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import {
  buildOpenRouterAlias,
  mergeOpenRouterCatalog,
  parseOpenRouterPrice,
} from "@/lib/openrouter"

import { AddModelSheet } from "./add-model-sheet"
import { DeleteModelDialog } from "./delete-model-dialog"
import { EditModelSheet } from "./edit-model-sheet"
import { OpenRouterDiscovery } from "./openrouter-discovery"
import { getProviderKey, getProviderLabel } from "./provider-label"
import { ProviderSection } from "./provider-section"

const PROVIDER_PRIORITY: Record<string, number> = {
  volcengine: 0,
  openai: 1,
  gemini: 2,
  anthropic: 3,
  zhipu: 4,
  deepseek: 5,
  openrouter: 6,
  qwen: 7,
  moonshot: 8,
  groq: 9,
  "github-copilot": 10,
  antigravity: 11,
  nvidia: 12,
  cerebras: 13,
  shengsuanyun: 14,
  ollama: 15,
  vllm: 16,
  mistral: 17,
  avian: 18,
}

interface ProviderGroup {
  key: string
  label: string
  models: ModelInfo[]
  hasDefault: boolean
  configuredCount: number
}

export function ModelsPage() {
  const { t } = useTranslation()
  const [models, setModels] = useState<ModelInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [fetchError, setFetchError] = useState("")

  const [editingModel, setEditingModel] = useState<ModelInfo | null>(null)
  const [deletingModel, setDeletingModel] = useState<ModelInfo | null>(null)
  const [catalogPrefill, setCatalogPrefill] =
    useState<Partial<ModelInfo> | null>(null)
  const [addOpen, setAddOpen] = useState(false)
  const [settingDefaultIndex, setSettingDefaultIndex] = useState<number | null>(
    null,
  )
  const openRouterCatalogQuery = useQuery({
    queryKey: ["openrouter-models", "all", ""],
    queryFn: () => getOpenRouterModels({ outputModalities: "all" }),
    staleTime: 300_000,
  })

  const fetchModels = useCallback(async () => {
    try {
      const data = await getModels()
      const sorted = [...data.models].sort((a, b) => {
        if (a.is_default && !b.is_default) return -1
        if (!a.is_default && b.is_default) return 1
        if (a.configured && !b.configured) return -1
        if (!a.configured && b.configured) return 1
        return a.model_name.localeCompare(b.model_name)
      })
      setModels(sorted)
      setFetchError("")
    } catch (e) {
      setFetchError(e instanceof Error ? e.message : t("models.loadError"))
    } finally {
      setLoading(false)
    }
  }, [t])

  const enrichedModels = useMemo(
    () =>
      mergeOpenRouterCatalog(models, openRouterCatalogQuery.data?.data ?? []),
    [models, openRouterCatalogQuery.data?.data],
  )

  useEffect(() => {
    fetchModels()
  }, [fetchModels])

  const handleSetDefault = async (model: ModelInfo) => {
    setSettingDefaultIndex(model.index)
    try {
      await setDefaultModel(model.model_name)
      await fetchModels()
    } catch {
      // ignore
    } finally {
      setSettingDefaultIndex(null)
    }
  }

  const grouped: Record<string, { label: string; models: ModelInfo[] }> = {}
  for (const model of enrichedModels) {
    const providerKey = getProviderKey(model.model)
    if (!grouped[providerKey]) {
      grouped[providerKey] = {
        label: getProviderLabel(model.model),
        models: [],
      }
    }
    grouped[providerKey].models.push(model)
  }

  const providerGroups: ProviderGroup[] = Object.entries(grouped)
    .map(([key, group]) => {
      const configuredCount = group.models.filter(
        (model) => model.configured,
      ).length
      return {
        key,
        label: group.label,
        models: group.models,
        hasDefault: group.models.some((model) => model.is_default),
        configuredCount,
      }
    })
    .sort((a, b) => {
      if (a.hasDefault && !b.hasDefault) return -1
      if (!a.hasDefault && b.hasDefault) return 1

      if (a.configuredCount !== b.configuredCount) {
        return b.configuredCount - a.configuredCount
      }

      const aPriority = PROVIDER_PRIORITY[a.key] ?? Number.MAX_SAFE_INTEGER
      const bPriority = PROVIDER_PRIORITY[b.key] ?? Number.MAX_SAFE_INTEGER
      if (aPriority !== bPriority) {
        return aPriority - bPriority
      }

      return a.label.localeCompare(b.label)
    })

  const defaultModel = enrichedModels.find((model) => model.is_default)

  const handleQuickAdd = (model: OpenRouterModel) => {
    setCatalogPrefill({
      model_name: buildOpenRouterAlias(
        model,
        models.map((entry) => entry.model_name),
      ),
      model: `openrouter/${model.id}`,
      api_base: "https://openrouter.ai/api/v1",
      input_price_per_m_token: parseOpenRouterPrice(model.pricing.prompt),
      output_price_per_m_token: parseOpenRouterPrice(model.pricing.completion),
      catalog: model,
    })
    setAddOpen(true)
  }

  return (
    <div className="flex h-full flex-col">
      <PageHeader title={t("navigation.models")}>
        <div className="flex items-center gap-3">
          <Button size="sm" variant="outline" onClick={() => setAddOpen(true)}>
            <IconPlus className="size-4" />
            {t("models.add.button")}
          </Button>
        </div>
      </PageHeader>

      <div className="min-h-0 flex-1 overflow-y-auto px-4 sm:px-6">
        <div className="pt-2">
          {!defaultModel && (
            <div className="text-muted-foreground flex items-center gap-1.5 text-sm">
              <span>{t("models.noDefaultHintPrefix")}</span>
              <IconStar className="size-3.5 shrink-0" />
              <span>{t("models.noDefaultHintSuffix")}</span>
            </div>
          )}
          <p className="text-muted-foreground mt-1 text-sm">
            {t("models.description")}
          </p>
        </div>

        <OpenRouterDiscovery onQuickAdd={handleQuickAdd} />

        {loading && (
          <div className="flex items-center justify-center py-20">
            <IconLoader2 className="text-muted-foreground size-6 animate-spin" />
          </div>
        )}

        {fetchError && (
          <div className="text-destructive bg-destructive/10 rounded-lg px-4 py-3 text-sm">
            {fetchError}
          </div>
        )}

        {!loading && !fetchError && (
          <div className="pb-8">
            {providerGroups.map((providerGroup) => (
              <ProviderSection
                key={providerGroup.key}
                provider={providerGroup.label}
                providerKey={providerGroup.key}
                models={providerGroup.models}
                onEdit={setEditingModel}
                onSetDefault={handleSetDefault}
                onDelete={setDeletingModel}
                settingDefaultIndex={settingDefaultIndex}
              />
            ))}
          </div>
        )}
      </div>

      <EditModelSheet
        model={editingModel}
        open={editingModel !== null}
        onClose={() => setEditingModel(null)}
        onSaved={fetchModels}
      />

      <AddModelSheet
        open={addOpen}
        onClose={() => {
          setAddOpen(false)
          setCatalogPrefill(null)
        }}
        onSaved={fetchModels}
        existingModelNames={models.map((model) => model.model_name)}
        initialModel={catalogPrefill}
      />

      <DeleteModelDialog
        model={deletingModel}
        onClose={() => setDeletingModel(null)}
        onDeleted={fetchModels}
      />
    </div>
  )
}
