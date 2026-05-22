/**
 * Clear all authentication-related cookies from the browser.
 *
 * This is used when:
 * 1. Session verification fails (stale cookies)
 * 2. Manual logout
 * 3. API returns 401 and refresh fails
 *
 * Note: HttpOnly cookies (gims_access_token, gims_refresh_token) cannot be
 * cleared by JavaScript. They must be cleared by the server via Set-Cookie
 * with MaxAge=-1. This function only clears non-HttpOnly cookies like csrf_token.
 *
 * For full cookie cleanup, always call the /auth/logout API endpoint.
 */
export function clearAuthCookies(): void {
  if (typeof document === "undefined") return;

  const cookiesToClear = [
    "gims_csrf_token",
    // Note: These are HttpOnly, so this won't actually work for them
    // but we include them for completeness in case they're ever non-HttpOnly
    "gims_access_token",
    "gims_refresh_token",
  ];

  cookiesToClear.forEach((cookieName) => {
    // Clear with various path combinations to ensure removal
    const paths = ["/", "", "/api"];
    paths.forEach((path) => {
      document.cookie = `${cookieName}=; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=${path}`;
      document.cookie = `${cookieName}=; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=${path}; SameSite=Strict`;
      document.cookie = `${cookieName}=; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=${path}; SameSite=Lax`;
    });
  });
}

/**
 * Attempt to logout via API to clear HttpOnly cookies server-side.
 * This is fire-and-forget - we don't care if it fails.
 */
export async function clearAuthCookiesViaApi(): Promise<void> {
  try {
    const { authService } = await import("../services/auth-service");
    await authService.logout();
  } catch {
    // Ignore errors - we still want to clear local state
    // The server might be down or the token already invalid
  }
}

/**
 * Full cleanup: clear local cookies AND call API to clear HttpOnly cookies.
 */
export async function fullAuthCleanup(): Promise<void> {
  // Clear what we can locally first
  clearAuthCookies();
  // Then try to clear HttpOnly cookies via API
  await clearAuthCookiesViaApi();
}
