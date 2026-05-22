"use client";

import { useMemo } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { notificationService } from "../services/notification.service";
import type { ListNotificationsParams, ListNotificationsResponse, Notification } from "../types";
import { isRealtimeNotificationId, useNotificationStore } from "../stores/use-notification-store";

export const notificationQueryKeys = {
  all: ["notifications"] as const,
  list: (params?: ListNotificationsParams) => ["notifications", params] as const,
  unreadCount: ["notifications", "unread-count"] as const,
};

function decrementUnreadCount(current: number | undefined, by = 1): number {
  const safeCurrent = current ?? 0;
  return Math.max(0, safeCurrent - by);
}

export function useNotifications(params?: ListNotificationsParams) {
  const realtimeNotifications = useNotificationStore((state) => state.realtimeNotifications);
  const query = useQuery({
    queryKey: notificationQueryKeys.list(params),
    queryFn: () => notificationService.list(params),
    retry: (failureCount, error) => {
      if (error && typeof error === "object" && "response" in error) {
        const axiosError = error as { response?: { status?: number } };
        if (axiosError.response?.status === 404) {
          return false;
        }
      }
      return failureCount < 1;
    },
  });

  const mergedData = useMemo<ListNotificationsResponse | undefined>(() => {
    if (!query.data) {
      if ((params?.page ?? 1) !== 1 || realtimeNotifications.length === 0) {
        return query.data;
      }

      const filteredRealtime = realtimeNotifications.filter((notification) => {
        if (params?.is_read !== undefined && notification.is_read !== params.is_read) {
          return false;
        }
        if (params?.type && notification.type !== params.type) {
          return false;
        }
        if (params?.entity && notification.entity_type !== params.entity) {
          return false;
        }
        return true;
      });

      if (filteredRealtime.length === 0) {
        return query.data;
      }

      return {
        success: true,
        data: filteredRealtime,
        meta: {
          pagination: {
            page: 1,
            per_page: params?.per_page ?? 20,
            total: filteredRealtime.length,
            total_pages: 1,
            has_next: false,
            has_prev: false,
          },
        },
        timestamp: new Date().toISOString(),
        request_id: "realtime-fallback",
      };
    }

    if ((params?.page ?? 1) !== 1) {
      return query.data;
    }

    if (realtimeNotifications.length === 0) {
      return query.data;
    }

    const filters = query.data.meta?.filters;
    const filteredRealtime = realtimeNotifications.filter((notification) => {
      if (params?.is_read !== undefined && notification.is_read !== params.is_read) {
        return false;
      }
      if (params?.type && notification.type !== params.type) {
        return false;
      }
      if (params?.entity && notification.entity_type !== params.entity) {
        return false;
      }

      if (filters && typeof filters === "object") {
        const f = filters as { type?: string; entity?: string; is_read?: boolean };
        if (f.type && notification.type !== f.type) return false;
        if (f.entity && notification.entity_type !== f.entity) return false;
        if (typeof f.is_read === "boolean" && notification.is_read !== f.is_read) return false;
      }

      return true;
    });

    if (filteredRealtime.length === 0) {
      return query.data;
    }

    const dedupe = new Map<string, Notification>();
    for (const notification of [...filteredRealtime, ...query.data.data]) {
      if (!dedupe.has(notification.id)) {
        dedupe.set(notification.id, notification);
      }
    }

    const mergedItems = Array.from(dedupe.values()).sort(
      (a, b) => Date.parse(b.created_at) - Date.parse(a.created_at)
    );

    return {
      ...query.data,
      data: mergedItems,
      meta: {
        ...query.data.meta,
        pagination: query.data.meta?.pagination
          ? {
              ...query.data.meta.pagination,
              total: query.data.meta.pagination.total + filteredRealtime.length,
            }
          : query.data.meta?.pagination,
      },
    };
  }, [params, query.data, realtimeNotifications]);

  return {
    ...query,
    data: mergedData,
  };
}

export function useNotificationCount() {
  const query = useQuery({
    queryKey: notificationQueryKeys.unreadCount,
    queryFn: () => notificationService.getUnreadCount(),
    staleTime: 5 * 60_000,
    gcTime: 30 * 60_000,
    refetchOnMount: false,
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
  });

  return {
    ...query,
    data: query.data ?? 0,
  };
}

export function useMarkAsRead() {
  const queryClient = useQueryClient();
  const markRealtimeAsRead = useNotificationStore((state) => state.markRealtimeAsRead);
  const realtimeNotifications = useNotificationStore((state) => state.realtimeNotifications);

  return useMutation({
    mutationFn: (id: string) => {
      if (isRealtimeNotificationId(id)) {
        const nowIso = new Date().toISOString();
        const localNotification = realtimeNotifications.find((item) => item.id === id);
        markRealtimeAsRead(id);

        const fallbackNotification: Notification = {
          id,
          user_id: "realtime",
          title: "Realtime notification",
          message: "Realtime notification",
          type: "info",
          entity_type: "system",
          entity_id: "",
          is_read: true,
          read_at: nowIso,
          created_at: nowIso,
        };

        return Promise.resolve({
          success: true,
          data: {
            ...(localNotification ?? fallbackNotification),
            is_read: true,
            read_at: nowIso,
          },
          timestamp: nowIso,
          request_id: "realtime",
        });
      }

      return notificationService.markAsRead(id);
    },
    onSuccess: (response, id) => {
      markRealtimeAsRead(id);

      if (isRealtimeNotificationId(id)) {
        return;
      }

      let hasUnreadTransition = false;

      queryClient.setQueriesData({ queryKey: notificationQueryKeys.all }, (oldData: unknown) => {
        if (!oldData || typeof oldData !== "object" || !("data" in oldData)) {
          return oldData;
        }

        const currentData = oldData as {
          data?: Array<{ id: string; is_read: boolean; read_at?: string | null }>;
        };
        if (!Array.isArray(currentData.data)) {
          return oldData;
        }

        const nextItems = currentData.data.map((item) => {
          if (item.id !== id || item.is_read) {
            return item;
          }

          hasUnreadTransition = true;
          return {
            ...item,
            is_read: true,
            read_at: response.data.read_at,
          };
        });

        return {
          ...currentData,
          data: nextItems,
        };
      });

      if (hasUnreadTransition) {
        queryClient.setQueryData<number>(notificationQueryKeys.unreadCount, (current) =>
          decrementUnreadCount(current)
        );
      } else {
        queryClient.invalidateQueries({ queryKey: notificationQueryKeys.unreadCount });
      }
    },
    onError: (error: unknown) => {
      const message =
        error instanceof Error ? error.message : "Failed to mark notification as read";
      toast.error(message);
    },
  });
}

export function useMarkAllAsRead() {
  const queryClient = useQueryClient();
  const markAllRealtimeAsRead = useNotificationStore((state) => state.markAllRealtimeAsRead);

  return useMutation({
    mutationFn: () => notificationService.markAllAsRead(),
    onSuccess: () => {
      markAllRealtimeAsRead();

      queryClient.setQueriesData({ queryKey: notificationQueryKeys.all }, (oldData: unknown) => {
        if (!oldData || typeof oldData !== "object" || !("data" in oldData)) {
          return oldData;
        }

        const currentData = oldData as {
          data?: Array<{ is_read: boolean; read_at?: string | null }>;
        };
        if (!Array.isArray(currentData.data)) {
          return oldData;
        }

        const nowIso = new Date().toISOString();
        return {
          ...currentData,
          data: currentData.data.map((item) =>
            item.is_read
              ? item
              : {
                  ...item,
                  is_read: true,
                  read_at: item.read_at ?? nowIso,
                }
          ),
        };
      });

      queryClient.setQueryData<number>(notificationQueryKeys.unreadCount, 0);
      queryClient.invalidateQueries({
        predicate: (query) =>
          query.queryKey[0] === "notifications" && query.queryKey[1] !== "unread-count",
      });
      toast.success("All notifications marked as read");
    },
    onError: (error: unknown) => {
      const message =
        error instanceof Error ? error.message : "Failed to mark all notifications as read";
      toast.error(message);
    },
  });
}

