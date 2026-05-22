"use client";

import type React from "react";

interface AppLayoutProps {
  readonly children: React.ReactNode;
}

// AppLayout is a minimal wrapper that doesn't add dashboard chrome.
// DashboardLayout (sidebar + header) is only applied to routes inside the (dashboard) route group
// via app/[locale]/(dashboard)/layout.tsx
// Public routes (locale-scoped login page at "/[locale]/login") are rendered without any wrapper.
export function AppLayout({ children }: AppLayoutProps) {
  // Just render children - dashboard layout is handled by route group layout
  return <>{children}</>;
}
