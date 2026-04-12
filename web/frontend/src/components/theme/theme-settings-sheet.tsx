import { IconPalette, IconSparkles } from "@tabler/icons-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet"
import { type ThemeMode, type ThemePalette, presetPalettes } from "@/lib/theme"

interface ThemeSettingsSheetProps {
  theme: ThemeMode
  palettes: ThemePalette[]
  activePalette: ThemePalette
  onThemeChange: (mode: ThemeMode) => void
  onPaletteSelect: (paletteId: string) => void
  onPaletteSave: (palette: ThemePalette) => void
}

const colorFields: Array<keyof Omit<ThemePalette, "id" | "name">> = [
  "flame",
  "ember",
  "cream",
  "smoke",
  "midnight",
]

function buildCustomId(name: string) {
  const slug = name
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, "-")
  return `custom-${slug || "palette"}`
}

export function ThemeSettingsSheet({
  theme,
  palettes,
  activePalette,
  onThemeChange,
  onPaletteSelect,
  onPaletteSave,
}: ThemeSettingsSheetProps) {
  const { t } = useTranslation()
  const [editor, setEditor] = useState(activePalette)

  useEffect(() => {
    setEditor(activePalette)
  }, [activePalette])

  const saveCurrent = () => {
    onPaletteSave({
      ...editor,
      id: buildCustomId(editor.name),
    })
  }

  const isBuiltIn = presetPalettes.some(
    (palette) => palette.id === activePalette.id,
  )

  return (
    <Sheet>
      <SheetTrigger asChild>
        <Button variant="ghost" size="icon" className="size-8">
          <IconPalette className="size-4.5" />
        </Button>
      </SheetTrigger>
      <SheetContent className="border-l-border/70 w-full overflow-y-auto sm:max-w-xl">
        <SheetHeader className="pb-0">
          <SheetTitle>{t("theme.title")}</SheetTitle>
          <SheetDescription>{t("theme.description")}</SheetDescription>
        </SheetHeader>

        <div className="space-y-6 p-4 pt-0">
          <Card className="border-border/70 bg-card/80">
            <CardHeader>
              <CardTitle>{t("theme.mode")}</CardTitle>
            </CardHeader>
            <CardContent className="flex gap-2">
              {(["dark", "light"] as ThemeMode[]).map((mode) => (
                <Button
                  key={mode}
                  variant={theme === mode ? "default" : "outline"}
                  className="flex-1"
                  onClick={() => onThemeChange(mode)}
                >
                  {t(`theme.modes.${mode}`)}
                </Button>
              ))}
            </CardContent>
          </Card>

          <Card className="border-border/70 bg-card/80">
            <CardHeader>
              <CardTitle>{t("theme.presets")}</CardTitle>
            </CardHeader>
            <CardContent className="grid gap-3 md:grid-cols-2">
              {palettes.map((palette) => (
                <button
                  key={palette.id}
                  type="button"
                  onClick={() => onPaletteSelect(palette.id)}
                  className={`rounded-2xl border p-3 text-left transition ${
                    activePalette.id === palette.id
                      ? "border-primary bg-primary/10"
                      : "border-border/70 hover:border-primary/40 hover:bg-accent/40"
                  }`}
                >
                  <div className="mb-3 flex items-center justify-between">
                    <span className="font-medium">{palette.name}</span>
                    {activePalette.id === palette.id ? (
                      <IconSparkles className="text-primary size-4" />
                    ) : null}
                  </div>
                  <div className="flex gap-2">
                    {colorFields.map((field) => (
                      <span
                        key={field}
                        className="h-8 flex-1 rounded-full border border-black/10"
                        style={{ backgroundColor: palette[field] }}
                      />
                    ))}
                  </div>
                </button>
              ))}
            </CardContent>
          </Card>

          <Card className="border-border/70 bg-card/80">
            <CardHeader>
              <CardTitle>{t("theme.editor")}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <Input
                value={editor.name}
                onChange={(event) =>
                  setEditor((prev) => ({ ...prev, name: event.target.value }))
                }
                placeholder={t("theme.namePlaceholder")}
              />
              <div className="grid gap-3 md:grid-cols-2">
                {colorFields.map((field) => (
                  <label key={field} className="space-y-2">
                    <span className="text-muted-foreground text-xs font-medium tracking-[0.18em] uppercase">
                      {t(`theme.fields.${field}`)}
                    </span>
                    <div className="flex gap-2">
                      <Input
                        type="color"
                        value={editor[field]}
                        onChange={(event) =>
                          setEditor((prev) => ({
                            ...prev,
                            [field]: event.target.value,
                          }))
                        }
                        className="h-11 w-14 rounded-xl p-1"
                      />
                      <Input
                        value={editor[field]}
                        onChange={(event) =>
                          setEditor((prev) => ({
                            ...prev,
                            [field]: event.target.value,
                          }))
                        }
                      />
                    </div>
                  </label>
                ))}
              </div>
              <div className="flex gap-2">
                <Button
                  className="flex-1"
                  onClick={() =>
                    onPaletteSave({
                      ...editor,
                      id: isBuiltIn
                        ? buildCustomId(editor.name)
                        : activePalette.id,
                    })
                  }
                >
                  {t("theme.apply")}
                </Button>
                <Button
                  variant="outline"
                  className="flex-1"
                  onClick={saveCurrent}
                >
                  {t("theme.saveNew")}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </SheetContent>
    </Sheet>
  )
}
