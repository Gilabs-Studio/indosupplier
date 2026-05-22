"use client";

import { useCallback, useEffect, useMemo } from "react";
import { useTranslations } from "next-intl";
import { LayoutDashboard, RotateCcw, Check } from "lucide-react";
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from "@dnd-kit/core";
import {
  SortableContext,
  sortableKeyboardCoordinates,
  rectSortingStrategy,
} from "@dnd-kit/sortable";
import { Button } from "@/components/ui/button";
import { DateRangePicker } from "@/components/ui/date-range-picker";
import type { DateRange } from "react-day-picker";
import { useDashboardLayout, useSaveLayout } from "../hooks/use-dashboard";
import { useDashboardStore } from "../stores/useDashboardStore";
import { SortableWidget } from "./sortable-widget";
import { WidgetRenderer } from "./widget-renderer";
import { LazyMount } from "./lazy-mount";
import { WidgetPicker } from "./widget-picker";
import { OnboardingWizard } from "./onboarding-wizard";
import {
  buildAccessibleDefaultWidgets,
  collectAccessibleMenuUrls,
  isWidgetVisibleByAccessibleMenus,
  WIDGET_REGISTRY,
} from "../config/widget-registry";
import { useOnboardingState } from "../hooks/use-onboarding";
import type { WidgetType, WidgetConfig } from "../types";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import { useUserPermissions } from "@/features/master-data/user-management/hooks/use-user-permissions";

function normalizeDashboardWidgets(widgets: WidgetConfig[]): WidgetConfig[] {
  const normalized: WidgetConfig[] = [];

  for (const widget of widgets) {
    const widgetType = widget.type as string;
    if (widgetType !== "crm_deals_overview") {
      normalized.push(widget);
      continue;
    }

    const baseOrder = normalized.length;
    normalized.push({
      ...widget,
      id: `${widget.id}-leads`,
      type: "crm_leads_list",
      order: baseOrder,
      colSpan: 2,
      rowSpan: 2,
    });
    normalized.push({
      ...widget,
      id: `${widget.id}-pipeline`,
      type: "crm_pipeline_summary",
      order: baseOrder + 1,
      colSpan: 2,
      rowSpan: 2,
    });
  }

  return normalized.map((widget, index) => ({ ...widget, order: index }));
}

export function DashboardGrid() {
  const t = useTranslations("dashboard");
  const user = useAuthStore((state) => state.user);
  const isTenantOwner = user?.role?.is_owner === true;
  const { data: permissionsData, isLoading: isPermissionsLoading } = useUserPermissions();
  const { data: layoutData } = useDashboardLayout();
  const { data: onboarding, isLoading: isOnboardingLoading } = useOnboardingState({
    enabled: isTenantOwner,
  });
  const { mutate: saveLayout, isPending: isSaving } = useSaveLayout();

  const {
    widgets,
    isEditMode,
    dateFilter,
    setWidgets,
    reorderWidgets,
    removeWidget,
    resizeWidgetCol,
    resizeWidgetRow,
    addWidget,
    resetLayout,
    toggleEditMode,
    setDateFilter,
  } = useDashboardStore();

  const hasPersistedLayout = Boolean(layoutData && layoutData.length > 0);
  const accessibleMenuUrls = useMemo(
    () => collectAccessibleMenuUrls(permissionsData?.data.menus ?? []),
    [permissionsData?.data.menus],
  );
  const hasAccessibleMenus = accessibleMenuUrls.size > 0;

  // Sync layout from DB into store once loaded.
  // When a saved layout exists, hydrate it into the store.
  useEffect(() => {
    if (hasPersistedLayout && layoutData) {
      setWidgets(normalizeDashboardWidgets(layoutData));
    }
  }, [hasPersistedLayout, layoutData, setWidgets]);

  useEffect(() => {
    if (
      isOnboardingLoading ||
      onboarding === undefined ||
      !onboarding.completed ||
      isPermissionsLoading ||
      hasPersistedLayout ||
      widgets.length > 0 ||
      !hasAccessibleMenus
    ) {
      return;
    }

    const defaultWidgets = buildAccessibleDefaultWidgets(accessibleMenuUrls);
    if (defaultWidgets.length > 0) {
      setWidgets(defaultWidgets);
    }
  }, [
    accessibleMenuUrls,
    hasAccessibleMenus,
    hasPersistedLayout,
    isOnboardingLoading,
    isPermissionsLoading,
    onboarding,
    setWidgets,
    widgets.length,
  ]);

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    }),
  );

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      const { active, over } = event;
      if (over && active.id !== over.id) {
        reorderWidgets(String(active.id), String(over.id));
      }
    },
    [reorderWidgets],
  );

  const handleDateChange = useCallback(
    (range: DateRange | undefined) => {
      setDateFilter({
        from: range?.from?.toISOString()?.slice(0, 10) ?? null,
        to: range?.to?.toISOString()?.slice(0, 10) ?? null,
      });
    },
    [setDateFilter],
  );

  const handleSaveAndExit = useCallback(() => {
    saveLayout(widgets);
    toggleEditMode();
  }, [saveLayout, widgets, toggleEditMode]);

  const handleReset = useCallback(() => {
    // Restore pre-edit snapshot, stay in edit mode (no DB save)
    resetLayout();
  }, [resetLayout]);

  const handleAddWidget = useCallback(async (type: WidgetType) => {
    const registry = WIDGET_REGISTRY[type];
    if (!registry) return;
    addWidget({
      id: `w-${Date.now()}`,
      type,
      title: "",
      size: registry.defaultSize,
      colSpan: registry.defaultColSpan,
      rowSpan: registry.defaultRowSpan,
      order: widgets.length,
      visible: true,
    });
    // Return a resolved promise so callers can await if needed
    return Promise.resolve();
  }, [addWidget, widgets.length]);

  const dateRange: DateRange | undefined =
    dateFilter.from || dateFilter.to
      ? {
          from: dateFilter.from ? new Date(dateFilter.from) : undefined,
          to: dateFilter.to ? new Date(dateFilter.to) : undefined,
        }
      : undefined;

  const visibleWidgets = widgets
    .filter((w) => w.visible)
    .filter((w) => isWidgetVisibleByAccessibleMenus(w.type, accessibleMenuUrls))
    .sort((a, b) => a.order - b.order);

  // While onboarding is in progress or still loading, render the wizard as the full dashboard content.
  // Once onboarding is complete, render the grid.
  if (
    isTenantOwner &&
    (isOnboardingLoading || (onboarding !== undefined && !onboarding?.completed))
  ) {
    return <OnboardingWizard />;
  }

  return (
    <div className="space-y-4">
        {/* Header */}
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <h1 className="text-xl font-bold tracking-tight lg:text-2xl">
            {t("title")}
          </h1>
          <div className="flex flex-wrap items-center gap-2">
            {!isEditMode && (
              <DateRangePicker
                dateRange={dateRange}
                onDateChange={handleDateChange}
              />
            )}
            {isEditMode ? (
              <>
                <WidgetPicker
                  existingWidgets={widgets}
                  onAddWidget={handleAddWidget}
                />
                <Button
                  variant="outline"
                  size="sm"
                  className="cursor-pointer gap-1.5"
                  onClick={handleReset}
                >
                  <RotateCcw className="h-4 w-4" />
                  {t("resetLayout")}
                </Button>
                <Button
                  size="sm"
                  className="cursor-pointer gap-1.5"
                  onClick={handleSaveAndExit}
                  disabled={isSaving}
                >
                  <Check className="h-4 w-4" />
                  {isSaving ? t("loading") : t("doneEditing")}
                </Button>
              </>
            ) : (
              <>
                <Button
                  variant="outline"
                  size="sm"
                  className="cursor-pointer gap-1.5"
                  onClick={toggleEditMode}
                >
                  <LayoutDashboard className="h-4 w-4" />
                  {t("customize")}
                </Button>
              </>
            )}
          </div>
        </div>

        {/* Widget Grid — grid-flow-dense fills vertical gaps automatically (no empty cells) */}
        <DndContext
          sensors={sensors}
          collisionDetection={closestCenter}
          onDragEnd={handleDragEnd}
        >
          <SortableContext
            items={visibleWidgets.map((w) => w.id)}
            strategy={rectSortingStrategy}
          >
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4 grid-flow-dense">
              {visibleWidgets.map((widget) => (
                <SortableWidget
                  key={widget.id}
                  widget={widget}
                  isEditMode={isEditMode}
                  onRemove={removeWidget}
                  onResizeCol={resizeWidgetCol}
                  onResizeRow={resizeWidgetRow}
                >
                  <LazyMount>
                    <WidgetRenderer widget={widget} />
                  </LazyMount>
                </SortableWidget>
              ))}
            </div>
          </SortableContext>
        </DndContext>

        {visibleWidgets.length === 0 && (
          <div className="flex h-48 items-center justify-center rounded-lg border border-dashed">
            <p className="text-sm text-muted-foreground">{t("emptyDashboard")}</p>
          </div>
        )}
      </div>
  );
}
