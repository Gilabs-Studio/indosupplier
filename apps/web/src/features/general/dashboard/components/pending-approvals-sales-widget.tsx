"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { Clock, FileText, Receipt, ShoppingCart, Truck } from "lucide-react";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { useHasPermission } from "@/features/master-data/user-management/hooks/use-has-permission";
import { useQuotations } from "@/features/sales/quotation/hooks/use-quotations";
import { useOrders } from "@/features/sales/order/hooks/use-orders";
import { useDeliveryOrders } from "@/features/sales/delivery/hooks/use-deliveries";
import { useInvoices } from "@/features/sales/invoice/hooks/use-invoices";
import { QuotationDetailModal } from "@/features/sales/quotation/components/quotation-detail-modal";
import { OrderDetailModal } from "@/features/sales/order/components/order-detail-modal";
import { DeliveryDetailModal } from "@/features/sales/delivery/components/delivery-detail-modal";
import { InvoiceDetailModal } from "@/features/sales/invoice/components/invoice-detail-modal";
import type { SalesQuotation } from "@/features/sales/quotation/types";
import type { SalesOrder } from "@/features/sales/order/types";
import type { DeliveryOrder } from "@/features/sales/delivery/types";
import type { CustomerInvoice } from "@/features/sales/invoice/types";
import { formatCurrency, formatDate as formatDateUtil } from "@/lib/utils";

const PER_PAGE = 8;
type SalesTab = "quotation" | "order" | "delivery" | "invoice";

function formatDate(dateStr: string) {
  return formatDateUtil(dateStr);
}

interface RowSkeletonProps {
  count?: number;
}

function RowSkeleton({ count = 3 }: RowSkeletonProps) {
  return (
    <div className="space-y-3">
      {Array.from({ length: count }).map((_, i) => (
        <Skeleton key={i} className="h-14 w-full rounded-md bg-muted" />
      ))}
    </div>
  );
}

interface EmptyStateProps {
  label: string;
}

function EmptyState({ label }: EmptyStateProps) {
  return (
    <div className="flex h-36 flex-col items-center justify-center text-muted-foreground">
      <Clock className="mb-2 h-5 w-5 opacity-40" />
      <p className="text-sm">{label}</p>
    </div>
  );
}

interface ApprovalRowProps {
  code: string;
  date: string;
  party?: string;
  amount?: string;
  onClick: () => void;
}

function ApprovalRow({ code, date, party, amount, onClick }: ApprovalRowProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="hover:bg-muted flex w-full cursor-pointer items-center justify-between rounded-md border px-4 py-3 text-left transition-colors"
    >
      <div className="min-w-0">
        <p className="truncate text-sm font-medium text-foreground">{code}</p>
        <p className="truncate text-xs text-muted-foreground">
          {date}
          {party ? ` · ${party}` : ""}
        </p>
      </div>
      {amount && (
        <span className="ml-2 shrink-0 text-sm font-semibold text-foreground">
          {amount}
        </span>
      )}
    </button>
  );
}

export function PendingApprovalsSalesWidget() {
  const t = useTranslations("dashboard");
  const tCommon = useTranslations("common");

  // Permissions — always call hooks unconditionally
  const canViewQuotation = useHasPermission("sales_quotation.read");
  const canViewOrder = useHasPermission("sales_order.read");
  const canViewDelivery = useHasPermission("delivery_order.read");
  const canViewInvoice = useHasPermission("customer_invoice.read");
  // approval permission hooks removed (not used after action container removal)

  const defaultTab =
    canViewQuotation
      ? "quotation"
      : canViewOrder
        ? "order"
        : canViewDelivery
          ? "delivery"
          : "invoice";

  const [activeTab, setActiveTab] = useState<SalesTab>(defaultTab);

  // Detail modal state
  const [selectedQuotation, setSelectedQuotation] =
    useState<SalesQuotation | null>(null);
  const [selectedOrder, setSelectedOrder] = useState<SalesOrder | null>(null);
  const [selectedDelivery, setSelectedDelivery] =
    useState<DeliveryOrder | null>(null);
  const [selectedInvoice, setSelectedInvoice] =
    useState<CustomerInvoice | null>(null);

  // Queries — hooks must be called unconditionally; use `enabled` where supported
  const {
    data: qData,
    isLoading: qLoading,
    isError: qError,
    refetch: qRefetch,
  } = useQuotations(
    { status: "sent", per_page: PER_PAGE },
    { enabled: canViewQuotation },
  );
  const {
    data: oData,
    isLoading: oLoading,
    isError: oError,
    refetch: oRefetch,
  } = useOrders(
    { status: "submitted", per_page: PER_PAGE },
    { enabled: canViewOrder },
  );
  const {
    data: dData,
    isLoading: dLoading,
    isError: dError,
    refetch: dRefetch,
  } = useDeliveryOrders(
    {
      status: "sent",
      per_page: PER_PAGE,
    },
    { enabled: canViewDelivery },
  );
  const {
    data: iData,
    isLoading: iLoading,
    isError: iError,
    refetch: iRefetch,
  } = useInvoices(
    {
      status: "submitted",
      per_page: PER_PAGE,
    },
    { enabled: canViewInvoice },
  );

  const quotations = qData?.data ?? [];
  const orders = oData?.data ?? [];
  const deliveries = dData?.data ?? [];
  const invoices = iData?.data ?? [];

  const hasAnyPermission =
    canViewQuotation || canViewOrder || canViewDelivery || canViewInvoice;

  if (!hasAnyPermission) return null;

  return (
    <>
      <Card className="flex h-full flex-col gap-0">
        <CardHeader>
          <div className="flex items-start justify-between">
            <div>
              <CardTitle>{t("widgets.pending_approvals_sales.title")}</CardTitle>
              <CardDescription className="mt-1">
                {t("widgets.pending_approvals_sales.description")}
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
            <Tabs
            value={activeTab}
            onValueChange={(value) => {
              if (
                value === "quotation" ||
                value === "order" ||
                value === "delivery" ||
                value === "invoice"
              ) {
                setActiveTab(value);
              }
            }}
            className="pt-0 flex flex-1 flex-col"
          >
            <TabsList className="mb-1 w-full justify-start gap-1 overflow-x-auto border-b border-border/70">
              {canViewQuotation && (
                <TabsTrigger value="quotation" className="gap-1 px-1.5 text-xs">
                  <FileText className="h-3 w-3 shrink-0" />
                  {t("approvals.tabs.quotations")}
                  {quotations.length > 0 && (
                    <Badge variant="secondary" className="ml-0.5">
                      {quotations.length}
                    </Badge>
                  )}
                </TabsTrigger>
              )}
              {canViewOrder && (
                <TabsTrigger value="order" className="gap-1 px-1.5 text-xs">
                  <ShoppingCart className="h-3 w-3 shrink-0" />
                  {t("approvals.tabs.salesOrders")}
                  {orders.length > 0 && (
                    <Badge variant="secondary" className="ml-0.5">
                      {orders.length}
                    </Badge>
                  )}
                </TabsTrigger>
              )}
              {canViewDelivery && (
                <TabsTrigger value="delivery" className="gap-1 px-1.5 text-xs">
                  <Truck className="h-3 w-3 shrink-0" />
                  {t("approvals.tabs.deliveryOrders")}
                  {deliveries.length > 0 && (
                    <Badge variant="secondary" className="ml-0.5">
                      {deliveries.length}
                    </Badge>
                  )}
                </TabsTrigger>
              )}
              {canViewInvoice && (
                <TabsTrigger value="invoice" className="gap-1 px-1.5 text-xs">
                  <Receipt className="h-3 w-3 shrink-0" />
                  {t("approvals.tabs.customerInvoices")}
                  {invoices.length > 0 && (
                    <Badge variant="secondary" className="ml-0.5">
                      {invoices.length}
                    </Badge>
                  )}
                </TabsTrigger>
              )}
            </TabsList>
            {/* action container removed as requested */}

            {/* Quotations */}
            <TabsContent value="quotation" className="mt-0 flex-1">
              {qLoading ? (
                <RowSkeleton />
              ) : qError ? (
                <div className="flex h-36 flex-col items-center justify-center gap-2 text-center">
                  <p className="text-sm text-muted-foreground">{t("error")}</p>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className="cursor-pointer"
                    onClick={() => {
                      void qRefetch();
                    }}
                  >
                    {tCommon("retry")}
                  </Button>
                </div>
              ) : quotations.length === 0 ? (
                <EmptyState label={t("approvals.empty")} />
              ) : (
                <div className="space-y-2">
                  {quotations.map((q) => (
                    <ApprovalRow
                      key={q.id}
                      code={q.code}
                      date={formatDate(q.quotation_date)}
                      party={q.customer?.name}
                      amount={formatCurrency(q.total_amount)}
                      onClick={() => setSelectedQuotation(q)}
                    />
                  ))}
                </div>
              )}
            </TabsContent>

            {/* Sales Orders */}
            <TabsContent value="order" className="mt-0 flex-1">
              {oLoading ? (
                <RowSkeleton />
              ) : oError ? (
                <div className="flex h-36 flex-col items-center justify-center gap-2 text-center">
                  <p className="text-sm text-muted-foreground">{t("error")}</p>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className="cursor-pointer"
                    onClick={() => {
                      void oRefetch();
                    }}
                  >
                    {tCommon("retry")}
                  </Button>
                </div>
              ) : orders.length === 0 ? (
                <EmptyState label={t("approvals.empty")} />
              ) : (
                <div className="space-y-2 mt-2">
                  {orders.map((o) => (
                    <ApprovalRow
                      key={o.id}
                      code={o.code}
                      date={formatDate(o.order_date)}
                      party={o.customer?.name}
                      amount={formatCurrency(o.total_amount)}
                      onClick={() => setSelectedOrder(o)}
                    />
                  ))}
                </div>
              )}
            </TabsContent>

            {/* Delivery Orders */}
            <TabsContent value="delivery" className="mt-0 flex-1">
              {dLoading ? (
                <RowSkeleton />
              ) : dError ? (
                <div className="flex h-36 flex-col items-center justify-center gap-2 text-center">
                  <p className="text-sm text-muted-foreground">{t("error")}</p>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className="cursor-pointer"
                    onClick={() => {
                      void dRefetch();
                    }}
                  >
                    {tCommon("retry")}
                  </Button>
                </div>
              ) : deliveries.length === 0 ? (
                <EmptyState label={t("approvals.empty")} />
              ) : (
                <div className="space-y-2 mt-2">
                  {deliveries.map((d) => (
                    <ApprovalRow
                      key={d.id}
                      code={d.code}
                      date={formatDate(d.delivery_date)}
                      party={d.sales_order?.code}
                      onClick={() => setSelectedDelivery(d)}
                    />
                  ))}
                </div>
              )}
            </TabsContent>

            {/* Customer Invoices */}
            <TabsContent value="invoice" className="mt-0 flex-1">
              {iLoading ? (
                <RowSkeleton />
              ) : iError ? (
                <div className="flex h-36 flex-col items-center justify-center gap-2 text-center">
                  <p className="text-sm text-muted-foreground">{t("error")}</p>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className="cursor-pointer"
                    onClick={() => {
                      void iRefetch();
                    }}
                  >
                    {tCommon("retry")}
                  </Button>
                </div>
              ) : invoices.length === 0 ? (
                <EmptyState label={t("approvals.empty")} />
              ) : (
                <div className="space-y-2 mt-2">
                  {invoices.map((inv) => (
                    <ApprovalRow
                      key={inv.id}
                      code={inv.code}
                      date={formatDate(inv.invoice_date)}
                      party={inv.sales_order?.customer?.name}
                      amount={formatCurrency(inv.amount)}
                      onClick={() => setSelectedInvoice(inv)}
                    />
                  ))}
                </div>
              )}
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>

      {/* Detail modals */}
      {selectedQuotation && (
        <QuotationDetailModal
          open={true}
          onClose={() => setSelectedQuotation(null)}
          quotation={selectedQuotation}
        />
      )}
      {selectedOrder && (
        <OrderDetailModal
          open={true}
          onClose={() => setSelectedOrder(null)}
          order={selectedOrder}
        />
      )}
      {selectedDelivery && (
        <DeliveryDetailModal
          open={true}
          onClose={() => setSelectedDelivery(null)}
          delivery={selectedDelivery}
        />
      )}
      {selectedInvoice && (
        <InvoiceDetailModal
          open={true}
          onClose={() => setSelectedInvoice(null)}
          invoice={selectedInvoice}
        />
      )}
    </>
  );
}
