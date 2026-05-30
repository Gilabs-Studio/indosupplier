"use client";

import { create } from "zustand";
import type { SystemAdmin } from "../types";
import { sysadminService } from "../services/sysadmin-service";

interface SysadminStore {
  admin: SystemAdmin | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  isSessionVerified: boolean;

  setAdmin: (admin: SystemAdmin | null) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  setSessionVerified: (verified: boolean) => void;
  logout: () => Promise<void>;
  checkSession: () => Promise<SystemAdmin | null>;
}

export const useSysadminStore = create<SysadminStore>()((set, get) => ({
  admin: null,
  isAuthenticated: false,
  isLoading: false,
  error: null,
  isSessionVerified: false,

  setAdmin: (admin: SystemAdmin | null) => {
    set({
      admin,
      isAuthenticated: !!admin,
      isSessionVerified: !!admin,
    });
  },

  setLoading: (loading: boolean) => {
    set({ isLoading: loading });
  },

  setError: (error: string | null) => {
    set({ error });
  },

  setSessionVerified: (verified: boolean) => {
    set({ isSessionVerified: verified });
  },

  logout: async () => {
    set({ isLoading: true });
    try {
      await sysadminService.logout();
    } catch (err) {
      console.error("Logout failed on server:", err);
    } finally {
      set({
        admin: null,
        isAuthenticated: false,
        isSessionVerified: false,
        isLoading: false,
        error: null,
      });
    }
  },

  checkSession: async () => {
    const state = get();
    if (state.isSessionVerified) {
      return state.admin;
    }

    set({ isLoading: true, error: null });
    try {
      const admin = await sysadminService.getMe();
      set({
        admin,
        isAuthenticated: true,
        isSessionVerified: true,
        isLoading: false,
      });
      return admin;
    } catch (err: any) {
      set({
        admin: null,
        isAuthenticated: false,
        isSessionVerified: true,
        isLoading: false,
      });
      return null;
    }
  },
}));
