"use client";

import React from "react";
import { useTranslations } from "next-intl";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { TransactionItem } from "../types/transaction.types";

interface TransactionStatusBadgeProps {
  status: TransactionItem["status"] | string;
  paymentStatus?: TransactionItem["payment_status"] | string;
  className?: string;
}

export function TransactionStatusBadge({
  status,
  paymentStatus,
  className,
}: TransactionStatusBadgeProps) {
  const t = useTranslations("buyer.transactions");

  // Single unified status badge: completed order is shown as 1 unified badge (not 2 separate badges)
  if (status === "completed") {
    return (
      <Badge
        variant="success"
        className={cn("px-2.5 py-0.5 text-[11px] font-semibold tracking-wide border border-success/30 shadow-none", className)}
      >
        {t("statusCompleted")}
      </Badge>
    );
  }

  if (status === "cancelled") {
    return (
      <Badge
        variant="destructive"
        className={cn("px-2.5 py-0.5 text-[11px] font-semibold tracking-wide border border-destructive/30 shadow-none", className)}
      >
        {t("statusCancelled")}
      </Badge>
    );
  }

  if (status === "shipped") {
    return (
      <Badge
        variant="info"
        className={cn("px-2.5 py-0.5 text-[11px] font-semibold tracking-wide border border-primary/30 shadow-none", className)}
      >
        {t("statusShipped")}
      </Badge>
    );
  }

  if (status === "processing") {
    return (
      <Badge
        variant="warning"
        className={cn("px-2.5 py-0.5 text-[11px] font-semibold tracking-wide border border-warning/30 shadow-none", className)}
      >
        {t("statusProcessing")}
      </Badge>
    );
  }

  // Pending status
  if (paymentStatus === "unpaid") {
    return (
      <Badge
        variant="outline"
        className={cn("px-2.5 py-0.5 text-[11px] font-semibold tracking-wide border-warning/40 bg-warning/10 text-warning shadow-none", className)}
      >
        {t("statusPending")}
      </Badge>
    );
  }

  return (
    <Badge
      variant="outline"
      className={cn("px-2.5 py-0.5 text-[11px] font-semibold tracking-wide border-border bg-muted/30 text-muted-foreground shadow-none", className)}
    >
      {t("statusPending")}
    </Badge>
  );
}
