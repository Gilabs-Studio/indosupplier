"use client";

import {
  createContext,
  useContext,
  useState,
  useEffect,
  useMemo,
  useCallback,
  type ReactNode,
} from "react";
import { useIsMobile } from "@/hooks/use-mobile";

interface SidebarContextValue {
  collapsed: boolean;
  setCollapsed: (collapsed: boolean) => void;
  toggleCollapsed: () => void;
  mobileOpen: boolean;
  setMobileOpen: (open: boolean) => void;
  toggleMobileOpen: () => void;
}

const SidebarContext = createContext<SidebarContextValue | undefined>(
  undefined,
);

const SIDEBAR_STORAGE_KEY = "sidebar-collapsed";

function getUserPreference(): boolean {
  try {
    const stored = globalThis.localStorage?.getItem(SIDEBAR_STORAGE_KEY);
    return stored === "true";
  } catch {
    return false;
  }
}

interface SidebarProviderProps {
  readonly children: ReactNode;
}

export function SidebarProvider({ children }: SidebarProviderProps) {
  const isMobile = useIsMobile();
  const [userPreference, setUserPreference] = useState(getUserPreference);
  const [mobileOpenState, setMobileOpenState] = useState(false);

  // On mobile, always collapse. On desktop, use user preference
  const collapsed = isMobile ? true : userPreference;

  // Mobile sidebar is always closed on desktop
  const mobileOpen = isMobile ? mobileOpenState : false;

  // Sync user preference to localStorage whenever it changes (only for desktop)
  useEffect(() => {
    if (!isMobile) {
      try {
        globalThis.localStorage.setItem(
          SIDEBAR_STORAGE_KEY,
          String(userPreference),
        );
      } catch {
        // Ignore localStorage errors (e.g., in private browsing)
      }
    }
  }, [userPreference, isMobile]);

  const toggleCollapsed = useCallback(() => {
    // Only allow toggle on desktop
    if (!isMobile) {
      setUserPreference((prev) => !prev);
    }
  }, [isMobile]);

  const handleSetCollapsed = useCallback(
    (value: boolean) => {
      // Only allow setting on desktop
      if (!isMobile) {
        setUserPreference(value);
      }
    },
    [isMobile],
  );

  const toggleMobileOpen = useCallback(() => {
    // Always allow toggle - the mobileOpen computed value will handle desktop case
    setMobileOpenState((prev) => !prev);
  }, []);

  const handleSetMobileOpen = useCallback((open: boolean) => {
    // Always allow setting - the mobileOpen computed value will handle desktop case
    setMobileOpenState(open);
  }, []);

  const contextValue = useMemo(
    () => ({
      collapsed,
      setCollapsed: handleSetCollapsed,
      toggleCollapsed,
      mobileOpen,
      setMobileOpen: handleSetMobileOpen,
      toggleMobileOpen,
    }),
    [
      collapsed,
      handleSetCollapsed,
      toggleCollapsed,
      mobileOpen,
      handleSetMobileOpen,
      toggleMobileOpen,
    ],
  );

  return (
    <SidebarContext.Provider value={contextValue}>
      {children}
    </SidebarContext.Provider>
  );
}

export function useSidebar() {
  const context = useContext(SidebarContext);
  if (!context) {
    throw new Error("useSidebar must be used within SidebarProvider");
  }
  return context;
}
