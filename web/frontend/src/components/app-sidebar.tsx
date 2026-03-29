import { IconChevronRight } from "@tabler/icons-react"
import {
  IconAtom,
  IconChevronsDown,
  IconChevronsUp,
  IconKey,
  IconLayoutKanban,
  IconListDetails,
  IconMasksTheater,
  IconMessageCircle,
  IconSettings,
  IconSparkles,
  IconTools,
} from "@tabler/icons-react"
import { Link, useRouterState } from "@tanstack/react-router"
import * as React from "react"
import { useTranslation } from "react-i18next"

import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible"
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from "@/components/ui/sidebar"
import { useSidebarChannels } from "@/hooks/use-sidebar-channels"
import { useSidebarPersonas } from "@/hooks/use-sidebar-personas"

interface NavItem {
  title: string
  to: string
  hash?: string
  icon: React.ComponentType<{ className?: string }>
  translateTitle?: boolean
}

interface NavGroup {
  label: string
  defaultOpen: boolean
  items: NavItem[]
  isChannelsGroup?: boolean
}

const baseNavGroups: Omit<NavGroup, "items">[] = [
  {
    label: "navigation.chat",
    defaultOpen: true,
  },
  {
    label: "navigation.model_group",
    defaultOpen: true,
  },
  {
    label: "navigation.agent_group",
    defaultOpen: true,
  },
  {
    label: "navigation.services",
    defaultOpen: true,
  },
]

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const routerState = useRouterState()
  const { t } = useTranslation()
  const currentPath = routerState.location.pathname
  const {
    channelItems,
    hasMoreChannels,
    showAllChannels,
    toggleShowAllChannels,
  } = useSidebarChannels({ t })
  const personaItems = useSidebarPersonas()

  const navGroups: NavGroup[] = React.useMemo(() => {
    return [
      {
        ...baseNavGroups[0],
        items: [
          {
            title: "navigation.chat",
            to: "/",
            icon: IconMessageCircle,
            translateTitle: true,
          },
        ],
      },
      {
        ...baseNavGroups[1],
        items: [
          {
            title: "navigation.models",
            to: "/models",
            icon: IconAtom,
            translateTitle: true,
          },
          {
            title: "navigation.credentials",
            to: "/credentials",
            icon: IconKey,
            translateTitle: true,
          },
        ],
      },
      {
        label: "navigation.channels_group",
        defaultOpen: true,
        items: channelItems.map((item) => ({
          title: item.title,
          to: item.url,
          icon: item.icon,
          translateTitle: false,
        })),
        isChannelsGroup: true,
      },
      {
        ...baseNavGroups[2],
        items: [
          {
            title: "navigation.skills",
            to: "/agent/skills",
            icon: IconSparkles,
            translateTitle: true,
          },
          {
            title: "navigation.tools",
            to: "/agent/tools",
            icon: IconTools,
            translateTitle: true,
          },
          {
            title: "navigation.personas",
            to: "/config",
            hash: "personas-section",
            icon: IconMasksTheater,
            translateTitle: true,
          },
          ...personaItems,
        ],
      },
      {
        ...baseNavGroups[3],
        items: [
          {
            title: "navigation.boards",
            to: "/boards",
            icon: IconLayoutKanban,
            translateTitle: true,
          },
          {
            title: "navigation.config",
            to: "/config",
            icon: IconSettings,
            translateTitle: true,
          },
          {
            title: "navigation.logs",
            to: "/logs",
            icon: IconListDetails,
            translateTitle: true,
          },
        ],
      },
    ]
  }, [channelItems, personaItems])

  return (
    <Sidebar
      {...props}
      className="border-r-sidebar-border/60 bg-sidebar/95 border-r pt-3"
    >
      <SidebarContent className="bg-background">
        <div className="text-sidebar-foreground mx-3 mb-4 rounded-3xl border border-white/10 bg-white/6 p-3 shadow-lg shadow-black/20">
          <div className="mb-3 flex items-center gap-3">
            <img className="size-11 rounded-2xl" src="/jane-mark.svg" alt="" />
            <div>
              <p className="font-serif text-base font-semibold tracking-[0.24em] uppercase">
                JANE-ai
              </p>
              <p className="text-sidebar-foreground/60 text-xs tracking-[0.28em] uppercase">
                Endgame mesh
              </p>
            </div>
          </div>
          <p className="text-sidebar-foreground/72 text-xs leading-5">
            Tactical reasoning, model routing, and channel control from one
            quiet interface.
          </p>
        </div>
        {navGroups.map((group) => (
          <Collapsible
            key={group.label}
            defaultOpen={group.defaultOpen}
            className="group/collapsible mb-1"
          >
            <SidebarGroup className="px-2 py-0">
              <SidebarGroupLabel asChild>
                <CollapsibleTrigger className="hover:bg-muted/60 flex w-full cursor-pointer items-center justify-between rounded-md px-2 py-1.5 transition-colors">
                  <span>{t(group.label)}</span>
                  <IconChevronRight className="size-3.5 opacity-50 transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                </CollapsibleTrigger>
              </SidebarGroupLabel>
              <CollapsibleContent>
                <SidebarGroupContent className="pt-1">
                  <SidebarMenu>
                    {group.items.map((item) => {
                      const isActive =
                        currentPath === item.to ||
                        (item.to !== "/" &&
                          currentPath.startsWith(`${item.to}/`))
                      return (
                        <SidebarMenuItem key={item.title}>
                          <SidebarMenuButton
                            asChild
                            isActive={isActive}
                            className={`h-9 px-3 ${isActive ? "bg-accent/80 text-foreground font-medium" : "text-muted-foreground hover:bg-muted/60"}`}
                          >
                            <Link to={item.to} hash={item.hash}>
                              <item.icon
                                className={`size-4 ${isActive ? "opacity-100" : "opacity-60"}`}
                              />
                              <span
                                className={
                                  isActive ? "opacity-100" : "opacity-80"
                                }
                              >
                                {item.translateTitle === false
                                  ? item.title
                                  : t(item.title)}
                              </span>
                            </Link>
                          </SidebarMenuButton>
                        </SidebarMenuItem>
                      )
                    })}
                    {group.isChannelsGroup && hasMoreChannels && (
                      <SidebarMenuItem key="channels-more-toggle">
                        <SidebarMenuButton
                          onClick={toggleShowAllChannels}
                          className="text-muted-foreground hover:bg-muted/60 h-9 px-3"
                        >
                          {showAllChannels ? (
                            <IconChevronsUp className="size-4 opacity-60" />
                          ) : (
                            <IconChevronsDown className="size-4 opacity-60" />
                          )}
                          <span className="opacity-80">
                            {showAllChannels
                              ? t("navigation.show_less_channels")
                              : t("navigation.show_more_channels")}
                          </span>
                        </SidebarMenuButton>
                      </SidebarMenuItem>
                    )}
                  </SidebarMenu>
                </SidebarGroupContent>
              </CollapsibleContent>
            </SidebarGroup>
          </Collapsible>
        ))}
      </SidebarContent>
      <SidebarRail />
    </Sidebar>
  )
}
