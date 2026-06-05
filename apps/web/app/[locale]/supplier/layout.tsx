"use client";

import React from "react";
import { usePathname } from "@/i18n/routing";
import SupplierLayoutComponent from "@/features/supplier/layout/components/supplier-layout";

export default function SupplierLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const pathname = usePathname();

  const portalKeywords = [
    "/supplier/dashboard",
    "/supplier/profile",
    "/supplier/billing",
    "/supplier/subscription",
    "/supplier/verification",
    "/supplier/rfq",
    "/supplier/auction",
    "/supplier/notifications",
    "/supplier/onboarding",
    "/supplier/support",
    "/supplier/ads",
    "/supplier/reviews"
  ];

  const isPortal = portalKeywords.some(keyword => 
    pathname === keyword || pathname.startsWith(keyword + "/")
  );

  if (!isPortal) {
    return <>{children}</>;
  }

  return <SupplierLayoutComponent>{children}</SupplierLayoutComponent>;
}
