import { IconCode, IconDeviceFloppy } from "@tabler/icons-react"
import { useQuery } from "@tanstack/react-query"
import { Link } from "@tanstack/react-router"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { getAutoStartStatus, getLauncherConfig } from "@/api/system"
import {
  AgentDefaultsSection,
  DevicesSection,
  LauncherSection,
  RuntimeSection,
} from "@/components/config/config-sections"
import {
  type CoreConfigForm,
  EMPTY_FORM,
  EMPTY_LAUNCHER_FORM,
  type LauncherForm,
  buildFormFromConfig,
  createEmptyPersona,
} from "@/components/config/form-model"
import { PersonasSection } from "@/components/config/personas-section"
import { WorkspacePromptsSection } from "@/components/config/workspace-prompts-section"
import { PageHeader } from "@/components/page-header"

// (Keep your existing imports for useTranslation, useQuery, useQueryClient, toast, icons, etc.)

export function ConfigPage() {
  const { t } = useTranslation()
  const [form, setForm] = useState<CoreConfigForm>(EMPTY_FORM)
  const [baseline, setBaseline] = useState<CoreConfigForm>(EMPTY_FORM)
  const [launcherForm, setLauncherForm] =
    useState<LauncherForm>(EMPTY_LAUNCHER_FORM)
  const [launcherBaseline, setLauncherBaseline] =
    useState<LauncherForm>(EMPTY_LAUNCHER_FORM)
  const [autoStartEnabled, setAutoStartEnabled] = useState(false)
  const [autoStartBaseline, setAutoStartBaseline] = useState(false)
  const [saving, setSaving] = useState(false)

  const { data, isLoading, error } = useQuery({
    queryKey: ["config"],
    queryFn: async () => {
      const res = await fetch("/api/config")
      if (!res.ok) {
        throw new Error("Failed to load config")
      }
      return res.json()
    },
  })

  const { data: launcherConfig, isLoading: isLauncherLoading } = useQuery({
    queryKey: ["system", "launcher-config"],
    queryFn: getLauncherConfig,
  })

  const {
    data: autoStartStatus,
    isLoading: isAutoStartLoading,
    error: autoStartError,
  } = useQuery({
    queryKey: ["system", "autostart"],
    queryFn: getAutoStartStatus,
  })

  useEffect(() => {
    if (!data) return
    const parsed = buildFormFromConfig(data)
    setForm(parsed)
    setBaseline(parsed)
  }, [data])

  useEffect(() => {
    if (!data) return
    const hash = window.location.hash.replace(/^#/, "")
    if (!hash) return
    window.setTimeout(() => {
      document.getElementById(hash)?.scrollIntoView({
        block: "start",
        behavior: "smooth",
      })
    }, 50)
  }, [data])

  useEffect(() => {
    if (!launcherConfig) return
    const parsed: LauncherForm = {
      port: String(launcherConfig.port),
      publicAccess: launcherConfig.public,
      allowedCIDRsText: (launcherConfig.allowed_cidrs ?? []).join("\n"),
    }
    setLauncherForm(parsed)
    setLauncherBaseline(parsed)
  }, [launcherConfig])

  useEffect(() => {
    if (!autoStartStatus) return
    setAutoStartEnabled(autoStartStatus.enabled)
    setAutoStartBaseline(autoStartStatus.enabled)
  }, [autoStartStatus])

  const configDirty = JSON.stringify(form) !== JSON.stringify(baseline)
  const launcherDirty =
    JSON.stringify(launcherForm) !== JSON.stringify(launcherBaseline)
  const autoStartDirty = autoStartEnabled !== autoStartBaseline
  const isDirty = configDirty || launcherDirty || autoStartDirty

  const autoStartSupported = autoStartStatus?.supported !== false
  const autoStartHint = autoStartError
    ? t("pages.config.autostart_load_error")
    : !autoStartSupported
      ? t("pages.config.autostart_unsupported")
      : t("pages.config.autostart_hint")

  const updateField = <K extends keyof CoreConfigForm>(
    key: K,
    value: CoreConfigForm[K],
  ) => {
    setForm((prev) => ({ ...prev, [key]: value }))
  }

  const updateLauncherField = <K extends keyof LauncherForm>(
    key: K,
    value: LauncherForm[K],
  ) => {
    setLauncherForm((prev) => ({ ...prev, [key]: value }))
  }

  const updatePersonaField = <
    K extends keyof CoreConfigForm["personas"][number],
  >(
    personaKey: string,
    key: K,
    value: CoreConfigForm["personas"][number][K],
  ) => {
    setForm((prev) => ({
      ...prev,
      personas: prev.personas.map((persona) =>
        persona.key === personaKey ? { ...persona, [key]: value } : persona,
      ),
    }))
  }

  const addPersona = () => {
    setForm((prev) => ({
      ...prev,
      personas: [...prev.personas, createEmptyPersona()],
    }))
  }

  const removePersona = (personaKey: string) => {
    setForm((prev) => ({
      ...prev,
      personas: prev.personas.filter((persona) => persona.key !== personaKey),
    }))
  }

  const handleReset = () => {
    setForm(baseline)
    setLauncherForm(launcherBaseline)
    setAutoStartEnabled(autoStartBaseline)
    toast.info(t("pages.config.reset_success"))
  }

  const handleSave = async () => {
    try {
      setSaving(true)
      // ... (Keep all your existing validation and patchAppConfig logic here exactly as is)
      // I am leaving this intact to save space and focus on the UI changes

      toast.success(t("pages.config.save_success"))
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : t("pages.config.save_error"),
      )
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="flex h-full flex-col text-white">
      <PageHeader
        title={t("navigation.config")}
        children={
          <Link
            to="/config/raw"
            className="group flex items-center gap-1.5 rounded-full border border-white/[0.05] bg-white/[0.02] px-4 py-1.5 text-[11px] font-bold tracking-wider text-white/60 uppercase shadow-inner backdrop-blur-xl transition-all hover:scale-[1.02] hover:bg-white/[0.05] hover:text-white"
          >
            <IconCode className="size-4 transition-transform group-hover:rotate-90" />
            {t("pages.config.open_raw")}
          </Link>
        }
      />
      <div className="flex-1 overflow-auto p-3 lg:p-6">
        <div className="mx-auto w-full max-w-[1000px] space-y-6">
          {isLoading ? (
            <div className="py-6 text-sm text-white/40">
              {t("labels.loading")}
            </div>
          ) : error ? (
            <div className="py-6 text-sm text-red-400">
              {t("pages.config.load_error")}
            </div>
          ) : (
            <div className="space-y-6">
              {isDirty && (
                <div className="flex items-center rounded-xl border border-amber-500/20 bg-amber-500/10 px-4 py-3 text-sm text-amber-200 shadow-inner backdrop-blur-md">
                  {t("pages.config.unsaved_changes")}
                </div>
              )}

              <AgentDefaultsSection form={form} onFieldChange={updateField} />

              <WorkspacePromptsSection />

              <PersonasSection
                sectionId="personas-section"
                personas={form.personas}
                disabled={saving}
                onAdd={addPersona}
                onRemove={removePersona}
                onChange={updatePersonaField}
              />

              <RuntimeSection form={form} onFieldChange={updateField} />

              <LauncherSection
                launcherForm={launcherForm}
                onFieldChange={updateLauncherField}
                disabled={saving || isLauncherLoading}
              />

              <DevicesSection
                form={form}
                onFieldChange={updateField}
                autoStartEnabled={autoStartEnabled}
                autoStartHint={autoStartHint}
                autoStartDisabled={
                  isAutoStartLoading ||
                  Boolean(autoStartError) ||
                  !autoStartSupported ||
                  saving
                }
                onAutoStartChange={setAutoStartEnabled}
              />

              {/* Glassmorphic Action Buttons */}
              <div className="mt-8 flex justify-end gap-3 rounded-full border border-white/[0.05] bg-white/[0.02] p-1.5 shadow-inner backdrop-blur-xl">
                <button
                  onClick={handleReset}
                  disabled={!isDirty || saving}
                  className="flex items-center gap-1.5 rounded-full px-5 py-2 text-[11px] font-bold tracking-wider text-white/40 uppercase transition-all hover:bg-white/10 hover:text-white disabled:opacity-20 disabled:hover:bg-transparent disabled:hover:text-white/40"
                >
                  {t("common.reset")}
                </button>
                <button
                  onClick={handleSave}
                  disabled={!isDirty || saving}
                  className="group flex items-center gap-1.5 rounded-full border border-indigo-500/20 bg-indigo-500/10 px-6 py-2 text-[11px] font-bold tracking-wider text-indigo-100 uppercase shadow-inner backdrop-blur-md transition-all hover:scale-[1.02] disabled:pointer-events-none disabled:opacity-20"
                >
                  <IconDeviceFloppy className="size-4 text-indigo-400 transition-transform group-hover:rotate-90" />
                  {saving ? t("common.saving") : t("common.save")}
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
