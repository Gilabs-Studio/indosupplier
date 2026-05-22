"use client";

import { useEffect, useRef } from "react";
import { toast } from "sonner";
import { useRateLimitStore } from "../stores/useRateLimitStore";
import { formatError } from "../i18n/error-messages";

/**
 * Format seconds into human-readable countdown string
 * Examples: "15m 30s", "2m 0s", "45s", "0s"
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

/**
 * Calculate seconds remaining from target time (Unix timestamp in seconds)
 * Simple: targetTime - currentTime
 */
function calculateSecondsRemaining(targetTime: number): number {
  const now = Math.floor(Date.now() / 1000);
  return Math.max(0, targetTime - now);
}

/**
 * Optimized hook for rate limit countdown
 *
 * Strategy:
 * 1. Store only the target reset time (Unix timestamp)
 * 2. Calculate countdown from current time to target time every second
 * 3. Single interval per component instance (lightweight)
 * 4. Auto-cleanup when countdown reaches 0
 *
 * This is optimal for large-scale projects because:
 * - Minimal state (only target time)
 * - Simple calculation (target - now)
 * - No complex logic or multiple intervals
 * - Efficient cleanup
 */
export function useRateLimitCountdown() {
  const resetTime = useRateLimitStore((state) => state.resetTime);
  const clearResetTime = useRateLimitStore((state) => state.clearResetTime);
  const intervalRef = useRef<NodeJS.Timeout | null>(null);
  const toastIdRef = useRef<string | number | null>(null);

  useEffect(() => {
    // Cleanup previous interval
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }

    // If no target time, dismiss toast and return
    if (!resetTime) {
      if (toastIdRef.current) {
        toast.dismiss(toastIdRef.current);
        toastIdRef.current = null;
      }
      return;
    }

    // Calculate initial countdown: targetTime - now
    let secondsRemaining = calculateSecondsRemaining(resetTime);

    // If already expired, clear and return
    if (secondsRemaining <= 0) {
      clearResetTime();
      if (toastIdRef.current) {
        toast.dismiss(toastIdRef.current);
        toastIdRef.current = null;
      }
      return;
    }

    // Show initial toast
    const showToast = (seconds: number) => {
      const countdownText = formatCountdown(seconds);
      const msg = formatError("backend", "rateLimit", {
        countdown: countdownText,
      });

      toastIdRef.current = toast.error(msg.title, {
        description: msg.description,
        duration: Infinity,
        id: "rate-limit-error",
      });
    };

    showToast(secondsRemaining);

    // Update countdown every second
    // Simple: recalculate targetTime - now
    intervalRef.current = setInterval(() => {
      // Get current target time from store (may have changed)
      const currentTargetTime = useRateLimitStore.getState().resetTime;

      if (!currentTargetTime) {
        // Target cleared, stop countdown
        if (intervalRef.current) {
          clearInterval(intervalRef.current);
          intervalRef.current = null;
        }
        if (toastIdRef.current) {
          toast.dismiss(toastIdRef.current);
          toastIdRef.current = null;
        }
        return;
      }

      // Recalculate: targetTime - now
      secondsRemaining = calculateSecondsRemaining(currentTargetTime);

      if (secondsRemaining <= 0) {
        // Countdown finished
        clearResetTime();
        if (intervalRef.current) {
          clearInterval(intervalRef.current);
          intervalRef.current = null;
        }
        if (toastIdRef.current) {
          toast.dismiss(toastIdRef.current);
          toastIdRef.current = null;
        }
        return;
      }

      // Update toast with new countdown
      showToast(secondsRemaining);
    }, 1000);

    // Cleanup on unmount or when resetTime changes
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [resetTime, clearResetTime]);
}
