export type ThemeMode = "light" | "dark"

export interface ThemePalette {
  id: string
  name: string
  flame: string
  ember: string
  cream: string
  smoke: string
  midnight: string
}

export interface ThemeSettings {
  mode: ThemeMode
  activePaletteId: string
  customPalettes: ThemePalette[]
}

const WHITE = "#fffaf2"
const BLACK = "#140f10"
const STORAGE_KEY = "theme-settings-v1"

export const presetPalettes: ThemePalette[] = [
  {
    id: "cyndaquil",
    name: "Cyndaquil",
    flame: "#ff8f2f",
    ember: "#ffb14a",
    cream: "#f7ebd2",
    smoke: "#6b7288",
    midnight: "#171d2b",
  },
  {
    id: "quilava",
    name: "Quilava",
    flame: "#ff6a3d",
    ember: "#ff9e57",
    cream: "#ffe7bf",
    smoke: "#7b6a79",
    midnight: "#221729",
  },
  {
    id: "volcanic-ash",
    name: "Volcanic Ash",
    flame: "#ff7b36",
    ember: "#ffcb72",
    cream: "#f4ead8",
    smoke: "#7a818c",
    midnight: "#121826",
  },
]

export function getStoredThemeSettings(): ThemeSettings {
  if (typeof window === "undefined") {
    return {
      mode: "dark",
      activePaletteId: presetPalettes[0].id,
      customPalettes: [],
    }
  }

  const legacyTheme = window.localStorage.getItem("theme")

  try {
    const parsed = JSON.parse(window.localStorage.getItem(STORAGE_KEY) ?? "")
    return sanitizeSettings(parsed, legacyTheme)
  } catch {
    return sanitizeSettings({}, legacyTheme)
  }
}

export function sanitizePalette(palette: Partial<ThemePalette>): ThemePalette {
  return {
    id: String(palette.id || createPaletteId(palette.name || "custom-palette")),
    name: String(palette.name || "Custom Palette"),
    flame: normalizeHex(palette.flame, "#ff8f2f"),
    ember: normalizeHex(palette.ember, "#ffb14a"),
    cream: normalizeHex(palette.cream, "#f7ebd2"),
    smoke: normalizeHex(palette.smoke, "#6b7288"),
    midnight: normalizeHex(palette.midnight, "#171d2b"),
  }
}

export function getAllPalettes(settings: ThemeSettings): ThemePalette[] {
  return [...presetPalettes, ...settings.customPalettes]
}

export function getActivePalette(settings: ThemeSettings): ThemePalette {
  return (
    getAllPalettes(settings).find(
      (palette) => palette.id === settings.activePaletteId,
    ) ?? presetPalettes[0]
  )
}

export function saveThemeSettings(settings: ThemeSettings) {
  if (typeof window === "undefined") return
  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(settings))
  window.localStorage.setItem("theme", settings.mode)
}

export function applyThemeSettings(settings: ThemeSettings) {
  if (typeof document === "undefined") return

  const root = document.documentElement
  const palette = getActivePalette(settings)
  const tokens = deriveThemeTokens(settings.mode, palette)

  root.classList.toggle("dark", settings.mode === "dark")

  Object.entries(tokens).forEach(([key, value]) => {
    root.style.setProperty(key, value)
  })
}

export function upsertCustomPalette(
  settings: ThemeSettings,
  palette: ThemePalette,
): ThemeSettings {
  const nextPalette = sanitizePalette(palette)
  const customPalettes = settings.customPalettes.some(
    (item) => item.id === nextPalette.id,
  )
    ? settings.customPalettes.map((item) =>
        item.id === nextPalette.id ? nextPalette : item,
      )
    : [...settings.customPalettes, nextPalette]

  return {
    ...settings,
    activePaletteId: nextPalette.id,
    customPalettes,
  }
}

function sanitizeSettings(
  raw: unknown,
  legacyTheme: string | null,
): ThemeSettings {
  const record = asRecord(raw)
  const mode =
    record.mode === "light" || record.mode === "dark"
      ? record.mode
      : legacyTheme === "light"
        ? "light"
        : "dark"
  const customPalettes = Array.isArray(record.customPalettes)
    ? record.customPalettes.map((item) => sanitizePalette(asRecord(item)))
    : []
  const allPalettes = [...presetPalettes, ...customPalettes]
  const activePaletteId = String(record.activePaletteId || presetPalettes[0].id)

  return {
    mode,
    activePaletteId: allPalettes.some((item) => item.id === activePaletteId)
      ? activePaletteId
      : presetPalettes[0].id,
    customPalettes,
  }
}

function deriveThemeTokens(mode: ThemeMode, palette: ThemePalette) {
  const background =
    mode === "dark"
      ? mixHex(palette.midnight, BLACK, 0.22)
      : mixHex(palette.cream, WHITE, 0.38)
  const foreground =
    mode === "dark"
      ? mixHex(palette.cream, WHITE, 0.12)
      : mixHex(palette.midnight, BLACK, 0.16)
  const card =
    mode === "dark"
      ? mixHex(palette.midnight, palette.flame, 0.05)
      : mixHex(background, WHITE, 0.28)
  const popover = card
  const secondary =
    mode === "dark"
      ? mixHex(palette.midnight, palette.cream, 0.08)
      : mixHex(palette.cream, palette.ember, 0.12)
  const muted =
    mode === "dark"
      ? mixHex(palette.midnight, WHITE, 0.06)
      : mixHex(background, palette.cream, 0.48)
  const accent =
    mode === "dark" ? mixHex(palette.flame, palette.ember, 0.4) : palette.ember
  const border =
    mode === "dark"
      ? withAlphaHex(mixHex(palette.midnight, palette.cream, 0.22), "b8")
      : mixHex(background, palette.midnight, 0.14)
  const input =
    mode === "dark"
      ? withAlphaHex(mixHex(palette.midnight, palette.cream, 0.18), "cc")
      : mixHex(background, palette.midnight, 0.18)
  const sidebar =
    mode === "dark" ? palette.midnight : mixHex(palette.midnight, BLACK, 0.08)
  const sidebarAccent =
    mode === "dark"
      ? mixHex(palette.midnight, palette.flame, 0.12)
      : mixHex(palette.midnight, palette.ember, 0.16)
  const glowTop = withAlphaRgb(palette.flame, 0.18)
  const glowBottom = withAlphaRgb(palette.cream, 0.12)
  const bgTop = mode === "dark" ? mixHex(palette.midnight, BLACK, 0.3) : WHITE
  const bgBottom =
    mode === "dark"
      ? mixHex(palette.midnight, palette.smoke, 0.1)
      : mixHex(palette.cream, palette.ember, 0.2)

  return {
    "--background": background,
    "--foreground": foreground,
    "--card": card,
    "--card-foreground": foreground,
    "--popover": popover,
    "--popover-foreground": foreground,
    "--primary": palette.flame,
    "--primary-foreground": getContrastColor(palette.flame),
    "--secondary": secondary,
    "--secondary-foreground": foreground,
    "--muted": muted,
    "--muted-foreground": mixHex(palette.smoke, foreground, 0.28),
    "--accent": accent,
    "--accent-foreground": getContrastColor(accent),
    "--destructive": "#dc5a43",
    "--border": border,
    "--input": input,
    "--ring": palette.flame,
    "--chart-1": palette.flame,
    "--chart-2": palette.ember,
    "--chart-3": palette.midnight,
    "--chart-4": palette.cream,
    "--chart-5": palette.smoke,
    "--sidebar": sidebar,
    "--sidebar-foreground": getContrastColor(sidebar),
    "--sidebar-primary": palette.flame,
    "--sidebar-primary-foreground": getContrastColor(palette.flame),
    "--sidebar-accent": sidebarAccent,
    "--sidebar-accent-foreground": getContrastColor(sidebarAccent),
    "--sidebar-border": mixHex(sidebar, palette.cream, 0.14),
    "--sidebar-ring": palette.flame,
    "--theme-flame-rgb": toRgbTriplet(palette.flame),
    "--theme-ember-rgb": toRgbTriplet(palette.ember),
    "--theme-cream-rgb": toRgbTriplet(palette.cream),
    "--theme-smoke-rgb": toRgbTriplet(palette.smoke),
    "--theme-midnight-rgb": toRgbTriplet(palette.midnight),
    "--theme-glow-top": glowTop,
    "--theme-glow-bottom": glowBottom,
    "--theme-bg-top": bgTop,
    "--theme-bg-bottom": bgBottom,
  }
}

function asRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : {}
}

function createPaletteId(name: string) {
  return (
    name
      .toLowerCase()
      .trim()
      .replace(/[^a-z0-9]+/g, "-") || "custom-palette"
  )
}

function normalizeHex(value: unknown, fallback: string) {
  if (typeof value !== "string") return fallback
  const hex = value.trim()
  return /^#([0-9a-f]{6})$/i.test(hex) ? hex.toLowerCase() : fallback
}

function withAlphaHex(hex: string, alphaHex: string) {
  return `${hex}${alphaHex}`
}

function withAlphaRgb(hex: string, alpha: number) {
  return `rgb(${toRgbTriplet(hex)} / ${alpha})`
}

function toRgbTriplet(hex: string) {
  const { r, g, b } = hexToRgb(hex)
  return `${r} ${g} ${b}`
}

function mixHex(from: string, to: string, amount: number) {
  const a = hexToRgb(from)
  const b = hexToRgb(to)
  const mix = (start: number, end: number) =>
    Math.round(start + (end - start) * amount)

  return rgbToHex({
    r: mix(a.r, b.r),
    g: mix(a.g, b.g),
    b: mix(a.b, b.b),
  })
}

function getContrastColor(hex: string) {
  const { r, g, b } = hexToRgb(hex)
  const luminance = (0.2126 * r + 0.7152 * g + 0.0722 * b) / 255
  return luminance > 0.62 ? BLACK : WHITE
}

function hexToRgb(hex: string) {
  const normalized = hex.replace("#", "")
  return {
    r: parseInt(normalized.slice(0, 2), 16),
    g: parseInt(normalized.slice(2, 4), 16),
    b: parseInt(normalized.slice(4, 6), 16),
  }
}

function rgbToHex({ r, g, b }: { r: number; g: number; b: number }) {
  return `#${[r, g, b]
    .map((value) => value.toString(16).padStart(2, "0"))
    .join("")}`
}
