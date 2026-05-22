"use client";

import type React from "react";

// Deprecated wrapper – sidebar is now controlled inside DashboardLayout.
// Kept only to avoid breaking any legacy imports.
export function SidebarWrapper({ children }: { children?: React.ReactNode }) {
  return <>{children}</>;
}
