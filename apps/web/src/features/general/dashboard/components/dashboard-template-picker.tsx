"use client";

import { useTranslations } from "next-intl";
import { motion } from "framer-motion";
import { BarChart2, ShoppingCart, Store, Users2, Headset } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { WidgetConfig } from "../types";

// ─── Template definitions ────────────────────────────────────────────────────

const ERP_WIDGETS: WidgetConfig[] = [
  { id: "t-1", type: "total_orders",          title: "", size: "sm", colSpan: 1, rowSpan: 1, order: 0, visible: true },
  { id: "t-2", type: "total_revenue",         title: "", size: "sm", colSpan: 1, rowSpan: 1, order: 1, visible: true },
  { id: "t-3", type: "total_customers",       title: "", size: "sm", colSpan: 1, rowSpan: 1, order: 2, visible: true },
  { id: "t-4", type: "total_products",        title: "", size: "sm", colSpan: 1, rowSpan: 1, order: 3, visible: true },
  { id: "t-5", type: "revenue_bar_chart",     title: "", size: "md", colSpan: 2, rowSpan: 2, order: 4, visible: true },
  { id: "t-6", type: "track_orders",          title: "", size: "md", colSpan: 2, rowSpan: 2, order: 5, visible: true },
  { id: "t-7", type: "track_purchase_orders", title: "", size: "md", colSpan: 2, rowSpan: 2, order: 6, visible: true },
  { id: "t-8", type: "pending_approvals_sales",    title: "", size: "md", colSpan: 2, rowSpan: 2, order: 7, visible: true },
  { id: "t-9", type: "pending_approvals_purchase",  title: "", size: "md", colSpan: 2, rowSpan: 2, order: 8, visible: true },
];

const CRM_WIDGETS: WidgetConfig[] = [
  { id: "t-1", type: "crm_total_contacts",    title: "", size: "sm", colSpan: 1, rowSpan: 1, order: 0, visible: true },
  { id: "t-2", type: "crm_active_leads",      title: "", size: "sm", colSpan: 1, rowSpan: 1, order: 1, visible: true },
  { id: "t-3", type: "crm_leads_list",         title: "", size: "md", colSpan: 2, rowSpan: 2, order: 2, visible: true },
  { id: "t-4", type: "crm_pipeline_summary",   title: "", size: "md", colSpan: 2, rowSpan: 2, order: 3, visible: true },
  { id: "t-5", type: "crm_activity_summary",   title: "", size: "md", colSpan: 2, rowSpan: 1, order: 4, visible: true },
  { id: "t-6", type: "total_customers",        title: "", size: "sm", colSpan: 1, rowSpan: 1, order: 5, visible: true },
  { id: "t-7", type: "travel_planner_overview", title: "", size: "xl", colSpan: 4, rowSpan: 2, order: 6, visible: true },
];

const POS_WIDGETS: WidgetConfig[] = [
  { id: "t-1", type: "pos_outlet_sales",         title: "", size: "lg", colSpan: 3, rowSpan: 2, order: 0, visible: true },
  { id: "t-2", type: "total_revenue",            title: "", size: "sm", colSpan: 1, rowSpan: 1, order: 1, visible: true },
  { id: "t-3", type: "total_orders",             title: "", size: "sm", colSpan: 1, rowSpan: 1, order: 2, visible: true },
  { id: "t-4", type: "pos_live_table_overview",  title: "", size: "xl", colSpan: 4, rowSpan: 2, order: 3, visible: true },
  { id: "t-5", type: "pos_cash_control",         title: "", size: "lg", colSpan: 2, rowSpan: 2, order: 4, visible: true },
  { id: "t-6", type: "best_selling",             title: "", size: "md", colSpan: 2, rowSpan: 2, order: 5, visible: true },
];

const HR_WIDGETS: WidgetConfig[] = [
  { id: "t-1", type: "employee_count",         title: "", size: "sm", colSpan: 1, rowSpan: 1, order: 0, visible: true },
  { id: "t-2", type: "hr_pending_leaves",      title: "", size: "sm", colSpan: 1, rowSpan: 1, order: 1, visible: true },
  { id: "t-3", type: "hr_attendance_today",    title: "", size: "md", colSpan: 2, rowSpan: 1, order: 2, visible: true },
];

interface Template {
  id: string;
  nameKey: string;
  descriptionKey: string;
  icon: React.ElementType;
  widgets: WidgetConfig[];
  previewBlocks: Array<{ cols: number }>;
}

const TEMPLATES: Template[] = [
  {
    id: "erp",
    nameKey: "templatePicker.templates.erp.name",
    descriptionKey: "templatePicker.templates.erp.description",
    icon: ShoppingCart,
    widgets: ERP_WIDGETS,
    previewBlocks: [{ cols: 1 }, { cols: 1 }, { cols: 1 }, { cols: 1 }, { cols: 2 }, { cols: 2 }],
  },
  {
    id: "crm",
    nameKey: "templatePicker.templates.crm.name",
    descriptionKey: "templatePicker.templates.crm.description",
    icon: Headset,
    widgets: CRM_WIDGETS,
    previewBlocks: [{ cols: 1 }, { cols: 1 }, { cols: 2 }, { cols: 2 }, { cols: 1 }, { cols: 4 }],
  },
  {
    id: "pos",
    nameKey: "templatePicker.templates.pos.name",
    descriptionKey: "templatePicker.templates.pos.description",
    icon: Store,
    widgets: POS_WIDGETS,
    previewBlocks: [{ cols: 3 }, { cols: 1 }, { cols: 4 }, { cols: 2 }, { cols: 2 }],
  },
  {
    id: "hr",
    nameKey: "templatePicker.templates.hr.name",
    descriptionKey: "templatePicker.templates.hr.description",
    icon: Users2,
    widgets: HR_WIDGETS,
    previewBlocks: [{ cols: 1 }, { cols: 1 }, { cols: 2 }],
  },
];

// ─── Mini preview grid ───────────────────────────────────────────────────────

function TemplatePreview({ blocks }: { blocks: Array<{ cols: number }> }) {
  return (
    <div className="grid grid-cols-4 gap-1 rounded-md bg-muted/50 p-2">
      {blocks.map((block, i) => (
        <div
          key={i}
          className="h-4 rounded-sm bg-muted-foreground/20"
          style={{ gridColumn: `span ${block.cols}` }}
        />
      ))}
    </div>
  );
}

// ─── Main component ──────────────────────────────────────────────────────────

interface DashboardTemplatePickerProps {
  readonly onSelect: (widgets: WidgetConfig[]) => void;
}

export function DashboardTemplatePicker({ onSelect }: DashboardTemplatePickerProps) {
  const t = useTranslations("dashboard");

  return (
    <div className="flex min-h-[60vh] items-center justify-center px-4 py-12">
      <motion.div
        initial={{ opacity: 0, y: 16 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.35, ease: "easeOut" }}
        className="w-full max-w-3xl"
      >
        {/* Header */}
        <div className="mb-8 text-center">
          <div className="mb-3 flex items-center justify-center gap-2 text-muted-foreground">
            <BarChart2 className="h-5 w-5" />
          </div>
          <h2 className="text-2xl font-semibold tracking-tight">
            {t("templatePicker.title")}
          </h2>
          <p className="mt-1.5 text-sm text-muted-foreground">
            {t("templatePicker.description")}
          </p>
        </div>

        {/* Template cards */}
        <div className="grid gap-4 sm:grid-cols-2">
          {TEMPLATES.map((tpl, i) => {
            const Icon = tpl.icon;
            return (
              <motion.div
                key={tpl.id}
                initial={{ opacity: 0, y: 12 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.3, delay: i * 0.06 }}
                className="group flex flex-col gap-3 rounded-xl border bg-card p-4 hover:border-primary/50 hover:shadow-sm"
              >
                {/* Card header */}
                <div className="flex items-center gap-2.5">
                  <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-muted">
                    <Icon className="h-4 w-4 text-muted-foreground" />
                  </div>
                  <div>
                    <p className="text-sm font-medium leading-tight">
                      {t(tpl.nameKey as Parameters<typeof t>[0])}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {t(tpl.descriptionKey as Parameters<typeof t>[0])}
                    </p>
                  </div>
                </div>

                {/* Mini layout preview */}
                <TemplatePreview blocks={tpl.previewBlocks} />

                {/* Action */}
                <Button
                  size="sm"
                  variant="outline"
                  className="w-full cursor-pointer text-xs transition-colors group-hover:border-primary/50 group-hover:bg-primary group-hover:text-primary-foreground"
                  onClick={() => onSelect(tpl.widgets)}
                >
                  {t("templatePicker.useTemplate")}
                </Button>
              </motion.div>
            );
          })}
        </div>
      </motion.div>
    </div>
  );
}
