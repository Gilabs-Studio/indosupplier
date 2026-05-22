"use client";

import { useEffect, useRef } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import { notificationQueryKeys } from "./use-notifications";
import { useNotificationStore } from "../stores/use-notification-store";
import { playNotificationSound } from "@/lib/notification-sound";
import type { ListUsersResponse, User } from "@/features/master-data/user-management/types";
import type { WebSocketMessage, Notification } from "../types";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
const WS_URL = API_BASE_URL.replace(/^http/, "ws");

function extractEmailFromNotificationMessage(message: string): string | null {
  const match = message.match(/\(([^)]+@[^)]+)\)/);
  return match?.[1]?.toLowerCase() ?? null;
}

export function useWebSocket() {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const processedNotificationIdsRef = useRef<Set<string>>(new Set());
  const isConnectingRef = useRef(false);
  const queryClient = useQueryClient();
  const queryClientRef = useRef(queryClient);
  const { isAuthenticated, user } = useAuthStore();
  const addRealtimeNotification = useNotificationStore((state) => state.addRealtimeNotification);
  const isAuthenticatedRef = useRef(isAuthenticated);

  // Update refs when values change
  useEffect(() => {
    isAuthenticatedRef.current = isAuthenticated;
    queryClientRef.current = queryClient;
  }, [isAuthenticated, queryClient]);

  useEffect(() => {
    // Don't connect if not authenticated
    if (!isAuthenticated) {
      // Close connection if logged out
      if (wsRef.current && wsRef.current.readyState !== WebSocket.CLOSED) {
        try {
          wsRef.current.close(1000, "Logged out");
        } catch {
          // Ignore errors
        }
        wsRef.current = null;
      }
      return;
    }

    // Don't connect if already connecting
    if (isConnectingRef.current) {
      return;
    }

    // Don't reconnect if already connected
    const currentWs = wsRef.current;
    if (
      currentWs &&
      (currentWs.readyState === WebSocket.OPEN || currentWs.readyState === WebSocket.CONNECTING)
    ) {
      return;
    }

    // Clear old connection ref
    if (currentWs) {
      if (currentWs.readyState === WebSocket.CLOSED || currentWs.readyState === WebSocket.CLOSING) {
        wsRef.current = null;
      } else {
        return;
      }
    }

    isConnectingRef.current = true;
    reconnectAttemptsRef.current = 0;

    // Build WebSocket URL - cookies will be sent automatically
    const wsUrl = `${WS_URL}/api/v1/ws/notifications`;

    // Store message handler for reuse
    const messageHandler = (event: MessageEvent) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        const eventType = message.type ?? message.event;

        switch (eventType) {
          case "notification.created": {
            const notification = message.data as Notification;

            // Keep page-1 merged view reactive without forcing list refetch.
            addRealtimeNotification(notification);

            const processedIds = processedNotificationIdsRef.current;
            const isFirstProcessing = !processedIds.has(notification.id);
            if (isFirstProcessing) {
              processedIds.add(notification.id);
              if (processedIds.size > 500) {
                const oldest = processedIds.values().next().value;
                if (oldest) {
                  processedIds.delete(oldest);
                }
              }
            }

            queryClientRef.current.setQueryData<number>(notificationQueryKeys.unreadCount, (current) =>
              Math.max(0, (current ?? 0) + (notification.is_read ? 0 : 1))
            );

            // Keep user-management table in sync when password reset requests are created.
            if (notification.entity_type === "password_reset_request") {
              const requestedEmail = extractEmailFromNotificationMessage(notification.message);

              if (requestedEmail) {
                queryClientRef.current.setQueriesData<ListUsersResponse>(
                  { queryKey: ["users"] },
                  (current) => {
                    if (!current) {
                      return current;
                    }

                    return {
                      ...current,
                      data: current.data.map((row: User) =>
                        row.email.toLowerCase() === requestedEmail
                          ? { ...row, password_reset_pending: true }
                          : row
                      ),
                    };
                  }
                );
              }

              queryClientRef.current.invalidateQueries({ queryKey: ["users"] });
            }

            if (isFirstProcessing) {
              toast.info(notification.title, {
                description: notification.message,
                duration: 5000,
              });

              if (!notification.is_read) {
                playNotificationSound();
              }
            }
            break;
          }

          case "notification.updated":
          case "notification.deleted": {
            // Invalidate queries to refresh data
            queryClientRef.current.invalidateQueries({
              predicate: (query) =>
                query.queryKey[0] === "notifications" && query.queryKey[1] !== "unread-count",
            });
            queryClientRef.current.invalidateQueries({ queryKey: notificationQueryKeys.unreadCount });
            break;
          }

          default:
            console.warn("Unknown WebSocket message type:", message.type);
        }
      } catch (error) {
        console.error("Error parsing WebSocket message:", error);
      }
    };

    try {
      const ws = new WebSocket(wsUrl);

      ws.onopen = () => {
        isConnectingRef.current = false;
        reconnectAttemptsRef.current = 0;
        
        // Clear any pending reconnect
        if (reconnectTimeoutRef.current) {
          clearTimeout(reconnectTimeoutRef.current);
          reconnectTimeoutRef.current = null;
        }
      };

      ws.onmessage = messageHandler;

      ws.onerror = () => {
        isConnectingRef.current = false;
      };

      ws.onclose = (event) => {
        isConnectingRef.current = false;
        
        // Clear ref only if this is the current connection
        if (wsRef.current === ws) {
          wsRef.current = null;
        }

        // Only reconnect if not a normal closure (1000) and we're authenticated
        if (event.code !== 1000 && isAuthenticatedRef.current) {
          // Exponential backoff: 1s, 2s, 4s, 8s, max 30s
          const maxAttempts = 5;
          if (reconnectAttemptsRef.current < maxAttempts) {
            reconnectAttemptsRef.current += 1;
            const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current - 1), 30000);
            
            reconnectTimeoutRef.current = setTimeout(() => {
              // Will trigger reconnect via effect
            }, delay);
          } else {
            console.warn("Max reconnection attempts reached. Stopping reconnection.");
          }
        }
      };

      wsRef.current = ws;
    } catch (error) {
      console.error("Failed to create WebSocket connection:", error);
      isConnectingRef.current = false;
    }

    return () => {
      // Cleanup on unmount
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
        reconnectTimeoutRef.current = null;
      }
    };
  }, [isAuthenticated, user?.id, addRealtimeNotification]);
}
