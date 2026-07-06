"use client";

const LAST_VISITED_PATH_STORAGE_KEY = "indosupplier:last-visited-path";

function normalizeInternalPath(path: string | null | undefined): string | null {
  if (!path) {
    return null;
  }

  const normalized = path.trim();
  if (!normalized.startsWith("/") || normalized.startsWith("//")) {
    return null;
  }

  return normalized;
}

export function isAuthRoute(path: string | null | undefined): boolean {
  const normalized = normalizeInternalPath(path);
  if (!normalized) {
    return false;
  }

  return (
    normalized === "/login" ||
    normalized.startsWith("/login?") ||
    normalized === "/sysadmin/login" ||
    normalized.startsWith("/sysadmin/login?")
  );
}

function shouldPersistPath(path: string): boolean {
  return (
    !isAuthRoute(path) &&
    path !== "/register" &&
    !path.startsWith("/register?") &&
    path !== "/supplier/register" &&
    !path.startsWith("/supplier/register?")
  );
}

export function persistLastVisitedPath(path: string): void {
  if (typeof window === "undefined") {
    return;
  }

  const normalized = normalizeInternalPath(path);
  if (!normalized || !shouldPersistPath(normalized)) {
    return;
  }

  window.sessionStorage.setItem(LAST_VISITED_PATH_STORAGE_KEY, normalized);
}

export function getStoredLastVisitedPath(): string | null {
  if (typeof window === "undefined") {
    return null;
  }

  const stored = window.sessionStorage.getItem(LAST_VISITED_PATH_STORAGE_KEY);
  const normalized = normalizeInternalPath(stored);

  if (!normalized || !shouldPersistPath(normalized)) {
    return null;
  }

  return normalized;
}

export function resolvePostLoginRedirectTarget(
  explicitRedirectTo?: string,
  fallbackPath = "/",
): string {
  const explicit = normalizeInternalPath(explicitRedirectTo);
  if (explicit && !isAuthRoute(explicit)) {
    return explicit;
  }

  const stored = getStoredLastVisitedPath();
  if (stored) {
    return stored;
  }

  return fallbackPath;
}
