"use client";

import React from "react";
import { useDemoQuickActions } from "../../hooks/use-demo-home";
import { QuickActionsBar } from "../quick-actions-bar";
import { QuickActionsSkeleton } from "../skeletons/quick-actions-skeleton";

interface QuickActionsSectionProps {
  locale: string;
}

export function QuickActionsSection({ locale }: Readonly<QuickActionsSectionProps>) {
  const { quickActions, isLoading, isEn } = useDemoQuickActions(locale);

  if (isLoading) {
    return <QuickActionsSkeleton />;
  }

  return <QuickActionsBar actions={quickActions} isEn={isEn} />;
}
