"use client";

import { useState, useMemo } from "react";
import { useTranslations } from "next-intl";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Plus,
  DollarSign,
  ShoppingCart,
  Users,
  Package,
  UserCheck,
  TrendingUp,
  TrendingDown,
  BarChart3,
  Wallet,
  PieChart,
  FileText,
  Receipt,
  Award,
  Star,
  Truck,
  Map,
  Warehouse,
  ClipboardCheck,
  PackageCheck,
  Route,
  Timer,
  CalendarClock,
  ArrowDownToLine,
  ArrowUpFromLine,
  Percent,
  Building2,
  RefreshCw,
  Gauge,
  Activity,
  Clock,
  Brain,
  ContactRound,
  UserPlus,
  Handshake,
  CalendarCheck,
  Store,
  LayoutGrid,
  CalendarCheck2,
  ClipboardList,
} from "lucide-react";
import { Loader2 } from "lucide-react";
import {
  collectAccessibleMenuUrls,
  getWidgetsByCategory,
  isWidgetVisibleByAccessibleMenus,
  WIDGET_REGISTRY,
} from "../config/widget-registry";
import type { WidgetType, WidgetConfig, WidgetCategory } from "../types";
import { useUserPermissions } from "@/features/master-data/user-management/hooks/use-user-permissions";

const ICON_MAP: Record<string, React.ElementType> = {
  DollarSign,
  ShoppingCart,
  Users,
  Package,
  UserCheck,
  TrendingUp,
  TrendingDown,
  BarChart3,
  Wallet,
  PieChart,
  FileText,
  Receipt,
  Award,
  Star,
  Truck,
  Map,
  Warehouse,
  ClipboardCheck,
  PackageCheck,
  Route,
  Timer,
  CalendarClock,
  ArrowDownToLine,
  ArrowUpFromLine,
  Percent,
  Building2,
  RefreshCw,
  Gauge,
  Activity,
  Clock,
  Brain,
  ContactRound,
  UserPlus,
  Handshake,
  CalendarCheck,
  Store,
  LayoutGrid,
  CalendarCheck2,
  ClipboardList,
};

const CATEGORY_LABEL_KEYS: Record<WidgetCategory, string> = {
  erp: "categories.erp",
  crm: "categories.crm",
  pos: "categories.pos",
  hr: "categories.hr",
  finance: "categories.finance",
  other: "categories.other",
};

interface WidgetPickerProps {
  readonly existingWidgets: WidgetConfig[];
  // May return a promise if adding involves async work
  readonly onAddWidget: (type: WidgetType) => void | Promise<void>;
}

function isPromiseLike(value: unknown): value is PromiseLike<unknown> {
  if (typeof value !== "object" || value === null) return false;
  const withThen = value as { then?: unknown };
  return typeof withThen.then === "function";
}

export function WidgetPicker({ existingWidgets, onAddWidget }: WidgetPickerProps) {
  const t = useTranslations("dashboard");
  const { data: permissionsData, isLoading: isPermissionsLoading } = useUserPermissions();
  const [open, setOpen] = useState(false);
  const [selectedCategory, setSelectedCategory] = useState<string>("all");
  const [addingSet, setAddingSet] = useState<Set<string>>(new Set());

  const grouped = useMemo(() => getWidgetsByCategory(), []);
  const existingTypes = useMemo(
    () => new Set(existingWidgets.map((w) => w.type)),
    [existingWidgets],
  );

  const categories = Object.keys(grouped);
  const filteredWidgets = useMemo(
    () =>
      selectedCategory === "all"
        ? Object.values(WIDGET_REGISTRY)
        : grouped[selectedCategory] ?? [],
    [selectedCategory, grouped],
  );

  const accessibleMenuUrls = useMemo(
    () => collectAccessibleMenuUrls(permissionsData?.data.menus ?? []),
    [permissionsData?.data.menus],
  );

  const visibleWidgets = useMemo(() => {
    if (isPermissionsLoading) {
      return [];
    }

    return filteredWidgets.filter((entry) =>
      isWidgetVisibleByAccessibleMenus(entry.type, accessibleMenuUrls),
    );
  }, [accessibleMenuUrls, filteredWidgets, isPermissionsLoading]);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm" className="cursor-pointer gap-1.5">
          <Plus className="h-4 w-4" />
          {t("addWidget")}
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>{t("widgetPicker.title")}</DialogTitle>
        </DialogHeader>

        {/* Category filter */}
        <div className="flex flex-wrap gap-1.5">
          <Badge
            variant={selectedCategory === "all" ? "default" : "outline"}
            className="cursor-pointer"
            onClick={() => setSelectedCategory("all")}
          >
            {t("widgetPicker.all")}
          </Badge>
          {categories.map((cat) => (
            <Badge
              key={cat}
              variant={selectedCategory === cat ? "default" : "outline"}
              className="cursor-pointer"
              onClick={() => setSelectedCategory(cat)}
            >
              {t(CATEGORY_LABEL_KEYS[cat as WidgetCategory] as Parameters<typeof t>[0]) ?? cat}
            </Badge>
          ))}
        </div>

        {/* Widget list */}
        <div className="max-h-80 space-y-1.5 overflow-y-auto">
          {visibleWidgets.map((entry) => {
            const Icon = ICON_MAP[entry.icon] ?? DollarSign;
            const exists = existingTypes.has(entry.type);
            const isAdding = addingSet.has(entry.type);

            return (
              <div
                key={entry.type}
                className={`flex items-center justify-between rounded-lg border p-3 transition-colors ${
                  exists || isAdding
                    ? "opacity-50 pointer-events-none"
                    : "cursor-pointer hover:bg-secondary"
                }`}
                onClick={async () => {
                  if (exists || isAdding) return;
                  // mark as adding
                  setAddingSet((prev) => {
                    const s = new Set(prev);
                    s.add(entry.type);
                    return s;
                  });

                  try {
                    const result = onAddWidget(entry.type);
                    if (isPromiseLike(result)) {
                      await result;
                    } else {
                      // Ensure UX shows a small progress even for sync adds
                      await new Promise((r) => setTimeout(r, 600));
                    }
                  } finally {
                    setAddingSet((prev) => {
                      const s = new Set(prev);
                      s.delete(entry.type);
                      return s;
                    });
                  }
                }}
                role="button"
                tabIndex={exists || isAdding ? -1 : 0}
                onKeyDown={async (e) => {
                  if (exists || isAdding || !(e.key === "Enter" || e.key === " ")) return;
                  // same flow as click
                  setAddingSet((prev) => {
                    const s = new Set(prev);
                    s.add(entry.type);
                    return s;
                  });
                  try {
                    const result = onAddWidget(entry.type);
                    if (isPromiseLike(result)) {
                      await result;
                    } else {
                      await new Promise((r) => setTimeout(r, 600));
                    }
                  } finally {
                    setAddingSet((prev) => {
                      const s = new Set(prev);
                      s.delete(entry.type);
                      return s;
                    });
                  }
                }}
              >
                <div className="flex items-center gap-3">
                  <div className="rounded-lg bg-primary/10 p-2">
                    <Icon className="h-4 w-4 text-primary" />
                  </div>
                  <div>
                    <p className="text-sm font-medium">
                      {t(entry.titleKey as Parameters<typeof t>[0])}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {t(entry.descriptionKey as Parameters<typeof t>[0])}
                    </p>
                  </div>
                </div>
                {exists ? (
                  <Badge variant="secondary">{t("widgetPicker.added")}</Badge>
                ) : addingSet.has(entry.type) ? (
                  <Loader2 className="h-4 w-4 text-muted-foreground animate-spin" />
                ) : (
                  <Plus className="h-4 w-4 text-muted-foreground" />
                )}
              </div>
            );
          })}
        </div>
      </DialogContent>
    </Dialog>
  );
}
