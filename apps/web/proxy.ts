import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import {
  getLocaleFromCountryCode,
  getLocalePreferenceFromCookieValue,
  LOCALE_PREFERENCE_COOKIE,
} from "@/lib/locale-preference";

/**
 * Proxy middleware for handling authentication routing.
 * 
 * IMPORTANT: Cookie presence does NOT guarantee valid session.
 * HttpOnly cookies cannot be validated by middleware - only the backend can verify them.
 * 
 * This middleware only handles:
 * 1. Root path redirect when cookies exist (for convenience)
 * 2. Login page access - always allow (let client-side verify session)
 * 
 * All protected route auth is handled by client-side AuthGuard which:
 * - Verifies session with backend on every page load
 * - Redirects to login if session is invalid
 * - Clears stale auth state properly
 */
export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;
  // Check for the actual HttpOnly cookie set by the backend
  const accessToken = request.cookies.get("indosupplier_access_token")?.value;
  const cookieLocale = getLocalePreferenceFromCookieValue(
    request.cookies.get(LOCALE_PREFERENCE_COOKIE)?.value,
  );

  // Extract locale from pathname
  const pathSegments = pathname.split("/").filter(Boolean);
  const pathLocale = pathSegments[0] === "en" || pathSegments[0] === "id" 
    ? pathSegments[0] 
    : null;
  const requestCountryCode =
    request.headers.get("x-vercel-ip-country") ??
    request.headers.get("cf-ipcountry") ??
    request.headers.get("x-country-code");
  const detectedLocale = pathLocale ?? cookieLocale ?? getLocaleFromCountryCode(requestCountryCode);

  const response = NextResponse.next();
  if (cookieLocale !== detectedLocale) {
    response.cookies.set({
      name: LOCALE_PREFERENCE_COOKIE,
      value: detectedLocale,
      path: "/",
      sameSite: "lax",
      maxAge: 60 * 60 * 24 * 365,
    });
  }

  // Normalize public auth routes to locale-prefixed paths.
  // This keeps links like /register/success?token=... working in Next.js proxy mode
  // without requiring middleware.ts locale rewriting.
  const hasLocalePrefix = pathSegments[0] === "en" || pathSegments[0] === "id";
  const isLegacySettingsPath = pathname === "/settings";
  const isLocaleAgnosticAuthPath =
    pathname === "/login" ||
    pathname === "/register" ||
    pathname === "/register/success" ||
    pathname === "/profile" ||
    isLegacySettingsPath;
  if (!hasLocalePrefix && isLocaleAgnosticAuthPath) {
    const normalizedPath = isLegacySettingsPath ? "/profile" : pathname;
    const targetURL = new URL(`/${detectedLocale}${normalizedPath}`, request.url);
    targetURL.search = request.nextUrl.search;
    if (isLegacySettingsPath && !targetURL.searchParams.has("tab")) {
      targetURL.searchParams.set("tab", "billing");
    }
    const redirectResponse = NextResponse.redirect(targetURL);
    redirectResponse.cookies.set({
      name: LOCALE_PREFERENCE_COOKIE,
      value: detectedLocale,
      path: "/",
      sameSite: "lax",
      maxAge: 60 * 60 * 24 * 365,
    });
    return redirectResponse;
  }

  // If accessing root and cookie exists, redirect to dashboard
  // NOTE: Client-side AuthGuard will verify the session and redirect to login if invalid
  if (pathname === "/" && accessToken) {
    const target = `/${detectedLocale}/dashboard`;
    const redirectResponse = NextResponse.redirect(new URL(target, request.url));
    redirectResponse.cookies.set({
      name: LOCALE_PREFERENCE_COOKIE,
      value: detectedLocale,
      path: "/",
      sameSite: "lax",
      maxAge: 60 * 60 * 24 * 365,
    });
    return redirectResponse;
  }

  // CRITICAL: Do NOT redirect from login page based on cookie presence alone
  // The cookie might be invalid (e.g., after server restart)
  // Let the login page's client-side logic handle the redirect if session is actually valid
  // This prevents the redirect loop when cookies exist but session is invalid

  // For all other routes, let client-side handle auth
  // The AuthGuard component will:
  // 1. Show loading state while verifying session
  // 2. Redirect to login if session is invalid
  // 3. Clear stale localStorage/cookies if needed
  return response;
}

// Config moved to middleware.ts if needed
