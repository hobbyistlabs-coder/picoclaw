import { IconLoader2, IconSparkles } from "@tabler/icons-react"
import { useQuery } from "@tanstack/react-query"
import { useDeferredValue, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

import { type OpenRouterModel, getOpenRouterModels } from "@/api/openrouter"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { parseOpenRouterPrice } from "@/lib/openrouter"

import { ModelCapabilityBadges } from "./model-capability-badges"

const OUTPUTS = ["all", "text", "image", "audio"] as const
const FEATURES = ["", "tools", "structured_outputs", "reasoning"] as const

function formatPrice(model: OpenRouterModel) {
  const input = parseOpenRouterPrice(model.pricing.prompt)
  const output = parseOpenRouterPrice(model.pricing.completion)
  if (input && output) return `In $${input}/M · Out $${output}/M`
  return input ? `$${input}/M` : "Variable"
}

export function OpenRouterDiscovery({
  onQuickAdd,
}: {
  onQuickAdd: (model: OpenRouterModel) => void
}) {
  const { t } = useTranslation()
  const [search, setSearch] = useState("")
  const [output, setOutput] = useState<(typeof OUTPUTS)[number]>("all")
  const [feature, setFeature] = useState<(typeof FEATURES)[number]>("")
  const deferredSearch = useDeferredValue(search)
  const { data, isLoading, error } = useQuery({
    queryKey: ["openrouter-models", output, feature],
    queryFn: () =>
      getOpenRouterModels({
        outputModalities: output,
        supportedParameters: feature || undefined,
      }),
    staleTime: 300_000,
  })

  const models = useMemo(() => {
    const query = deferredSearch.trim().toLowerCase()
    return (data?.data ?? [])
      .filter((model) =>
        [model.name, model.id, model.description].some((value) =>
          value.toLowerCase().includes(query),
        ),
      )
      .slice(0, 12)
  }, [data?.data, deferredSearch])

  return (
    <section className="from-card via-card to-muted/40 my-6 rounded-2xl border bg-linear-to-br p-5">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <IconSparkles className="size-4" />
            <h2 className="text-base font-semibold">
              {t("models.discovery.title")}
            </h2>
          </div>
          <p className="text-muted-foreground max-w-2xl text-sm">
            {t("models.discovery.description")}
          </p>
        </div>
        <Input
          value={search}
          onChange={(event) => setSearch(event.target.value)}
          placeholder={t("models.discovery.searchPlaceholder")}
          className="w-full lg:max-w-sm"
        />
      </div>

      <div className="mt-4 flex flex-wrap gap-2">
        {OUTPUTS.map((value) => (
          <Button
            key={value}
            variant={output === value ? "default" : "outline"}
            size="sm"
            onClick={() => setOutput(value)}
          >
            {t(`models.discovery.outputs.${value}`)}
          </Button>
        ))}
        {FEATURES.map((value) => (
          <Button
            key={value || "all"}
            variant={feature === value ? "default" : "outline"}
            size="sm"
            onClick={() => setFeature(value)}
          >
            {t(`models.discovery.features.${value || "all"}`)}
          </Button>
        ))}
      </div>

      {isLoading ? (
        <div className="flex items-center gap-2 py-8 text-sm">
          <IconLoader2 className="size-4 animate-spin" />
          {t("models.discovery.loading")}
        </div>
      ) : null}
      {error instanceof Error ? (
        <p className="text-destructive mt-4 text-sm">{error.message}</p>
      ) : null}

      <div className="mt-5 grid grid-cols-1 gap-3 xl:grid-cols-2">
        {models.map((model) => (
          <article
            key={model.id}
            className="bg-background/80 rounded-xl border p-4 shadow-sm"
          >
            <div className="flex items-start justify-between gap-3">
              <div className="space-y-1">
                <h3 className="text-sm font-semibold">{model.name}</h3>
                <p className="text-muted-foreground font-mono text-[11px]">
                  {model.id}
                </p>
              </div>
              <Button size="sm" onClick={() => onQuickAdd(model)}>
                {t("models.discovery.quickAdd")}
              </Button>
            </div>
            <p className="text-muted-foreground mt-3 line-clamp-3 text-sm leading-6">
              {model.description}
            </p>
            <div className="mt-3 flex flex-wrap items-center gap-2 text-xs">
              <span className="bg-muted rounded-full px-2 py-1">
                {t("models.discovery.context", {
                  count: model.context_length.toLocaleString(),
                })}
              </span>
              <span className="bg-muted rounded-full px-2 py-1">
                {formatPrice(model)}
              </span>
              {model.top_provider.is_moderated ? (
                <span className="bg-muted rounded-full px-2 py-1">
                  {t("models.discovery.moderated")}
                </span>
              ) : null}
            </div>
            <div className="mt-3">
              <ModelCapabilityBadges model={model} limit={5} />
            </div>
          </article>
        ))}
      </div>
    </section>
  )
}
