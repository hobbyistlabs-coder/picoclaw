import { useCallback, useEffect, useState } from "react"

import {
  type ThemeMode,
  type ThemePalette,
  type ThemeSettings,
  applyThemeSettings,
  getActivePalette,
  getAllPalettes,
  getStoredThemeSettings,
  saveThemeSettings,
  upsertCustomPalette,
} from "@/lib/theme"

export function useTheme() {
  const [settings, setSettings] = useState<ThemeSettings>(
    getStoredThemeSettings,
  )

  useEffect(() => {
    applyThemeSettings(settings)
    saveThemeSettings(settings)
  }, [settings])

  const setTheme = useCallback((mode: ThemeMode) => {
    setSettings((prev) => ({ ...prev, mode }))
  }, [])

  const toggleTheme = useCallback(() => {
    setSettings((prev) => ({
      ...prev,
      mode: prev.mode === "dark" ? "light" : "dark",
    }))
  }, [])

  const selectPalette = useCallback((paletteId: string) => {
    setSettings((prev) => ({ ...prev, activePaletteId: paletteId }))
  }, [])

  const savePalette = useCallback((palette: ThemePalette) => {
    setSettings((prev) => upsertCustomPalette(prev, palette))
  }, [])

  return {
    theme: settings.mode,
    settings,
    palettes: getAllPalettes(settings),
    activePalette: getActivePalette(settings),
    setTheme,
    toggleTheme,
    selectPalette,
    savePalette,
  }
}
