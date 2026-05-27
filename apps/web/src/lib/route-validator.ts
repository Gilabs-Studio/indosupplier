const VALID_ROUTES = ["/", "/login"] as const;

export function isValidRoute(href: string | null | undefined): boolean {
  if (!href) {
    return false;
  }

  const normalized = href.trim();
  if (normalized === "") {
    return false;
  }

  const path = normalized.startsWith("/") ? normalized : `/${normalized}`;
  return (VALID_ROUTES as readonly string[]).includes(path);
}

export function getValidRoutes(): readonly string[] {
  return VALID_ROUTES;
}
