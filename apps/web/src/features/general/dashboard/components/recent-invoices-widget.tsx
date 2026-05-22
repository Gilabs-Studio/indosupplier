"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useHasPermission } from "@/features/master-data/user-management/hooks/use-has-permission";
import { InvoiceDetailModal } from "@/features/sales/invoice/components/invoice-detail-modal";
import type { CustomerInvoice } from "@/features/sales/invoice/types";
import { formatCurrency, formatDate } from "@/lib/utils";
import type { InvoiceRow } from "../types";

interface RecentInvoicesWidgetProps {
  readonly data?: InvoiceRow[];
}

type BadgeVariant = "success" | "warning" | "destructive";

const STATUS_VARIANT: Record<string, BadgeVariant> = {
  paid: "success",
  unpaid: "warning",
  overdue: "destructive",
};

export function RecentInvoicesWidget({ data }: RecentInvoicesWidgetProps) {
  const t = useTranslations("dashboard");
  const invoices = data ?? [];
  const [selectedInvoiceId, setSelectedInvoiceId] = useState<string | null>(null);
  const canViewInvoice = useHasPermission("customer_invoice.read");

  const formatIssueDate = (value?: string | null) => {
    if (!value) return "-";
    return formatDate(value) || value;
  };

  return (
    <>
      <Card className="h-full">
      <CardHeader className="pb-2">
        <CardTitle className="text-base font-semibold">
          {t("widgets.recent_invoices.title")}
        </CardTitle>
      </CardHeader>
      <CardContent>
        {invoices.length === 0 ? (
          <div className="flex h-36 items-center justify-center">
            <p className="text-sm text-muted-foreground">{t("noData")}</p>
          </div>
        ) : (
          <div className="space-y-2">
            {invoices.slice(0, 8).map((inv) => (
              <div
                key={inv.id}
                onClick={canViewInvoice ? () => setSelectedInvoiceId(inv.id) : undefined}
                role={canViewInvoice ? "button" : undefined}
                tabIndex={canViewInvoice ? 0 : undefined}
                onKeyDown={canViewInvoice ? (e) => { if (e.key === "Enter") setSelectedInvoiceId(inv.id); } : undefined}
                className={"flex items-center justify-between rounded-lg border p-3" + (canViewInvoice ? " cursor-pointer hover:bg-muted/50" : "")}
              >
                <div className="min-w-0 flex-1">
                  <p className={"truncate text-sm font-medium" + (canViewInvoice ? " text-primary" : "")}>{inv.company}</p>
                  <p className="text-xs text-muted-foreground tabular-nums">
                    {inv.contact} &middot; {formatIssueDate(inv.issue_date)}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-sm font-semibold">
                    {formatCurrency(inv.value)}
                  </span>
                  <Badge variant={STATUS_VARIANT[inv.status] ?? "warning"}>
                    {inv.status}
                  </Badge>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>

    {canViewInvoice && (
      <InvoiceDetailModal
        open={!!selectedInvoiceId}
        onClose={() => setSelectedInvoiceId(null)}
        invoice={selectedInvoiceId ? ({ id: selectedInvoiceId } as unknown as CustomerInvoice) : null}
      />
    )}
    </>
  );
}
