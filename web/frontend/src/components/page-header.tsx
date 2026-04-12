import { IconMenu2 } from "@tabler/icons-react"
import type { ReactNode } from "react"

import { SidebarTrigger } from "@/components/ui/sidebar"

interface PageHeaderProps {
  title: string
  description?: string
  titleExtra?: ReactNode
  children?: ReactNode
}

export function PageHeader({
  title,
  description,
  titleExtra,
  children,
}: PageHeaderProps) {
  return (
    <div className="flex shrink-0 flex-col gap-4 px-4 pt-4 pb-2 md:px-6">
      <div className="flex flex-col justify-between gap-4 xl:flex-row xl:items-center">
        <div className="flex min-w-0 items-start gap-3 md:gap-4">
          <SidebarTrigger className="border-border/60 bg-background text-muted-foreground hover:bg-accent hover:text-foreground hidden h-9 w-9 rounded-lg border sm:flex [&>svg]:size-5">
            <IconMenu2 />
          </SidebarTrigger>
          <div className="min-w-0">
            <h2 className="truncate font-serif text-xl font-semibold tracking-[0.08em] text-white/92">
              {title}
            </h2>
            {description ? (
              <p className="text-muted-foreground mt-1 text-xs leading-5 md:text-sm">
                {description}
              </p>
            ) : null}
          </div>
        </div>
        <div className="flex flex-col gap-3 xl:items-end">
          {titleExtra ? <div className="flex flex-wrap items-center gap-2">{titleExtra}</div> : null}
          {children ? <div className="flex flex-wrap items-center gap-2">{children}</div> : null}
        </div>
      </div>
    </div>
  )
}
