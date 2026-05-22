"use client";

import { useEffect, useMemo, useRef } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import type { OnboardingState } from "../services/onboarding-service";

const ONBOARDING_KEY = ["onboarding"] as const;
const ONBOARDING_UPDATED_EVENT = "general.onboarding.updated";

type OnboardingWsMessage = {
  type?: string;
  event?: string;
  data?: {
    business_type?: string;
    completed?: boolean;
  };
  payload?: {
    business_type?: string;
    completed?: boolean;
  };
};

function getWsUrl(): string {
  const apiOrigin = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8087";
  const wsOrigin = apiOrigin.replace(/^http/, "ws");
  return `${wsOrigin}/api/v1/general/onboarding/ws`;
}

function nextDelay(attempt: number): number {
  return Math.min(1000 * 2 ** attempt, 30_000);
}

export function useOnboardingWebSocket() {
  const queryClient = useQueryClient();
  const { isAuthenticated } = useAuthStore();
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectAttemptRef = useRef(0);
  const unmountedRef = useRef(false);
  const lastInvalidateRef = useRef(0);

  const wsUrl = useMemo(() => getWsUrl(), []);

  useEffect(() => {
    if (!isAuthenticated) {
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current);
        reconnectTimerRef.current = null;
      }
      wsRef.current?.close();
      wsRef.current = null;
      return;
    }

    unmountedRef.current = false;

    function connect() {
      if (unmountedRef.current) return;

      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        reconnectAttemptRef.current = 0;
      };

      ws.onmessage = (ev) => {
        let message: OnboardingWsMessage;
        try {
          message = JSON.parse(ev.data as string) as OnboardingWsMessage;
        } catch {
          return;
        }

        const eventType = message.event ?? message.type;
        if (eventType !== ONBOARDING_UPDATED_EVENT) return;

        const now = Date.now();
        if (now - lastInvalidateRef.current < 300) return;
        lastInvalidateRef.current = now;

        queryClient.invalidateQueries({ queryKey: ONBOARDING_KEY });
        queryClient.setQueryData<OnboardingState>(ONBOARDING_KEY, (current) => {
          if (!current) return current;
          const payload = message.data ?? message.payload;
          return {
            ...current,
            business_type: payload?.business_type ?? current.business_type,
            completed: payload?.completed ?? current.completed,
          };
        });
      };

      ws.onerror = () => {
        ws.close();
      };

      ws.onclose = () => {
        if (unmountedRef.current) return;
        const delay = nextDelay(reconnectAttemptRef.current);
        reconnectAttemptRef.current += 1;
        reconnectTimerRef.current = setTimeout(connect, delay);
      };
    }

    connect();

    return () => {
      unmountedRef.current = true;
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current);
        reconnectTimerRef.current = null;
      }
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, [isAuthenticated, queryClient, wsUrl]);
}
