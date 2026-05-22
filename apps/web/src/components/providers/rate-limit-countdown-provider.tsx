"use client";

import { useRateLimitCountdown } from "@/lib/hooks/useRateLimitCountdown";

/**
 * Client component provider for rate limit countdown
 * This component uses the useRateLimitCountdown hook to manage
 * real-time countdown updates for rate limit errors
 */
export function RateLimitCountdownProvider() {
  useRateLimitCountdown();
  return null; // This component doesn't render anything
}
