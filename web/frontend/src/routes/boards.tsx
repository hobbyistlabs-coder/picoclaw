import { createFileRoute } from "@tanstack/react-router"

import { BoardsPage } from "@/components/boards/boards-page"

export const Route = createFileRoute("/boards")({
  component: BoardsPage,
})
