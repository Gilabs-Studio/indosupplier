"use client";

import { create } from "zustand";
import type { Notification } from "../types";

const REALTIME_NOTIFICATION_PREFIX = "realtime:";

export function isRealtimeNotificationId(id: string): boolean {
  return id.startsWith(REALTIME_NOTIFICATION_PREFIX);
}

export function buildRealtimeNotificationId(seed: string): string {
  return `${REALTIME_NOTIFICATION_PREFIX}${seed}`;
}

interface NotificationStore {
  isDrawerOpen: boolean;
  realtimeNotifications: Notification[];
  openDrawer: () => void;
  closeDrawer: () => void;
  toggleDrawer: () => void;
  addRealtimeNotification: (notification: Notification) => void;
  markRealtimeAsRead: (id: string) => void;
  markAllRealtimeAsRead: () => void;
}

export const useNotificationStore = create<NotificationStore>((set) => ({
  isDrawerOpen: false,
  realtimeNotifications: [],
  openDrawer: () => set({ isDrawerOpen: true }),
  closeDrawer: () => set({ isDrawerOpen: false }),
  toggleDrawer: () => set((state) => ({ isDrawerOpen: !state.isDrawerOpen })),
  addRealtimeNotification: (notification) =>
    set((state) => {
      const deduped = state.realtimeNotifications.filter((item) => item.id !== notification.id);
      return {
        realtimeNotifications: [notification, ...deduped].slice(0, 100),
      };
    }),
  markRealtimeAsRead: (id) =>
    set((state) => ({
      realtimeNotifications: state.realtimeNotifications.map((item) =>
        item.id === id
          ? {
              ...item,
              is_read: true,
              read_at: item.read_at ?? new Date().toISOString(),
            }
          : item
      ),
    })),
  markAllRealtimeAsRead: () =>
    set((state) => {
      const nowIso = new Date().toISOString();
      return {
        realtimeNotifications: state.realtimeNotifications.map((item) =>
          item.is_read
            ? item
            : {
                ...item,
                is_read: true,
                read_at: item.read_at ?? nowIso,
              }
        ),
      };
    }),
}));

