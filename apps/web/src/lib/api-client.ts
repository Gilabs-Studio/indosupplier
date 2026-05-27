import axios, {
  AxiosInstance,
  AxiosError,
  AxiosResponseHeaders,
  InternalAxiosRequestConfig,
  RawAxiosResponseHeaders,
} from "axios";
import type { User } from "@/features/auth/types";
import { toast } from "sonner";
import { formatError } from "./i18n/error-messages";
import { useRateLimitStore } from "./stores/useRateLimitStore";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8088";
const CSRF_BOOTSTRAP_COOLDOWN_MS = 2000;
const GLOBAL_TOAST_DEDUPE_WINDOW_MS = 2500;
const MAX_RATE_LIMIT_WINDOW_SECONDS = 60 * 60;

const globalToastLastShownAt = new Map<string, number>();

// Memory cache for CSRF token to support cross-origin API calls
let memoryCsrfToken: string | null = null;
let csrfBootstrapPromise: Promise<string | null> | null = null;
let csrfBootstrapCooldownUntil = 0;

type AuthTelemetryEvent =
  | "csrf_bootstrap_failed"
  | "csrf_invalid_retry"
  | "csrf_retry_failed";

export function emitAuthTelemetry(
  event: AuthTelemetryEvent,
  details?: Record<string, unknown>,
): void {
  if (typeof window !== "undefined") {
    window.dispatchEvent(
      new CustomEvent("indosupplier:auth-telemetry", {
        detail: {
          event,
          details,
          timestamp: new Date().toISOString(),
        },
      }),
    );
  }

  if (process.env.NODE_ENV !== "production") {
    console.warn("[auth-telemetry]", event, details ?? {});
  }
}

const csrfBootstrapClient = axios.create({
  baseURL: `${API_BASE_URL}/api/v1`,
  headers: {
    "Content-Type": "application/json",
  },
  timeout: 10000,
  withCredentials: true,
});

function isMutationMethod(method?: string): boolean {
  if (!method) return false;
  return ["post", "put", "patch", "delete"].includes(method.toLowerCase());
}

function normalizeQueryParams(value: unknown): unknown {
  if (typeof value === "string") {
    const trimmed = value.trim();
    return trimmed.length > 0 ? trimmed : undefined;
  }

  if (Array.isArray(value)) {
    return value
      .map((item) => normalizeQueryParams(item))
      .filter((item) => item !== undefined);
  }

  if (value && typeof value === "object" && !(value instanceof Date)) {
    const normalized: Record<string, unknown> = {};
    for (const [key, nested] of Object.entries(value as Record<string, unknown>)) {
      const next = normalizeQueryParams(nested);
      if (next !== undefined) {
        normalized[key] = next;
      }
    }
    return normalized;
  }

  return value;
}

function normalizeRateLimitResetTime(resetHeader: string | number): number | null {
  const now = Math.floor(Date.now() / 1000);
  let parsed =
    typeof resetHeader === "number"
      ? resetHeader
      : Number.parseInt(String(resetHeader), 10);

  if (!Number.isFinite(parsed) || parsed <= 0) {
    return null;
  }

  // Some proxies send Unix milliseconds; normalize to seconds.
  if (parsed > 1_000_000_000_000) {
    parsed = Math.floor(parsed / 1000);
  }

  // If the value is tiny, treat it as delta-seconds.
  if (parsed < 1_000_000_000) {
    parsed = now + parsed;
  }

  // Ignore obviously stale or unrealistic future values to avoid sticky lockouts.
  if (parsed <= now || parsed > now + MAX_RATE_LIMIT_WINDOW_SECONDS) {
    return null;
  }

  return parsed;
}

type ServerUser = {
  id: string;
  name: string;
  email: string;
} | null | undefined;

function normalizeUserResponse(rawUser: ServerUser): User | null {
  if (!rawUser) return null;

  return {
    id: rawUser.id,
    name: rawUser.name,
    email: rawUser.email,
  };
}

function showGlobalErrorToast(
  title: string,
  options?: {
    description?: string;
    dedupeKey?: string;
  },
): void {
  const baseKey = options?.dedupeKey ?? `${title}::${options?.description ?? ""}`;
  const dedupeKey = baseKey.slice(0, 240);
  const now = Date.now();
  const lastShownAt = globalToastLastShownAt.get(dedupeKey) ?? 0;

  if (now - lastShownAt < GLOBAL_TOAST_DEDUPE_WINDOW_MS) {
    return;
  }

  globalToastLastShownAt.set(dedupeKey, now);
  toast.error(title, {
    description: options?.description,
    id: `global-error:${dedupeKey}`,
  });
}

/**
 * Robustly extract CSRF token from an Axios headers object.
 *
 * Axios v1+ uses an AxiosHeaders instance whose internal storage may not be
 * enumerable via `for...in`. The `.get()` method is the canonical API for
 * case-insensitive lookup, so we try it first before falling back to bracket
 * access and the slower `for...in` walk.
 */
function extractCsrfFromHeaders(
  headers: RawAxiosResponseHeaders | AxiosResponseHeaders | null | undefined,
): string | null {
  if (!headers) return null;

  // Primary: AxiosHeaders v1+ case-insensitive .get()
  if (typeof headers.get === "function") {
    const val = headers.get("x-csrf-token");
    if (val) return String(val);
  }

  // Fallback: plain object bracket access (works for both casings)
  if (headers["x-csrf-token"]) return String(headers["x-csrf-token"]);
  if (headers["X-CSRF-Token"]) return String(headers["X-CSRF-Token"]);

  // Last resort: enumerate for any remaining case variant
  for (const key in headers) {
    if (Object.prototype.hasOwnProperty.call(headers, key) &&
        key.toLowerCase() === "x-csrf-token") {
      return String(headers[key]);
    }
  }
  return null;
}

/**
 * Get CSRF token from memory or cookie.
 * The csrf_token is exposed by the API via the X-CSRF-Token header.
 */
export function getCSRFToken(): string | null {
  if (memoryCsrfToken) return memoryCsrfToken;
  if (typeof document === "undefined") return null;
  // Fallback to cookie if same-origin scenario
  const match = document.cookie.match(/(?:^|;\s*)indosupplier_csrf_token=([^;]*)/);
  return match ? decodeURIComponent(match[1]) : null;
}

/**
 * Manually set the CSRF token in memory for cross-origin setups.
 */
export function setCSRFTokenMemory(token: string): void {
  if (token) {
    memoryCsrfToken = token;
  }
}

export function clearCSRFTokenMemory(): void {
  memoryCsrfToken = null;
}

/**
 * Ensure CSRF token exists before unsafe requests.
 *
 * This prevents first-refresh failures on cold page loads in cross-origin
 * production setups where document.cookie cannot read API cookies.
 */
async function ensureCSRFToken(options?: {
  forceRefresh?: boolean;
  reason?: string;
}): Promise<string | null> {
  if (options?.forceRefresh) {
    clearCSRFTokenMemory();
  }

  const existing = getCSRFToken();
  if (existing) {
    return existing;
  }

  // Double-Submit Cookie only applies to browser requests.
  if (typeof window === "undefined") {
    return null;
  }

  if (Date.now() < csrfBootstrapCooldownUntil) {
    return null;
  }

  // Single-flight: avoid parallel /auth/csrf calls returning different tokens.
  if (csrfBootstrapPromise) {
    return csrfBootstrapPromise;
  }

  csrfBootstrapPromise = csrfBootstrapClient
    .get<{ data?: { csrf_token?: string } }>("/auth/csrf")
    .then((response) => {
      const bodyToken = response.data?.data?.csrf_token ?? null;
      const headerToken = extractCsrfFromHeaders(response.headers);
      const nextToken = bodyToken ?? headerToken;

      if (nextToken) {
        memoryCsrfToken = nextToken;
        csrfBootstrapCooldownUntil = 0;
      }

      return nextToken;
    })
    .catch((bootstrapError: unknown) => {
      csrfBootstrapCooldownUntil = Date.now() + CSRF_BOOTSTRAP_COOLDOWN_MS;
      emitAuthTelemetry("csrf_bootstrap_failed", {
        reason: options?.reason ?? "unknown",
        cooldown_ms: CSRF_BOOTSTRAP_COOLDOWN_MS,
        message:
          bootstrapError instanceof Error
            ? bootstrapError.message
            : "bootstrap_failed",
      });
      return null;
    })
    .finally(() => {
      csrfBootstrapPromise = null;
    });

  return csrfBootstrapPromise;
}

export const apiClient: AxiosInstance = axios.create({
  baseURL: `${API_BASE_URL}/api/v1`,
  headers: {
    "Content-Type": "application/json",
  },
  timeout: 10000,
  withCredentials: true, // IMPORTANT: Send and receive cookies
});

// Flag to prevent multiple refresh attempts
let isRefreshing = false;
const failedQueue: Array<{
  resolve: (value?: unknown) => void;
  reject: (error?: unknown) => void;
}> = [];

const processQueue = (error: AxiosError | null) => {
  const queue = [...failedQueue];
  failedQueue.splice(0, failedQueue.length);
  queue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve();
    }
  });
};

// Request interceptor for CSRF token
apiClient.interceptors.request.use(
  async (config: InternalAxiosRequestConfig) => {
    if (config.params !== undefined) {
      config.params = normalizeQueryParams(config.params) as InternalAxiosRequestConfig["params"];
    }

    // Add CSRF token header for unsafe methods (POST, PUT, PATCH, DELETE)
    const unsafeMethods = ["POST", "PUT", "PATCH", "DELETE"];
    if (config.method && unsafeMethods.includes(config.method.toUpperCase())) {
      let csrfToken = getCSRFToken();
      if (!csrfToken) {
        csrfToken = await ensureCSRFToken();
      }

      if (csrfToken && config.headers) {
        config.headers["X-CSRF-Token"] = csrfToken;
      }
    }

    return config;
  },
  (error: AxiosError) => {
    return Promise.reject(error);
  },
);

interface ErrorDetails {
  field?: string;
  resource?: string;
  value?: string;
  [key: string]: unknown;
}

interface ApiErrorResponse {
  success: false;
  error: {
    code: string;
    message: string;
    details?: ErrorDetails;
    field_errors?: Array<{ field: string; message: string }>;
  };
  timestamp: string;
  request_id: string;
}

// Response interceptor for error handling
apiClient.interceptors.response.use(
  (response) => {
    // Read and cache CSRF token from headers (vital for cross-origin setups)
    const csrfHeader = extractCsrfFromHeaders(response.headers);
    if (csrfHeader) {
      memoryCsrfToken = csrfHeader;
    }

    // Clear rate limit reset time on successful response
    if (response.status !== 429) {
      const currentResetTime = useRateLimitStore.getState().resetTime;
      if (currentResetTime) {
        useRateLimitStore.getState().clearResetTime();
      }
    }
    return response;
  },
  async (error: AxiosError<ApiErrorResponse>) => {
    // Try to extract CSRF token even from error responses
    if (error.response?.headers) {
      const csrfHeader = extractCsrfFromHeaders(error.response.headers);
      if (csrfHeader) {
        memoryCsrfToken = csrfHeader;
      }
    }

    const originalRequest = error.config as InternalAxiosRequestConfig & {
      _retry?: boolean;
    };
    const requestUrl = originalRequest?.url || "";
    const requestMethod = originalRequest?.method;

    // Skip toast for auth endpoints - these handle their own errors silently
    // 401/403 is expected when checking session or after logout
    const isAuthEndpoint =
      requestUrl.includes("/auth/refresh") ||
      requestUrl.includes("/auth/login") ||
      requestUrl.includes("/auth/logout");

    // Mutations usually show contextual toasts at feature level.
    // Suppress generic global toasts for mutation failures to avoid duplicate notifications.
    const shouldSuppressGlobalToast =
      isMutationMethod(requestMethod) && !isAuthEndpoint;

    const notifyGlobalError = (
      title: string,
      description?: string,
      dedupeKey?: string,
    ) => {
      if (shouldSuppressGlobalToast) {
        return;
      }

      showGlobalErrorToast(title, { description, dedupeKey });
    };

    // Network error
    if (!error.response) {
      if (isAuthEndpoint) {
        return Promise.reject(error);
      }
      if (error.code === "ECONNABORTED") {
        const msg = formatError("network", "timeout");
        notifyGlobalError(msg.title, msg.description, "network-timeout");
      } else if (
        error.code === "ERR_NETWORK" ||
        error.message === "Network Error"
      ) {
        const msg = formatError("network", "connectionFailed");
        notifyGlobalError(msg.title, msg.description, "network-connection-failed");
      } else {
        const msg = formatError("network", "generic");
        notifyGlobalError(msg.title, msg.description, "network-generic");
      }
      return Promise.reject(error);
    }

    const status = error.response.status;
    const errorData = error.response.data;

    // Skip toast for auth endpoints on any error except 429 (rate limit).
    // /auth/refresh-token, /auth/login, /auth/logout all handle errors internally.
    // The backend may return 500 for an expired/invalid refresh token (instead of 401),
    // so we must silence ALL non-429 errors from these endpoints to prevent spurious toasts
    // when useLoginGuard probes session validity on the login page.
    if (isAuthEndpoint && status !== 429) {
      return Promise.reject(error);
    }

    const normalizedError = (() => {
      if (!errorData || typeof errorData !== "object") {
        return null;
      }

      const legacy = (errorData as {
        error?: {
          code?: string;
          details?: Record<string, unknown>;
          field_errors?: Array<{ field?: string; message?: string }>;
          message?: string;
        };
      }).error;

      if (legacy && typeof legacy === "object") {
        return {
          code: legacy.code,
          details: legacy.details,
          fieldErrors: legacy.field_errors,
          message: legacy.message,
        };
      }

      const modern = errorData as unknown as {
        code?: string;
        details?: ErrorDetails;
        error?: string;
        message?: string;
      };

      if (typeof modern.code === "string") {
        return {
          code: modern.code,
          details: modern.details,
          fieldErrors: undefined,
          message: typeof modern.error === "string" ? modern.error : modern.message,
        };
      }

      return null;
    })();

    if (!normalizedError) {
      const msg = formatError("backend", "invalidFormat");
      notifyGlobalError(msg.title, msg.description, "backend-invalid-format");
      return Promise.reject(error);
    }

    const errorCode = normalizedError.code;
    const errorDetails = normalizedError.details;
    const fieldErrors = normalizedError.fieldErrors;

    // Handle CSRF errors
    if (errorCode === "CSRF_INVALID") {
      // CSRF token invalid - try to get a new one
      const msg = formatError("backend", "csrfError");
      notifyGlobalError(
        msg.title || "Session expired",
        msg.description || "Please try again.",
        "backend-csrf-invalid",
      );
      return Promise.reject(error);
    }

    // Handle resource conflicts (including explicit backend codes)
    if (
      errorCode === "RESOURCE_ALREADY_EXISTS" ||
      errorCode === "CONFLICT" ||
      errorCode === "EMAIL_ALREADY_TAKEN" ||
      errorCode === "COMPANY_NAME_TAKEN"
    ) {
      if (errorCode === "EMAIL_ALREADY_TAKEN" && errorDetails?.field === "email") {
        const msg = formatError("backend", "emailExists", {
          email: String(errorDetails.value || ""),
        });
        notifyGlobalError(msg.title, msg.description, "backend-email-exists");
      } else if (
        errorDetails?.field === "email" &&
        errorDetails?.resource === "user"
      ) {
        const msg = formatError("backend", "emailExists", {
          email: String(errorDetails.value || ""),
        });
        notifyGlobalError(msg.title, msg.description, "backend-email-exists");
      } else if (errorDetails?.field && errorDetails?.resource) {
        const fieldName = typeof errorDetails.field === "string" ? errorDetails.field : "field";
        const msg = formatError("backend", "resourceExists", {
          field: fieldName,
          value: String(errorDetails.value || ""),
        });
        notifyGlobalError(msg.title, msg.description, "backend-resource-exists");
      } else {
        const msg = formatError("backend", "conflict");
        notifyGlobalError(msg.title, msg.description, "backend-conflict");
      }
      return Promise.reject(error);
    }

    // Handle internal server error with details
    if (errorCode === "INTERNAL_SERVER_ERROR" && errorDetails) {
      if (errorDetails.field === "email" && errorDetails.resource === "user") {
        const msg = formatError("backend", "emailExists", {
          email: String(errorDetails.value || ""),
        });
        notifyGlobalError(msg.title, msg.description, "backend-email-exists");
        return Promise.reject(error);
      }
      if (errorDetails.field && errorDetails.resource) {
        const fieldName = typeof errorDetails.field === "string" ? errorDetails.field : "field";
        const msg = formatError("backend", "resourceExists", {
          field: fieldName,
          value: String(errorDetails.value || ""),
        });
        notifyGlobalError(msg.title, msg.description, "backend-resource-exists");
        return Promise.reject(error);
      }
    }

    // Handle validation errors
    if (errorCode === "VALIDATION_ERROR" && fieldErrors && fieldErrors.length > 0) {
      const firstError = fieldErrors[0];
      const fieldName = typeof firstError.field === "string" ? firstError.field : "field";
      const fieldMessage = typeof firstError.message === "string" ? firstError.message : "Invalid value";
      const msg = formatError("backend", "fieldError", {
        field: fieldName,
        message: fieldMessage,
      });
      notifyGlobalError(msg.title, msg.description, "backend-field-error");
      return Promise.reject(error);
    }

    // Handle 401 Unauthorized
    if (status === 401) {
      const originalRequest = error.config as InternalAxiosRequestConfig & {
        _retry?: boolean;
        skipAuthRedirectOn401?: boolean;
      };
      const requestUrl = originalRequest?.url || "";

      // Skip token refresh for authentication endpoints
      if (
        requestUrl.includes("/auth/login") ||
        requestUrl.includes("/auth/refresh")
      ) {
        return Promise.reject(error);
      }

      // When requested (e.g. mutations), only show toast and reject — no refresh, no logout
      if (originalRequest?.skipAuthRedirectOn401) {
        if (typeof window !== "undefined") {
          const msg = formatError("backend", "unauthorized");
          showGlobalErrorToast(msg.title, {
            description: msg.description,
            dedupeKey: "backend-unauthorized",
          });
        }
        return Promise.reject(error);
      }

      // Skip refresh if this is already a retry
      if (originalRequest?._retry) {
        // Refresh failed, logout user
        if (typeof window !== "undefined") {
          const msg = formatError("backend", "unauthorized");
          showGlobalErrorToast(msg.title, {
            description: msg.description,
            dedupeKey: "backend-unauthorized",
          });

          // Clear all auth state and cookies
          import("@/features/auth/stores/use-auth-store").then(({ useAuthStore }) => {
            useAuthStore.getState().logout();
          });
          import("@/features/auth/utils/clear-auth-cookies").then(({ fullAuthCleanup }) => {
            fullAuthCleanup();
          });

          setTimeout(() => {
            // Extract locale from current path (/en/... or /id/...) so we land on the right login
            const segments = window.location.pathname.split("/").filter(Boolean);
            const locale = ["en", "id"].includes(segments[0]) ? segments[0] : "en";
            window.location.href = `/${locale}/login`;
          }, 1000);
        }
        return Promise.reject(error);
      }

      // Try to refresh token using cookies
      if (!isRefreshing) {
        isRefreshing = true;

        return import("@/features/auth/services/auth-service")
          .then(({ authService }) => authService.refreshToken())
          .then((refreshResponse) => {
            if (refreshResponse.success && refreshResponse.data) {
              // Update auth store with new user data
              import("@/features/auth/stores/use-auth-store").then(
                ({ useAuthStore }) => {
                  const normalized = normalizeUserResponse(refreshResponse.data?.user ?? null);
                  useAuthStore.getState().setUser(normalized);
                  if (normalized) {
                    useAuthStore.setState({ isAuthenticated: true });
                  }
                },
              );

              originalRequest._retry = true;
              processQueue(null);
              isRefreshing = false;

              // Retry original request
              return apiClient(originalRequest);
            }

            throw new Error("Refresh token failed");
          })
          .catch((refreshError) => {
            isRefreshing = false;
            if (typeof window !== "undefined") {
              const msg = formatError("backend", "unauthorized");
              showGlobalErrorToast(msg.title, {
                description: msg.description,
                dedupeKey: "backend-unauthorized",
              });

              // Clear all auth state and cookies
              import("@/features/auth/stores/use-auth-store").then(({ useAuthStore }) => {
                useAuthStore.getState().logout();
              });
              import("@/features/auth/utils/clear-auth-cookies").then(({ fullAuthCleanup }) => {
                fullAuthCleanup();
              });

              setTimeout(() => {
                // Extract locale from current path (/en/... or /id/...) so we land on the right login
                const segments = window.location.pathname.split("/").filter(Boolean);
                const locale = ["en", "id"].includes(segments[0]) ? segments[0] : "en";
                window.location.href = `/${locale}/login`;
              }, 1000);
            }
            processQueue(refreshError as AxiosError);
            return Promise.reject(refreshError);
          });
      } else {
        // Already refreshing, queue this request
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        })
          .then(() => {
            originalRequest._retry = true;
            return apiClient(originalRequest);
          })
          .catch((err) => Promise.reject(err));
      }
    } else if (status === 402) {
      const subscription = errorDetails?.subscription as
        | {
            force_billing_redirect?: boolean;
          }
        | undefined;

      const msg = formatError("backend", "forbidden");
      notifyGlobalError(
        "Payment required",
        errorData?.error?.message || msg.description,
        "backend-payment-required",
      );

      if (typeof window !== "undefined" && subscription?.force_billing_redirect) {
        const onProfileBilling = window.location.pathname.includes("/profile");
        if (!onProfileBilling) {
          const segments = window.location.pathname.split("/").filter(Boolean);
          const locale = ["en", "id"].includes(segments[0]) ? segments[0] : "en";
          window.location.href = `/${locale}/profile?tab=billing&billing_tab=seat`;
        }
      }
    } else if (status === 403) {
      if (errorCode === "ACCOUNT_SUSPENDED") {
        notifyGlobalError(
          "Account suspended",
          errorData?.error?.message || "Subscription is suspended due to unpaid billing.",
          "backend-account-suspended",
        );
      } else {
        const msg = formatError("backend", "forbidden");
        notifyGlobalError(msg.title, msg.description, "backend-forbidden");
      }
    } else if (status === 404) {
      const isMutation =
        error.config?.method &&
        ["post", "put", "patch", "delete"].includes(
          error.config.method.toLowerCase(),
        );
      if (isMutation) {
        const msg = formatError("backend", "notFound");
        notifyGlobalError(msg.title, msg.description, "backend-not-found");
      }
    } else if (status === 409) {
      const msg = formatError("backend", "conflict");
      notifyGlobalError(msg.title, msg.description, "backend-conflict");
    } else if (status === 422) {
      // 422 Unprocessable Entity — business rule violation (e.g. WAREHOUSE_HAS_STOCK).
      // Suppress the global toast so each caller can render its own contextual UI.
      return Promise.reject(error);
    } else if (status === 503) {
      const msg = formatError("backend", "serviceUnavailable");
      notifyGlobalError(msg.title, msg.description, "backend-service-unavailable");
    } else if (status === 429) {
      // Rate limit handling
      const headers = error.response?.headers || {};
      const resetHeader =
        headers["x-ratelimit-reset"] || headers["X-RateLimit-Reset"];

      if (resetHeader) {
        const resetTimeValue = normalizeRateLimitResetTime(
          resetHeader as string | number,
        );

        if (resetTimeValue !== null) {
          useRateLimitStore.getState().setResetTime(resetTimeValue);
        } else {
          useRateLimitStore.getState().clearResetTime();
        }
      }

      const rateLimitMessage =
        errorData?.error?.message ||
        "Too many login attempts. Please wait before trying again.";

      const customError = { ...error } as AxiosError<ApiErrorResponse>;
      customError.message = rateLimitMessage;

      if (customError.response?.data) {
        if (!customError.response.data.error) {
          customError.response.data.error = {
            code: "RATE_LIMIT_EXCEEDED",
            message: rateLimitMessage,
          };
        } else {
          customError.response.data.error.message = rateLimitMessage;
        }
      }

      // Show toast with countdown
      const countdown = useRateLimitStore.getState().getCountdownText();
      const msg = formatError("backend", "rateLimit", { countdown });
      showGlobalErrorToast(msg.title, {
        description: msg.description,
        dedupeKey: "backend-rate-limit",
      });

      return Promise.reject(customError);
    } else if (status >= 500) {
      const msg = formatError("backend", "serverError");
      notifyGlobalError(msg.title, msg.description, "backend-server-error");
    } else {
      const requestUrl = error.config?.url || "";
      if (!requestUrl.includes("/auth/login")) {
        const msg = formatError("backend", "unexpectedError");
        notifyGlobalError(msg.title, msg.description, "backend-unexpected-error");
      }
    }

    return Promise.reject(error);
  },
);

export default apiClient;
