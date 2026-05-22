/**
 * Cookie utility functions for secure cookie management
 */

/**
 * Sets a cookie with appropriate security flags for production
 * @param name Cookie name
 * @param value Cookie value
 * @param maxAge Max age in seconds (default: 7 days)
 * @param sameSite SameSite attribute (default: 'Lax')
 */
export function setSecureCookie(
  name: string,
  value: string,
  maxAge: number = 7 * 24 * 60 * 60, // 7 days default
  sameSite: "Strict" | "Lax" | "None" = "Lax",
): void {
  if (typeof globalThis.window === "undefined") {
    return;
  }

  // Check if we're in HTTPS (production)
  const isSecure = globalThis.window.location.protocol === "https:";
  const isProduction = process.env.NODE_ENV === "production";

  // Build cookie string
  let cookieString = `${name}=${value}; path=/; max-age=${maxAge}; SameSite=${sameSite}`;

  // Add Secure flag only in production with HTTPS
  // Note: Browser will ignore Secure flag if not in HTTPS, but we add it for production
  if (isSecure || isProduction) {
    cookieString += "; Secure";
  }

  globalThis.document.cookie = cookieString;
}

/**
 * Deletes a cookie
 * @param name Cookie name
 */
export function deleteCookie(name: string): void {
  if (typeof globalThis.window === "undefined") {
    return;
  }

  globalThis.document.cookie = `${name}=; path=/; max-age=0`;
}

/**
 * Gets a cookie value by name
 * @param name Cookie name
 * @returns Cookie value or null
 */
export function getCookie(name: string): string | null {
  if (typeof globalThis.window === "undefined") {
    return null;
  }

  const value = `; ${globalThis.document.cookie}`;
  const parts = value.split(`; ${name}=`);

  if (parts.length === 2) {
    return parts.pop()?.split(";").shift() || null;
  }

  return null;
}
