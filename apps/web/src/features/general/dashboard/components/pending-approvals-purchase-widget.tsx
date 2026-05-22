"use client";

import { useTranslations } from "next-intl";
import { useState } from "react";
import { Clock, ClipboardList, FileCheck, Package, ShoppingBag } from "lucide-react";
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
import { Link } from "@/i18n/routing";
import { useHasPermission } from "@/features/master-data/user-management/hooks/use-has-permission";
import { usePurchaseRequisitions } from "@/features/purchase/requisitions/hooks/use-purchase-requisitions";
import { usePurchaseOrders } from "@/features/purchase/orders/hooks/use-purchase-orders";
import { useGoodsReceipts } from "@/features/purchase/goods-receipt/hooks/use-goods-receipts";
import { useSupplierInvoices } from "@/features/purchase/supplier-invoices/hooks/use-supplier-invoices";
import { formatCurrency, formatDate as formatDateUtil } from "@/lib/utils";

const PER_PAGE = 8;
type PurchaseTab = "requisition" | "order" | "goods-receipt" | "supplier-invoice";

function formatDate(dateStr?: string | null) {
  if (!dateStr) return "-";
  return formatDateUtil(dateStr);
}

function RowSkeleton({ count = 3 }: { count?: number }) {
  return (
    <div className="space-y-3">
      {Array.from({ length: count }).map((_, i) => (
        <Skeleton key={i} className="h-14 w-full rounded-md bg-muted" />
      ))}
    </div>
  );
}

function EmptyState({ label }: { label: string }) {
  return (
    <div className="flex h-36 flex-col items-center justify-center text-muted-foreground">
      <Clock className="mb-2 h-5 w-5 opacity-40" />
      <p className="text-sm">{label}</p>
    </div>
  );
}

interface PurchaseApprovalRowProps {
  href: string;
  code: string;
  date?: string | null;
  party?: string | null;
  amount?: number;
}

function PurchaseApprovalRow({ href, code, date, party, amount }: PurchaseApprovalRowProps) {
  return (
    <Link
      href={href}
      className="hover:bg-muted flex cursor-pointer items-center justify-between rounded-md border px-4 py-3 transition-colors"
    >
      <div className="min-w-0">
        <p className="truncate text-sm font-medium text-foreground">{code}</p>
        <p className="truncate text-xs text-muted-foreground">
          {formatDate(date)}
          {party ? ` · ${party}` : ""}
        </p>
      </div>
      {amount !== undefined && (
        <span className="ml-2 shrink-0 text-sm font-semibold text-foreground">
          {formatCurrency(amount)}
        </span>
      )}
    </Link>
  );
}

export function PendingApprovalsPurchaseWidget() {
  const t = useTranslations("dashboard");
  const tCommon = useTranslations("common");

  // Permissions — always call hooks unconditionally
  const canViewRequisition = useHasPermission("purchase_requisition.read");
  const canViewOrder = useHasPermission("purchase_order.read");
  const canViewGoodsReceipt = useHasPermission("goods_receipt.read");
  const canViewSupplierInvoice = useHasPermission("supplier_invoice.read");
  // approval permission hooks removed (not used after action container removal)

  const defaultTab =
    canViewRequisition
      ? "requisition"
      : canViewOrder
        ? "order"
        : canViewGoodsReceipt
          ? "goods-receipt"
          : "supplier-invoice";

  const [activeTab, setActiveTab] = useState<PurchaseTab>(defaultTab);

  const {
    data: prData,
    isLoading: prLoading,
    isError: prError,
    refetch: prRefetch,
  } = usePurchaseRequisitions(
    {
      status: "SUBMITTED",
      per_page: PER_PAGE,
    },
    { enabled: canViewRequisition },
  );
  const {
    data: poData,
    isLoading: poLoading,
    isError: poError,
    refetch: poRefetch,
  } = usePurchaseOrders(
    {
      status: "SUBMITTED",
      per_page: PER_PAGE,
    },
    { enabled: canViewOrder },
  );
  const {
    data: grData,
    isLoading: grLoading,
    isError: grError,
    refetch: grRefetch,
  } = useGoodsReceipts(
    {
      status: "SUBMITTED",
      per_page: PER_PAGE,
    },
    { enabled: canViewGoodsReceipt },
  );
  const {
    data: siData,
    isLoading: siLoading,
    isError: siError,
    refetch: siRefetch,
  } = useSupplierInvoices(
    {
      status: "SUBMITTED",
      per_page: PER_PAGE,
    },
    { enabled: canViewSupplierInvoice },
  );

  const requisitions = prData?.data ?? [];
  const orders = poData?.data ?? [];
  const goodsReceipts = grData?.data ?? [];
  const supplierInvoices = siData?.data ?? [];

  const hasAnyPermission =
    canViewRequisition ||
    canViewOrder ||
    canViewGoodsReceipt ||
    canViewSupplierInvoice;

  if (!hasAnyPermission) return null;

  return (
    <Card className="flex h-full flex-col gap-0">
      <CardHeader>
        <div className="flex items-start justify-between">
          <div>
            <CardTitle>{t("widgets.pending_approvals_purchase.title")}</CardTitle>
            <CardDescription className="mt-1">
              {t("widgets.pending_approvals_purchase.description")}
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
          <Tabs
          value={activeTab}
          onValueChange={(value) => {
            if (
              value === "requisition" ||
              value === "order" ||
              value === "goods-receipt" ||
              value === "supplier-invoice"
            ) {
              setActiveTab(value);
            }
          }}
          className="pt-0 flex flex-1 flex-col"
        >
          <TabsList className="mb-1 w-full justify-start gap-1 overflow-x-auto border-b border-border/70">
            {canViewRequisition && (
              <TabsTrigger value="requisition" className="gap-1 px-1.5 text-xs">
                <ClipboardList className="h-3 w-3 shrink-0" />
                {t("approvals.tabs.requisitions")}
                {requisitions.length > 0 && (
                  <Badge variant="secondary" className="ml-0.5">
                    {requisitions.length}
                  </Badge>
                )}
              </TabsTrigger>
            )}
            {canViewOrder && (
              <TabsTrigger value="order" className="gap-1 px-1.5 text-xs">
                <ShoppingBag className="h-3 w-3 shrink-0" />
                {t("approvals.tabs.purchaseOrders")}
                {orders.length > 0 && (
                  <Badge variant="secondary" className="ml-0.5">
                    {orders.length}
                  </Badge>
                )}
              </TabsTrigger>
            )}
            {canViewGoodsReceipt && (
              <TabsTrigger value="goods-receipt" className="gap-1 px-1.5 text-xs">
                <Package className="h-3 w-3 shrink-0" />
                {t("approvals.tabs.goodsReceipts")}
                {goodsReceipts.length > 0 && (
                  <Badge variant="secondary" className="ml-0.5">
                    {goodsReceipts.length}
                  </Badge>
                )}
              </TabsTrigger>
            )}
            {canViewSupplierInvoice && (
              <TabsTrigger value="supplier-invoice" className="gap-1 px-1.5 text-xs">
                <FileCheck className="h-3 w-3 shrink-0" />
                {t("approvals.tabs.supplierInvoices")}
                {supplierInvoices.length > 0 && (
                  <Badge variant="secondary" className="ml-0.5">
                    {supplierInvoices.length}
                  </Badge>
                )}
              </TabsTrigger>
            )}
          </TabsList>
          {/* action container removed as requested */}

          {/* Purchase Requisitions */}
          <TabsContent value="requisition" className="mt-0 flex-1">
            {prLoading ? (
              <RowSkeleton />
            ) : prError ? (
              <div className="flex h-36 flex-col items-center justify-center gap-2 text-center">
                <p className="text-sm text-muted-foreground">{t("error")}</p>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="cursor-pointer"
                  onClick={() => {
                    void prRefetch();
                  }}
                >
                  {tCommon("retry")}
                </Button>
              </div>
            ) : requisitions.length === 0 ? (
              <EmptyState label={t("approvals.empty")} />
            ) : (
              <div className="space-y-2">
                {requisitions.map((pr) => (
                  <PurchaseApprovalRow
                    key={pr.id}
                    href="/purchase/purchase-requisitions"
                    code={pr.code}
                    date={pr.request_date}
                    party={pr.supplier?.name}
                    amount={pr.total_amount}
                  />
                ))}
              </div>
            )}
          </TabsContent>

          {/* Purchase Orders */}
          <TabsContent value="order" className="mt-0 flex-1">
            {poLoading ? (
              <RowSkeleton />
            ) : poError ? (
              <div className="flex h-36 flex-col items-center justify-center gap-2 text-center">
                <p className="text-sm text-muted-foreground">{t("error")}</p>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="cursor-pointer"
                  onClick={() => {
                    void poRefetch();
                  }}
                >
                  {tCommon("retry")}
                </Button>
              </div>
            ) : orders.length === 0 ? (
              <EmptyState label={t("approvals.empty")} />
            ) : (
              <div className="space-y-2 mt-2">
                {orders.map((po) => (
                  <PurchaseApprovalRow
                    key={po.id}
                    href="/purchase/purchase-orders"
                    code={po.code}
                    date={po.order_date}
                    party={po.supplier?.name}
                    amount={po.total_amount}
                  />
                ))}
              </div>
            )}
          </TabsContent>

          {/* Goods Receipts */}
          <TabsContent value="goods-receipt" className="mt-0 flex-1">
            {grLoading ? (
              <RowSkeleton />
            ) : grError ? (
              <div className="flex h-36 flex-col items-center justify-center gap-2 text-center">
                <p className="text-sm text-muted-foreground">{t("error")}</p>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="cursor-pointer"
                  onClick={() => {
                    void grRefetch();
                  }}
                >
                  {tCommon("retry")}
                </Button>
              </div>
            ) : goodsReceipts.length === 0 ? (
              <EmptyState label={t("approvals.empty")} />
            ) : (
              <div className="space-y-2 mt-2">
                {goodsReceipts.map((gr) => (
                  <PurchaseApprovalRow
                    key={gr.id}
                    href="/purchase/goods-receipt"
                    code={gr.code}
                    date={gr.receipt_date}
                    party={gr.supplier?.name}
                  />
                ))}
              </div>
            )}
          </TabsContent>

          {/* Supplier Invoices */}
          <TabsContent value="supplier-invoice" className="mt-0 flex-1">
            {siLoading ? (
              <RowSkeleton />
            ) : siError ? (
              <div className="flex h-36 flex-col items-center justify-center gap-2 text-center">
                <p className="text-sm text-muted-foreground">{t("error")}</p>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="cursor-pointer"
                  onClick={() => {
                    void siRefetch();
                  }}
                >
                  {tCommon("retry")}
                </Button>
              </div>
            ) : supplierInvoices.length === 0 ? (
              <EmptyState label={t("approvals.empty")} />
            ) : (
              <div className="space-y-2 mt-2">
                {supplierInvoices.map((si) => (
                  <PurchaseApprovalRow
                    key={si.id}
                    href="/purchase/supplier-invoices"
                    code={si.code}
                    date={si.invoice_date}
                    party={si.supplier_name}
                    amount={si.amount}
                  />
                ))}
              </div>
            )}
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  );
}
