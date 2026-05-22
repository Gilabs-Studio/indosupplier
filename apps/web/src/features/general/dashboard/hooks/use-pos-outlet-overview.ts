"use client";

import { startTransition, useEffect, useMemo, useRef, useState } from "react";
import type { WidgetType, POSOutletItem, POSSummaryData } from "../types";
import { useDashboardWidgetOverview } from "./use-dashboard";

export function usePosOutletOverview(
  widgetType: WidgetType,
  enabled: boolean,
  options?: { allowAllOutlets?: boolean },
) {
  const allowAllOutlets = options?.allowAllOutlets ?? true;
  const [selectedOutletId, setSelectedOutletId] = useState<string>(allowAllOutlets ? "all" : "");
  const didSetInitialOutlet = useRef(false);

  const overviewQuery = useDashboardWidgetOverview(widgetType, {
    enabled,
  });

  const selectedOutletQuery = useDashboardWidgetOverview(widgetType, {
    enabled: enabled && selectedOutletId !== "all" && selectedOutletId !== "",
    outletId: selectedOutletId === "all" ? undefined : selectedOutletId,
  });

  const overviewSummary = overviewQuery.data?.pos_summary;
  const selectedSummary = selectedOutletQuery.data?.pos_summary;

  const visibleOutlets = useMemo<POSOutletItem[]>(() => overviewSummary?.outlets ?? [], [overviewSummary?.outlets]);

  useEffect(() => {
    if (!enabled || didSetInitialOutlet.current || visibleOutlets.length === 0) {
      return;
    }

    if (allowAllOutlets && selectedOutletId !== "all") {
      return;
    }

    didSetInitialOutlet.current = true;
    startTransition(() => {
      setSelectedOutletId(allowAllOutlets ? "all" : visibleOutlets[0].outlet_id);
    });
  }, [allowAllOutlets, enabled, selectedOutletId, visibleOutlets]);

  useEffect(() => {
    if (allowAllOutlets && selectedOutletId === "all") {
      return;
    }

    const hasSelectedOutlet = visibleOutlets.some((outlet) => outlet.outlet_id === selectedOutletId);
    if (!hasSelectedOutlet) {
      startTransition(() => {
        setSelectedOutletId(allowAllOutlets ? "all" : visibleOutlets[0]?.outlet_id ?? "");
      });
    }
  }, [allowAllOutlets, selectedOutletId, visibleOutlets]);

  const filteredOutlets = useMemo(() => {
    if (allowAllOutlets && selectedOutletId === "all") {
      return visibleOutlets;
    }

    return visibleOutlets.filter((outlet) => outlet.outlet_id === selectedOutletId);
  }, [allowAllOutlets, visibleOutlets, selectedOutletId]);

  const posSummary: POSSummaryData | undefined =
    allowAllOutlets && selectedOutletId === "all" ? overviewSummary : selectedSummary ?? overviewSummary;

  const focusedOutlet = useMemo(() => {
    if (!allowAllOutlets || selectedOutletId !== "all") {
      return visibleOutlets.find((outlet) => outlet.outlet_id === selectedOutletId);
    }

    if (visibleOutlets.length === 1) {
      return visibleOutlets[0];
    }

    return undefined;
  }, [allowAllOutlets, selectedOutletId, visibleOutlets]);

  const isLoading = overviewQuery.isLoading || (enabled && (!allowAllOutlets || selectedOutletId !== "all") && selectedOutletQuery.isLoading);
  const isError = overviewQuery.isError || selectedOutletQuery.isError;

  const refetch = async () => {
    await overviewQuery.refetch();
    if (!allowAllOutlets || selectedOutletId !== "all") {
      await selectedOutletQuery.refetch();
    }
  };

  return {
    posSummary,
    visibleOutlets,
    filteredOutlets,
    selectedOutletId,
    setSelectedOutletId,
    hasOutletFilter: visibleOutlets.length > 1,
    allowAllOutlets,
    isSingleOutletScope: visibleOutlets.length === 1,
    focusedOutlet,
    isLoading,
    isError,
    refetch,
  };
}