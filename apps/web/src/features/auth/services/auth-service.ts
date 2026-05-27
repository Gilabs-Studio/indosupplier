import apiClient, {
  clearCSRFTokenMemory,
  emitAuthTelemetry,
  setCSRFTokenMemory,
} from "@/lib/api-client";
import type { LoginRequest, LoginResponse } from "../types";

const REFRESH_SINGLE_FLIGHT_RESET_MS = 500;

let inFlightRefreshPromise: Promise<LoginResponse> | null = null;
let refreshPromiseResetTimer: ReturnType<typeof setTimeout> | null = null;

function isCSRFInvalidError(error: unknown): boolean {
  const axiosError = error as {
    response?: { status?: number; data?: { error?: { code?: string } } };
  };

  return (
    axiosError?.response?.status === 403 &&
    axiosError?.response?.data?.error?.code === "CSRF_INVALID"
  );
}

function getAxiosStatus(error: unknown): number | null {
  const axiosError = error as { response?: { status?: number } };
  const status = axiosError?.response?.status;
  return typeof status === "number" ? status : null;
}

function isTransientRefreshError(error: unknown): boolean {
  const status = getAxiosStatus(error);
  return status !== null && status >= 500;
}

async function executeRefreshRequest(): Promise<LoginResponse> {
  try {
    const response = await apiClient.post<LoginResponse>("/auth/refresh-token");
    return response.data;
  } catch (error: unknown) {
    if (isCSRFInvalidError(error)) {
      emitAuthTelemetry("csrf_invalid_retry", {
        endpoint: "/auth/refresh-token",
        source: "authService.refreshToken",
      });

      clearCSRFTokenMemory();
      const csrfToken = await authService.prefetchCSRFToken();

      try {
        const retryResponse = await apiClient.post<LoginResponse>(
          "/auth/refresh-token",
          {},
          csrfToken ? { headers: { "X-CSRF-Token": csrfToken } } : undefined,
        );
        return retryResponse.data;
      } catch (retryError: unknown) {
        emitAuthTelemetry("csrf_retry_failed", {
          endpoint: "/auth/refresh-token",
          source: "authService.refreshToken",
        });
        throw retryError;
      }
    }

    if (!isTransientRefreshError(error)) {
      throw error;
    }

    const retryResponse = await apiClient.post<LoginResponse>("/auth/refresh-token");
    return retryResponse.data;
  }
}

function scheduleRefreshPromiseReset(): void {
  if (refreshPromiseResetTimer) {
    clearTimeout(refreshPromiseResetTimer);
  }

  refreshPromiseResetTimer = setTimeout(() => {
    inFlightRefreshPromise = null;
    refreshPromiseResetTimer = null;
  }, REFRESH_SINGLE_FLIGHT_RESET_MS);
}

function refreshTokenOnce(): Promise<LoginResponse> {
  if (inFlightRefreshPromise) {
    return inFlightRefreshPromise;
  }

  inFlightRefreshPromise = executeRefreshRequest().finally(() => {
    scheduleRefreshPromiseReset();
  });

  return inFlightRefreshPromise;
}

export const authService = {
  async prefetchCSRFToken(): Promise<string | null> {
    const response = await apiClient.get<{
      data: { csrf_token?: string; message?: string };
    }>("/auth/csrf");

    const bodyToken: string | null = response.data?.data?.csrf_token ?? null;
    const headerToken: string | null =
      (typeof response.headers.get === "function"
        ? (response.headers.get("x-csrf-token") as string | null)
        : null) ??
      (response.headers["x-csrf-token"] as string | undefined) ??
      (response.headers["X-CSRF-Token"] as string | undefined) ??
      null;

    const csrfToken = bodyToken ?? headerToken;

    if (csrfToken) {
      setCSRFTokenMemory(csrfToken);
    }

    return csrfToken;
  },

  async login(
    credentials: LoginRequest,
    csrfToken?: string | null,
  ): Promise<LoginResponse> {
    const response = await apiClient.post<LoginResponse>(
      "/auth/login",
      credentials,
      csrfToken ? { headers: { "X-CSRF-Token": csrfToken } } : undefined,
    );
    return response.data;
  },

  async refreshToken(): Promise<LoginResponse> {
    return refreshTokenOnce();
  },

  async getMe(): Promise<LoginResponse> {
    return this.refreshToken();
  },

  async logout(): Promise<void> {
    await apiClient.post("/auth/logout");
  },
};
