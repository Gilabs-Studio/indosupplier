"use client";

import React from "react";
import SysadminLayoutComponent from "@/features/sysadmin/dashboard/components/sysadmin-layout";

export default function SysadminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <SysadminLayoutComponent>{children}</SysadminLayoutComponent>;
}
