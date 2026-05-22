import { create } from "zustand";
import type { WidgetConfig, WidgetColSpan, WidgetRowSpan, DashboardDateFilter } from "../types";
import { DEFAULT_WIDGETS } from "../config/widget-registry";

interface DashboardState {
  widgets: WidgetConfig[];
  isEditMode: boolean;
  /**
   * Snapshot of widgets captured when entering edit mode.
   * Used by resetLayout to restore the layout to its pre-edit state
   * without touching the DB (the user stays in edit mode after reset).
   */
  preEditSnapshot: WidgetConfig[] | null;
  dateFilter: DashboardDateFilter;
  // Actions
  setWidgets: (widgets: WidgetConfig[]) => void;
  reorderWidgets: (activeId: string, overId: string) => void;
  toggleWidgetVisibility: (widgetId: string) => void;
  addWidget: (widget: WidgetConfig) => void;
  removeWidget: (widgetId: string) => void;
  resizeWidget: (widgetId: string, size: WidgetConfig["size"]) => void;
  resizeWidgetCol: (widgetId: string, colSpan: WidgetColSpan) => void;
  resizeWidgetRow: (widgetId: string, rowSpan: WidgetRowSpan) => void;
  resetLayout: () => void;
  toggleEditMode: () => void;
  setDateFilter: (filter: Partial<DashboardDateFilter>) => void;
}

export const useDashboardStore = create<DashboardState>((set, get) => ({
  // Start with an empty grid; the dashboard-grid component hydrates from the saved layout
  // (or DEFAULT_WIDGETS once onboarding is complete and no layout has been saved yet).
  widgets: [],
  isEditMode: false,
  preEditSnapshot: null,
  dateFilter: {
    from: null,
    to: null,
    year: new Date().getFullYear(),
  },

  setWidgets: (widgets) => {
    set({ widgets });
  },

  reorderWidgets: (activeId, overId) => {
    const { widgets } = get();
    const oldIndex = widgets.findIndex((w) => w.id === activeId);
    const newIndex = widgets.findIndex((w) => w.id === overId);
    if (oldIndex === -1 || newIndex === -1) return;

    const updated = [...widgets];
    const [moved] = updated.splice(oldIndex, 1);
    updated.splice(newIndex, 0, moved);
    const reordered = updated.map((w, i) => ({ ...w, order: i }));
    set({ widgets: reordered });
  },

  toggleWidgetVisibility: (widgetId) => {
    const { widgets } = get();
    const updated = widgets.map((w) =>
      w.id === widgetId ? { ...w, visible: !w.visible } : w,
    );
    set({ widgets: updated });
  },

  addWidget: (widget) => {
    const { widgets } = get();
    const maxOrder = Math.max(...widgets.map((w) => w.order), -1);
    const newWidget = { ...widget, order: maxOrder + 1 };
    const updated = [...widgets, newWidget];
    set({ widgets: updated });
  },

  removeWidget: (widgetId) => {
    const { widgets } = get();
    const updated = widgets
      .filter((w) => w.id !== widgetId)
      .map((w, i) => ({ ...w, order: i }));
    set({ widgets: updated });
  },

  resizeWidget: (widgetId, size) => {
    const { widgets } = get();
    const updated = widgets.map((w) =>
      w.id === widgetId ? { ...w, size } : w,
    );
    set({ widgets: updated });
  },

  resizeWidgetCol: (widgetId, colSpan) => {
    const { widgets } = get();
    const updated = widgets.map((w) =>
      w.id === widgetId ? { ...w, colSpan } : w,
    );
    set({ widgets: updated });
  },

  resizeWidgetRow: (widgetId, rowSpan) => {
    const { widgets } = get();
    const updated = widgets.map((w) =>
      w.id === widgetId ? { ...w, rowSpan } : w,
    );
    set({ widgets: updated });
  },

  resetLayout: () => {
    const { preEditSnapshot } = get();
    // Restore to the snapshot taken before entering edit mode, stay in edit mode
    set({ widgets: preEditSnapshot ?? DEFAULT_WIDGETS });
  },

  toggleEditMode: () => {
    const { isEditMode, widgets } = get();
    if (!isEditMode) {
      // Entering edit mode — capture a snapshot of the current layout
      set({ isEditMode: true, preEditSnapshot: [...widgets] });
    } else {
      // Leaving edit mode — clear snapshot
      set({ isEditMode: false, preEditSnapshot: null });
    }
  },

  setDateFilter: (filter) => {
    set((s) => ({ dateFilter: { ...s.dateFilter, ...filter } }));
  },
}));
