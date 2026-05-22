"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
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
import { Package } from "lucide-react";
import { useUserPermission } from "@/hooks/use-user-permission";
import { formatCurrency } from "@/lib/utils";
import { DOLinkedDialog } from "@/features/sales/order/components/do-linked-dialog";
import { DOStatusBadge } from "@/features/sales/order/components/do-status-badge";
import { InvoiceLinkedDialog } from "@/features/sales/order/components/invoice-linked-dialog";
import { InvoiceStatusBadge } from "@/features/sales/order/components/invoice-status-badge";
import { OrderDetailModal } from "@/features/sales/order/components/order-detail-modal";
import { OrderStatusBadge } from "@/features/sales/order/components/order-status-badge";
import { useOrders } from "@/features/sales/order/hooks/use-orders";
import type { SalesOrder } from "@/features/sales/order/types";

export function TrackOrderCard() {
  const t = useTranslations("dashboard");
  const tCommon = useTranslations("common");
  const [selectedOrder, setSelectedOrder] = useState<SalesOrder | null>(null);
  const [doDialogOrder, setDoDialogOrder] = useState<SalesOrder | null>(null);
  const [invoiceDialogOrder, setInvoiceDialogOrder] = useState<SalesOrder | null>(null);

  const canViewOrder = useUserPermission("sales_order.read");
  const canViewDO = useUserPermission("delivery_order.read");
  const canViewInvoice = useUserPermission("customer_invoice.read");

  const {
    data,
    isLoading: isOrdersLoading,
    isError: isOrdersError,
    refetch: refetchOrders,
  } = useOrders({
    page: 1,
    per_page: 8,
    sort_by: "created_at",
    sort_dir: "desc",
  });

  const orders = data?.data ?? [];

  const renderDOBadges = (order: SalesOrder) => {
    if (!order.delivery_orders?.length) {
      return <span className="text-muted-foreground text-xs">-</span>;
    }

    return (
      <button
        type="button"
        onClick={canViewDO ? () => setDoDialogOrder(order) : undefined}
        className={canViewDO ? "cursor-pointer" : "cursor-default"}
        title={`${order.delivery_orders.length} Delivery Order(s)`}
      >
        <span className="flex items-center gap-1">
          <DOStatusBadge
            status={order.delivery_orders[0].status}
            className="text-xs font-medium hover:opacity-80 transition-opacity"
          />
          {order.delivery_orders.length > 1 && (
            <span className="text-xs text-muted-foreground">+{order.delivery_orders.length - 1}</span>
          )}
        </span>
      </button>
    );
  };

  const renderInvoiceBadges = (order: SalesOrder) => {
    if (!order.customer_invoices?.length) {
      return <span className="text-muted-foreground text-xs">-</span>;
    }

    return (
      <button
        type="button"
        onClick={canViewInvoice ? () => setInvoiceDialogOrder(order) : undefined}
        className={canViewInvoice ? "cursor-pointer" : "cursor-default"}
        title={`${order.customer_invoices.length} Invoice(s)`}
      >
        <span className="flex items-center gap-1">
          <InvoiceStatusBadge
            status={order.customer_invoices[0].status}
            className="text-xs font-medium hover:opacity-80 transition-opacity"
          />
          {order.customer_invoices.length > 1 && (
            <span className="text-xs text-muted-foreground">+{order.customer_invoices.length - 1}</span>
          )}
        </span>
      </button>
    );
  };

  return (
    <>
      <Card className="h-full">
      <CardHeader>
        <CardTitle>{t("trackOrders.title")}</CardTitle>
        <CardDescription>{t("trackOrders.subtitle")}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("trackOrders.columns.code")}</TableHead>
                <TableHead>{t("trackOrders.columns.status")}</TableHead>
                <TableHead>{t("trackOrders.columns.fulfillment")}</TableHead>
                <TableHead>{t("trackOrders.columns.do")}</TableHead>
                <TableHead>{t("trackOrders.columns.invoice")}</TableHead>
                <TableHead className="text-right">{t("trackOrders.columns.totalAmount")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isOrdersLoading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <TableRow key={i}>
                    {Array.from({ length: 6 }).map((__, j) => (
                      <TableCell key={j}>
                        <Skeleton className="h-4 w-full" />
                      </TableCell>
                    ))}
                  </TableRow>
                ))
              ) : isOrdersError ? (
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
                          void refetchOrders();
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
                      className={canViewOrder ? "font-medium text-primary hover:underline cursor-pointer" : "font-medium"}
                      onClick={() => canViewOrder && setSelectedOrder(order)}
                    >
                      {order.code}
                    </TableCell>
                    <TableCell>
                      <OrderStatusBadge status={order.status} className="text-xs font-medium" />
                    </TableCell>
                    <TableCell>
                      {order.fulfillment ? (
                        <div className="flex flex-col gap-0.5">
                          <div className="flex items-center gap-1 text-xs">
                            <Package className="h-3 w-3 text-muted-foreground" />
                            <span className="font-medium">
                              {order.fulfillment.total_delivered}/{order.fulfillment.total_ordered}
                            </span>
                            <span className="text-muted-foreground">{t("trackOrders.fulfillment.delivered")}</span>
                          </div>
                          {order.fulfillment.total_pending > 0 && (
                            <span className="text-xs text-warning">
                              {order.fulfillment.total_pending} {t("trackOrders.fulfillment.pending")}
                            </span>
                          )}
                          {order.fulfillment.total_remaining > 0 && (
                            <span className="text-xs text-muted-foreground">
                              {order.fulfillment.total_remaining} {t("trackOrders.fulfillment.remaining")}
                            </span>
                          )}
                        </div>
                      ) : (
                        <span className="text-muted-foreground text-xs">-</span>
                      )}
                    </TableCell>
                    <TableCell>{renderDOBadges(order)}</TableCell>
                    <TableCell>{renderInvoiceBadges(order)}</TableCell>
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

      {canViewOrder && selectedOrder && (
        <OrderDetailModal
          open={!!selectedOrder}
          onClose={() => setSelectedOrder(null)}
          order={selectedOrder}
        />
      )}

      {doDialogOrder && (
        <DOLinkedDialog
          salesOrderId={doDialogOrder.id}
          salesOrderCode={doDialogOrder.code}
          open={!!doDialogOrder}
          onOpenChange={(open) => {
            if (!open) setDoDialogOrder(null);
          }}
        />
      )}

      {invoiceDialogOrder && (
        <InvoiceLinkedDialog
          salesOrderId={invoiceDialogOrder.id}
          salesOrderCode={invoiceDialogOrder.code}
          open={!!invoiceDialogOrder}
          onOpenChange={(open) => {
            if (!open) setInvoiceDialogOrder(null);
          }}
        />
      )}
    </>
  );
}
