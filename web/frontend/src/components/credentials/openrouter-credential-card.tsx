import { IconKey, IconLoader2, IconRouter } from "@tabler/icons-react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

import { CredentialCard } from "./credential-card"

interface OpenRouterCredentialCardProps {
  configured: boolean
  activeAction: string
  token: string
  onTokenChange: (value: string) => void
  onSaveToken: () => void
}

export function OpenRouterCredentialCard({
  configured,
  activeAction,
  token,
  onTokenChange,
  onSaveToken,
}: OpenRouterCredentialCardProps) {
  const { t } = useTranslation()
  const loading = activeAction === "openrouter:token"

  return (
    <CredentialCard
      title={
        <span className="inline-flex items-center gap-2">
          <span className="border-muted inline-flex size-6 items-center justify-center rounded-full border">
            <IconRouter className="size-3.5" />
          </span>
          <span>OpenRouter</span>
        </span>
      }
      description={t("credentials.providers.openrouter.description")}
      status={configured ? "connected" : "not_logged_in"}
      authMethod="api_key"
      details={<p>{t("credentials.providers.openrouter.restartHint")}</p>}
      actions={
        <div className="border-muted flex h-[120px] items-center gap-2 rounded-lg border p-3">
          <Input
            value={token}
            onChange={(e) => onTokenChange(e.target.value)}
            type="password"
            placeholder={t("credentials.fields.openrouterToken")}
          />
          <Button
            size="sm"
            disabled={loading || !token.trim()}
            onClick={onSaveToken}
          >
            {loading && <IconLoader2 className="size-4 animate-spin" />}
            <IconKey className="size-4" />
            {t("credentials.actions.saveToken")}
          </Button>
        </div>
      }
    />
  )
}
