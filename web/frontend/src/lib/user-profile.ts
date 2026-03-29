const USER_NAME_STORAGE_KEY = "jane-ai:user-name"

export function getUserDisplayName() {
  if (typeof window === "undefined") return "You"
  return window.localStorage.getItem(USER_NAME_STORAGE_KEY)?.trim() || "You"
}
