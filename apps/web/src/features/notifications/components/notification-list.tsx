"use client";

import { Bell, Check, CheckCheck, Copy } from "lucide-react";
import { formatDistanceToNow } from "date-fns";
import { MouseEvent, useState } from "react";
import { useTranslations } from "next-intl";

import { useRouter } from "@/i18n/routing";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { setPasswordResetTokenPrefill } from "@/lib/password-reset-token-prefill";

import {
  useMarkAsRead,
  useNotifications,
} from "../hooks/use-notifications";
import type { Notification } from "../types";

function extractResetToken(message: string): string | null {
  const match = message.match(/token\s*:\s*([A-Za-z0-9._-]+)/i);
  return match?.[1] ?? null;
}

function getNotificationTimestamp(createdAt: string): number {
  const parsed = Date.parse(createdAt);
  return Number.isFinite(parsed) ? parsed : 0;
}

export function NotificationList() {
  const t = useTranslations("notifications");
  const router = useRouter();
  const [page, setPage] = useState(1);

  const {
    data: response,
    isLoading,
    isError,
    error,
    refetch,
  } = useNotifications({
    page,
    per_page: 20,
  });

  const markAsRead = useMarkAsRead();
  const notifications = response?.data ?? [];
  const pagination = response?.meta?.pagination;

  const handleMarkAsRead = async (id: string) => {
    await markAsRead.mutateAsync(id);
  };

  const resolveEntityLink = (notification: Notification): string | null => {
    if (notification.entity_link) {
      return notification.entity_link;
    }

    switch (notification.entity_type) {
      case "sales_quotation":
        return `/sales/quotations?open_quotation=${notification.entity_id}`;
      case "sales_order":
        return `/sales/orders?open_order=${notification.entity_id}`;
      case "purchase_requisition":
        return `/purchase/purchase-requisitions?open_purchase_requisition=${notification.entity_id}`;
      case "purchase_order":
        return `/purchase/purchase-orders?open_purchase_order=${notification.entity_id}`;
      case "delivery_order":
        return `/sales/delivery-orders?open_delivery_order=${notification.entity_id}`;
      case "goods_receipt":
        return `/purchase/goods-receipt?open_goods_receipt=${notification.entity_id}`;
      case "supplier_invoice":
        return `/purchase/supplier-invoices?open_supplier_invoice=${notification.entity_id}`;
      case "supplier_invoice_dp":
        return `/purchase/supplier-invoice-down-payments?open_supplier_invoice_dp=${notification.entity_id}`;
      case "customer_invoice":
        return `/sales/invoices?open_customer_invoice=${notification.entity_id}`;
      case "customer_invoice_dp":
        return `/sales/customer-invoice-down-payments?open_customer_invoice_dp=${notification.entity_id}`;
      case "non_trade_payable":
        return `/finance/ap/non-trade-payables?open_non_trade_payable=${notification.entity_id}`;
      case "payment":
        return `/finance/ap/payments?open_payment=${notification.entity_id}`;
      case "budget":
        return `/finance/budget?open_budget=${notification.entity_id}`;
      case "financial_closing":
        return `/finance/closing?open_financial_closing=${notification.entity_id}`;
      case "asset_maintenance":
        return `/finance/asset-maintenance?open_asset_maintenance=${notification.entity_id}`;
      case "travel_plan":
        return `/travel/travel-planner?open_trip=${notification.entity_id}`;
      case "leave_request":
        return `/hrd/leave-requests?open_leave_request=${notification.entity_id}`;
      case "overtime":
        return `/hrd/overtime?open_overtime=${notification.entity_id}`;
      case "recruitment":
        return `/hrd/recruitment?open_recruitment=${notification.entity_id}`;
      case "crm_visit":
        return `/crm/visits?open_crm_visit=${notification.entity_id}`;
      case "company":
        return `/master-data/company?open_company=${notification.entity_id}`;
      case "employee":
        return `/master-data/employees?open_employee=${notification.entity_id}`;
      case "supplier":
        return `/master-data/suppliers?open_supplier=${notification.entity_id}`;
      case "customer":
        return `/master-data/customers?open_customer=${notification.entity_id}`;
      case "product":
        return `/master-data/products?open_product=${notification.entity_id}`;
      case "stock_opname":
        return `/stock/opname?open_stock_opname=${notification.entity_id}`;
      case "salary":
        return `/hrd/salary-structures?open_salary=${notification.entity_id}`;
      case "password_reset_request":
        return `/master-data/users?open_user=${notification.entity_id}`;
      case "pos_self_order":
        return `/pos/fb/live-table?outlet_id=${notification.entity_id}`;
      case "pos_order":
      case "pos_payment":
        return `/pos/fb/live-table`;
      default:
        return null;
    }
  };

  const handleOpenNotification = async (notification: Notification) => {
    // Auto mark as read jika belum dibaca
    if (!notification.is_read) {
      try {
        await markAsRead.mutateAsync(notification.id);
      } catch {
        // Biarkan navigasi tetap jalan meski gagal mark as read
      }
    }

    if (notification.entity_type === "password_reset_request") {
      const token = extractResetToken(notification.message);
      if (token) {
        setPasswordResetTokenPrefill({
          userId: notification.entity_id,
          token,
          createdAt: getNotificationTimestamp(notification.created_at),
        });
      }
      router.push(`/master-data/users?reset_user=${notification.entity_id}&open_change_password=1`);
      return;
    }

    const path = resolveEntityLink(notification);
    if (path) {
      router.push(path);
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-3">
        {[...Array(5)].map((_, i) => (
          <Card key={i}>
            <CardContent className="p-4">
              <Skeleton className="mb-2 h-4 w-3/4" />
              <Skeleton className="h-3 w-1/2" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  if (isError) {
    return (
      <Card>
        <CardContent className="p-8 text-center">
          <p className="mb-4 text-sm text-destructive">
            {error instanceof Error ? error.message : t("errorLoading")}
          </p>
          <Button onClick={() => refetch()} variant="outline" size="sm" className="cursor-pointer">
            {t("retry")}
          </Button>
        </CardContent>
      </Card>
    );
  }

  if (notifications.length === 0) {
    return (
      <Card>
        <CardContent className="p-8 text-center">
          <div className="mb-3 inline-block rounded-full bg-muted p-3">
            <Bell className="h-5 w-5 text-muted-foreground" />
          </div>
          <p className="text-sm text-muted-foreground">{t("empty")}</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-4">

      <div className="space-y-3">
        {notifications.map((notification) => (
          <NotificationItem
            key={notification.id}
            notification={notification}
            onMarkAsRead={handleMarkAsRead}
            onOpen={handleOpenNotification}
          />
        ))}
      </div>

      {pagination && pagination.total_pages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            {t("showing", {
              from: (pagination.page - 1) * pagination.per_page + 1,
              to: Math.min(pagination.page * pagination.per_page, pagination.total),
              total: pagination.total,
            })}
          </p>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((prev) => Math.max(1, prev - 1))}
              disabled={!pagination.has_prev || page === 1}
              className="cursor-pointer"
            >
              {t("previous")}
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((prev) => prev + 1)}
              disabled={!pagination.has_next}
              className="cursor-pointer"
            >
              {t("next")}
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}

interface NotificationItemProps {
  readonly notification: Notification;
  readonly onMarkAsRead: (id: string) => void;
  readonly onOpen: (notification: Notification) => void;
}

function NotificationItem({ notification, onMarkAsRead, onOpen }: NotificationItemProps) {
  const t = useTranslations("notifications");
  const markAsRead = useMarkAsRead();
  const router = useRouter();
  const [copiedToken, setCopiedToken] = useState(false);

  const timeAgo = formatDistanceToNow(new Date(notification.created_at), {
    addSuffix: true,
  });

  const isPasswordResetRequest = notification.entity_type === "password_reset_request";

  const rawToken = isPasswordResetRequest ? extractResetToken(notification.message) : null;

  const formatTokenPreview = (token: string) => {
    if (token.length <= 20) return token;
    return `${token.slice(0, 10)}...${token.slice(-6)}`;
  };

  const compactMessage = () => {
    if (!isPasswordResetRequest) return notification.message;
    if (!rawToken) return notification.message;
    return notification.message.replace(/token\s*:\s*[A-Za-z0-9._-]+/i, "Token: ••••••••");
  };

  const handleMarkAsRead = () => {
    if (!notification.is_read) {
      onMarkAsRead(notification.id);
    }
  };

  const handleCopyToken = async (event: MouseEvent<HTMLButtonElement>) => {
    event.stopPropagation();
    if (!rawToken) return;

    try {
      await navigator.clipboard.writeText(rawToken);
      setCopiedToken(true);
      setTimeout(() => setCopiedToken(false), 1500);

      setPasswordResetTokenPrefill({
        userId: notification.entity_id,
        token: rawToken,
        createdAt: getNotificationTimestamp(notification.created_at),
      });

      if (!notification.is_read) {
        await markAsRead.mutateAsync(notification.id);
      }

      router.push(`/master-data/users?reset_user=${notification.entity_id}&open_change_password=1`);
    } catch {
      // Keep silent and rely on existing UI behavior.
    }
  };
  return (
    <Card
      className={`cursor-pointer transition-all hover:shadow-sm ${
        !notification.is_read ? "border-primary/20 bg-primary/5" : ""
      }`}
      onClick={() => onOpen(notification)}
    >
      <CardContent className="px-4">
        <div className="flex items-start justify-between gap-4">
          <div className="flex-1 space-y-1">
            <div className="flex items-start gap-2">
              {!notification.is_read && (
                <span className="mt-1.5 h-2 w-2 shrink-0 rounded-full bg-primary" />
              )}
              <div className="flex-1 min-w-0">
                <h4 className={`text-sm font-medium ${!notification.is_read ? "font-semibold" : ""}`}>
                  {notification.title}
                </h4>
                <p className="mt-1 text-sm text-muted-foreground wrap-break-word line-clamp-2">{compactMessage()}</p>
                {isPasswordResetRequest && rawToken && (
                  <div className="mt-2 flex items-center gap-2">
                    <span className="rounded-md bg-muted/70 px-2 py-0.5 font-mono text-xs text-foreground break-all">
                      {formatTokenPreview(rawToken)}
                    </span>
                    <Button
                      variant="ghost"
                      size="icon-sm"
                      onClick={handleCopyToken}
                      className="h-7 w-7 cursor-pointer"
                      title={copiedToken ? t("tokenCopied") : t("copyToken")}
                    >
                      {copiedToken ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
                    </Button>
                  </div>
                )}
                <p className="mt-2 text-xs text-muted-foreground">{timeAgo}</p>
              </div>
            </div>
          </div>

          {!notification.is_read && (
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={(event) => {
                event.stopPropagation();
                handleMarkAsRead();
              }}
              disabled={markAsRead.isPending}
              title={t("markAsRead")}
              className="h-8 w-8 cursor-pointer"
            >
              <CheckCheck className="h-3.5 w-3.5" />
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
