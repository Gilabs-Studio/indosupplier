"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { Package } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { useUserPermission } from "@/hooks/use-user-permission";
import { formatCurrency } from "@/lib/utils";
import { GoodsReceiptStatusBadge } from "@/features/purchase/goods-receipt/components/goods-receipt-status-badge";
import { SupplierInvoiceStatusBadge } from "@/features/purchase/supplier-invoices/components/supplier-invoice-status-badge";
import { GRLinkedDialog } from "@/features/purchase/orders/components/gr-linked-dialog";
import { PurchaseOrderDetail } from "@/features/purchase/orders/components/purchase-order-detail";
import { PurchaseOrderStatusBadge } from "@/features/purchase/orders/components/purchase-order-status-badge";
import { SILinkedDialog } from "@/features/purchase/orders/components/si-linked-dialog";
import { usePurchaseOrders } from "@/features/purchase/orders/hooks/use-purchase-orders";
import type { PurchaseOrderListItem } from "@/features/purchase/orders/types";

export function TrackPurchaseOrderCard() {
  const t = useTranslations("dashboard");
  const tCommon = useTranslations("common");

  const [selectedOrderId, setSelectedOrderId] = useState<string | null>(null);
  const [grDialogItem, setGrDialogItem] = useState<PurchaseOrderListItem | null>(null);
  const [siDialogItem, setSiDialogItem] = useState<PurchaseOrderListItem | null>(null);

  const canViewPO = useUserPermission("purchase_order.read");
  const canViewGR = useUserPermission("goods_receipt.read");
  const canViewSI = useUserPermission("supplier_invoice.read");

  const {
    data,
    isLoading: isPOsLoading,
    isError: isPOsError,
    refetch: refetchPOs,
  } = usePurchaseOrders({
    page: 1,
    per_page: 8,
    sort_by: "created_at",
    sort_dir: "desc",
  });

  const orders = data?.data ?? [];

  const renderGRBadges = (order: PurchaseOrderListItem) => {
    if (!order.goods_receipts?.length) {
      return <span className="text-muted-foreground text-xs">-</span>;
    }

    return (
      <button
        type="button"
        onClick={canViewGR ? () => setGrDialogItem(order) : undefined}
        className={canViewGR ? "cursor-pointer" : "cursor-default"}
        title={`${order.goods_receipts.length} Goods Receipt(s)`}
      >
        <span className="flex items-center gap-1">
          <GoodsReceiptStatusBadge
            status={order.goods_receipts[0].status}
            className="text-xs font-medium hover:opacity-80 transition-opacity"
          />
          {order.goods_receipts.length > 1 && (
            <span className="text-xs text-muted-foreground">+{order.goods_receipts.length - 1}</span>
          )}
        </span>
      </button>
    );
  };

  const renderSIBadges = (order: PurchaseOrderListItem) => {
    if (!order.supplier_invoices?.length) {
      return <span className="text-muted-foreground text-xs">-</span>;
    }

    return (
      <button
        type="button"
        onClick={canViewSI ? () => setSiDialogItem(order) : undefined}
        className={canViewSI ? "cursor-pointer" : "cursor-default"}
        title={`${order.supplier_invoices.length} Supplier Invoice(s)`}
      >
        <span className="flex items-center gap-1">
          <SupplierInvoiceStatusBadge
            status={order.supplier_invoices[0].status}
            className="text-xs font-medium hover:opacity-80 transition-opacity"
          />
          {order.supplier_invoices.length > 1 && (
            <span className="text-xs text-muted-foreground">+{order.supplier_invoices.length - 1}</span>
          )}
        </span>
      </button>
    );
  };

  return (
    <>
      <Card className="h-full">
        <CardHeader>
          <CardTitle>{t("trackPurchaseOrders.title")}</CardTitle>
          <CardDescription>{t("trackPurchaseOrders.subtitle")}</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("trackPurchaseOrders.columns.code")}</TableHead>
                  <TableHead>{t("trackPurchaseOrders.columns.status")}</TableHead>
                  <TableHead>{t("trackPurchaseOrders.columns.fulfillment")}</TableHead>
                  <TableHead>{t("trackPurchaseOrders.columns.gr")}</TableHead>
                  <TableHead>{t("trackPurchaseOrders.columns.invoice")}</TableHead>
                  <TableHead className="text-right">{t("trackPurchaseOrders.columns.totalAmount")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {isPOsLoading ? (
                  Array.from({ length: 5 }).map((_, i) => (
                    <TableRow key={i}>
                      {Array.from({ length: 6 }).map((__, j) => (
                        <TableCell key={j}>
                          <Skeleton className="h-4 w-full" />
                        </TableCell>
                      ))}
                    </TableRow>
                  ))
                ) : isPOsError ? (
                  <TableRow>
                    <TableCell colSpan={6} className="py-8 text-center">
                      <div className="space-y-2">
                        <p className="text-sm text-muted-foreground">{t("error")}</p>
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          className="cursor-pointer"
                          onClick={() => {
                            void refetchPOs();
                          }}
                        >
                          {tCommon("retry")}
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ) : orders.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center py-8 text-muted-foreground">
                      {t("noData")}
                    </TableCell>
                  </TableRow>
                ) : (
                  orders.map((order) => (
                    <TableRow key={order.id}>
                      <TableCell
                        className={canViewPO ? "font-medium text-primary hover:underline cursor-pointer" : "font-medium"}
                        onClick={() => canViewPO && setSelectedOrderId(order.id)}
                      >
                        {order.code}
                      </TableCell>
                      <TableCell>
                        <PurchaseOrderStatusBadge status={order.status ?? ""} className="text-xs font-medium" />
                      </TableCell>
                      <TableCell>
                        {order.fulfillment ? (
                          <div className="flex flex-col gap-0.5">
                            <div className="flex items-center gap-1 text-xs">
                              <Package className="h-3 w-3 text-muted-foreground" />
                              <span className="font-medium">
                                {order.fulfillment.total_received}/{order.fulfillment.total_ordered}
                              </span>
                              <span className="text-muted-foreground">{t("trackPurchaseOrders.fulfillment.received")}</span>
                            </div>
                            {order.fulfillment.total_pending > 0 && (
                              <span className="text-xs text-warning">
                                {order.fulfillment.total_pending} {t("trackPurchaseOrders.fulfillment.pending")}
                              </span>
                            )}
                            {order.fulfillment.total_remaining > 0 && (
                              <span className="text-xs text-muted-foreground">
                                {order.fulfillment.total_remaining} {t("trackPurchaseOrders.fulfillment.remaining")}
                              </span>
                            )}
                          </div>
                        ) : (
                          <span className="text-muted-foreground text-xs">-</span>
                        )}
                      </TableCell>
                      <TableCell>{renderGRBadges(order)}</TableCell>
                      <TableCell>{renderSIBadges(order)}</TableCell>
                      <TableCell className="text-right font-mono font-medium">
                        {formatCurrency(order.total_amount ?? 0)}
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      <PurchaseOrderDetail
        open={!!selectedOrderId}
        onClose={() => setSelectedOrderId(null)}
        purchaseOrderId={selectedOrderId}
      />

      {grDialogItem && (
        <GRLinkedDialog
          purchaseOrderCode={grDialogItem.code}
          items={grDialogItem.goods_receipts ?? []}
          open={!!grDialogItem}
          onOpenChange={(open) => {
            if (!open) setGrDialogItem(null);
          }}
        />
      )}

      {siDialogItem && (
        <SILinkedDialog
          purchaseOrderCode={siDialogItem.code}
          purchaseOrderId={siDialogItem.id}
          open={!!siDialogItem}
          onOpenChange={(open) => {
            if (!open) setSiDialogItem(null);
          }}
        />
      )}
    </>
  );
}
