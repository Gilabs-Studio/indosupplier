"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";

const MAX_RATE_LIMIT_WINDOW_SECONDS = 60 * 60;

interface RateLimitState {
  resetTime: number | null; // Unix timestamp in seconds (target time)
  setResetTime: (resetTime: number | null) => void;
  clearResetTime: () => void;
  getCountdownText: () => string; // Get current countdown text
}

/**
 * Format seconds into human-readable countdown string
 */
function formatCountdown(seconds: number): string {
  if (seconds <= 0) {
    return "a moment";
  }

  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;

  if (minutes > 0) {
    return `${minutes}m ${remainingSeconds}s`;
  } else {
    return `${remainingSeconds}s`;
  }
}

function sanitizeResetTime(resetTime: number | null): number | null {
  if (!resetTime || !Number.isFinite(resetTime)) {
    return null;
  }

  const now = Math.floor(Date.now() / 1000);
  if (resetTime <= now) {
    return null;
  }

  // Guard against malformed persisted values (for example ms treated as sec).
  if (resetTime > now + MAX_RATE_LIMIT_WINDOW_SECONDS) {
    return null;
  }

  return resetTime;
}

/**
 * Store for rate limit reset time (target time)
 * Used to track when rate limit will reset so we can show countdown
 *
 * Strategy: Only update reset time if:
 * - No reset time is set, OR
 * - Current reset time has already passed (expired), OR
 * - New reset time is significantly different (more than 5 seconds difference)
 * This prevents countdown from resetting when multiple requests hit rate limit
 * with slightly different reset times from backend
 *
 * Uses localStorage persistence to survive hard refresh
 */
export const useRateLimitStore = create<RateLimitState>()(
  persist(
    (set, get) => ({
      resetTime: null,
      setResetTime: (resetTime: number | null) => {
        const currentResetTime = get().resetTime;
        const now = Math.floor(Date.now() / 1000);

        if (!resetTime) {
          set({ resetTime: null });
          return;
        }

        // If no current reset time, set it
        if (!currentResetTime) {
          set({ resetTime });
          return;
        }

        // If current reset time has expired, update with new one
        if (currentResetTime <= now) {
          set({ resetTime });
          return;
        }

        // If new reset time is significantly different (more than 5 seconds), update
        // This handles cases where backend gives slightly different times
        const timeDiff = Math.abs(resetTime - currentResetTime);
        if (timeDiff > 5) {
          // Only update if new reset time is later (longer wait)
          if (resetTime > currentResetTime) {
            set({ resetTime });
          }
        }
        // Otherwise, keep current reset time (prevent reset)
      },
      clearResetTime: () => {
        set({ resetTime: null });
      },
      getCountdownText: () => {
        const resetTime = get().resetTime;
        if (!resetTime) {
          return "a moment";
        }

        const now = Math.floor(Date.now() / 1000);
        const secondsRemaining = Math.max(0, resetTime - now);

        // Auto-clear if expired
        if (secondsRemaining <= 0) {
          get().clearResetTime();
          return "a moment";
        }

        return formatCountdown(secondsRemaining);
      },
    }),
    {
      name: "rate-limit-store", // localStorage key
      // Only persist valid future reset times to avoid sticky lockouts.
      partialize: (state) => {
        return { resetTime: sanitizeResetTime(state.resetTime) };
      },
      // Auto-clear invalid or expired reset time when loading from localStorage.
      onRehydrateStorage: () => {
        return (state) => {
          if (state) {
            state.resetTime = sanitizeResetTime(state.resetTime);
          }
        };
      },
    },
  ),
);
