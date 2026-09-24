"use client";

import React, { useState } from "react";
import Image from "next/image";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import { useBuyerTransactions } from "@/features/buyer/transactions/hooks/useBuyerTransactions";
import { BuyerLayout } from "../../components/buyer-layout";
import { TransactionItem } from "@/features/buyer/transactions/types/transaction.types";
import { formatCurrency } from "@/lib/utils";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Search, Loader2, Calendar, FileText, ShoppingBag, Eye, MessageSquare, CheckCircle2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { TransactionStatusBadge } from "./transaction-status-badge";
import { BuyerReviewDialog } from "./buyer-review-dialog";

export function BuyerTransactionsPage() {
  const t = useTranslations("buyer.transactions");
  const [activeTab, setActiveTab] = useState<string>("all");
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [currentPage, setCurrentPage] = useState<number>(1);
  const [reviewTarget, setReviewTarget] = useState<TransactionItem | null>(null);
  const [isReviewOpen, setIsReviewOpen] = useState<boolean>(false);
  const itemsPerPage = 5;

  const { data, isLoading } = useBuyerTransactions({
    page: currentPage,
    per_page: itemsPerPage,
    status: activeTab,
  });

  const statusTabs = [
    { id: "all", label: t("tabAll") },
    { id: "pending", label: t("tabUnpaid") },
    { id: "processing", label: t("tabProcessing") },
    { id: "shipped", label: t("tabShipped") },
    { id: "completed", label: t("tabCompleted") },
    { id: "cancelled", label: t("tabCancelled") },
  ];

  // Client-side search filter simulation for search input
  const filteredItems = data?.items.filter((item: TransactionItem) =>
    item.product_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    item.po_number.toLowerCase().includes(searchQuery.toLowerCase()) ||
    item.supplier_name.toLowerCase().includes(searchQuery.toLowerCase())
  ) || [];

  const handleTabChange = (tabId: string) => {
    setActiveTab(tabId);
    setCurrentPage(1);
  };

  const renderContent = () => {
    if (isLoading) {
      return (
        <div className="flex flex-col items-center justify-center py-20 gap-3">
          <Loader2 className="h-8 w-8 text-primary animate-spin" />
          <p className="text-sm text-muted-foreground">{t("loading")}</p>
        </div>
      );
    }

    if (filteredItems.length === 0) {
      return (
        <div className="bg-card rounded-xl border border-border p-12 text-center shadow-xs flex flex-col items-center justify-center gap-4">
          <div className="bg-muted p-4 rounded-full text-muted-foreground">
            <ShoppingBag className="h-8 w-8" />
          </div>
          <div className="space-y-1">
            <h3 className="font-bold text-base text-foreground">{t("empty")}</h3>
            <p className="text-xs text-muted-foreground">{t("emptyDesc")}</p>
          </div>
        </div>
      );
    }

    return (
      <div className="space-y-4">
        {filteredItems.map((tx: TransactionItem) => (
          <div
            key={tx.id}
            className="bg-card rounded-xl border border-border p-5 shadow-xs space-y-4 hover:shadow-md transition-all duration-300"
          >
            {/* Card Header (Date, PO number, Unified Status Badge) */}
            <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border pb-3">
              <div className="flex items-center gap-3 text-xs text-muted-foreground">
                <span className="flex items-center gap-1">
                  <Calendar className="h-3.5 w-3.5" />
                  {new Date(tx.created_at).toLocaleDateString("id-ID", {
                    year: "numeric",
                    month: "short",
                    day: "numeric",
                  })}
                </span>
                <span className="h-3 w-px bg-border" />
                <span className="flex items-center gap-1 font-semibold text-foreground">
                  <FileText className="h-3.5 w-3.5" />
                  {tx.po_number}
                </span>
              </div>

              {/* 1 Unified Status Badge */}
              <div className="flex items-center gap-2">
                <TransactionStatusBadge status={tx.status} paymentStatus={tx.payment_status} />
              </div>
            </div>

            {/* Card Body (Real Product Image, Supplier info, Product detail) */}
            <div className="flex items-start gap-4">
              <div className="relative h-14 w-14 sm:h-16 sm:w-16 rounded-lg border border-border bg-muted/30 overflow-hidden shrink-0">
                {tx.product_image ? (
                  <Image
                    src={tx.product_image}
                    alt={tx.product_name}
                    fill
                    sizes="64px"
                    className="object-cover"
                  />
                ) : (
                  <div className="h-full w-full flex items-center justify-center bg-primary/10 text-primary">
                    <ShoppingBag className="h-6 w-6" />
                  </div>
                )}
              </div>
              <div className="min-w-0 flex-1 space-y-1">
                <h4 className="font-extrabold text-sm text-foreground truncate">
                  {tx.product_name}
                </h4>
                <p className="text-xs text-muted-foreground font-semibold">
                  {t("supplier")}: <span className="text-foreground">{tx.supplier_name}</span>
                </p>
                <p className="text-xs text-muted-foreground font-medium">
                  {tx.quantity_value} {tx.quantity_unit} x {formatCurrency(tx.price_per_unit)}
                </p>
              </div>
              <div className="text-right space-y-0.5 shrink-0">
                <p className="text-xs text-muted-foreground font-semibold">{t("total")}</p>
                <p className="text-base font-extrabold text-foreground">{formatCurrency(tx.total_amount)}</p>
              </div>
            </div>

            {/* Card Footer (Actions) */}
            <div className="flex items-center justify-between gap-4 pt-1">
              <p className="text-xs text-muted-foreground truncate max-w-[50%]">
                {tx.notes && `${t("notes")}: "${tx.notes}"`}
              </p>

              <div className="flex items-center gap-2">
                <Link href={`/chat`}>
                  <Button variant="outline" size="sm" className="h-8.5 text-xs font-semibold cursor-pointer">
                    <MessageSquare className="mr-1.5 h-3.5 w-3.5" />
                    {t("chatBtn")}
                  </Button>
                </Link>

                <Link href={`/transactions/${tx.id}`}>
                  <Button variant="outline" size="sm" className="h-8.5 text-xs font-semibold cursor-pointer">
                    <Eye className="mr-1.5 h-3.5 w-3.5" />
                    {t("detailBtn")}
                  </Button>
                </Link>

                {tx.status === "completed" && (
                  tx.has_reviewed ? (
                    <Badge
                      variant="outline"
                      className="h-8.5 px-3 text-xs font-semibold border-success/30 bg-success/10 text-success gap-1.5"
                    >
                      <CheckCircle2 className="h-3.5 w-3.5" />
                      {t("alreadyReviewed")}
                    </Badge>
                  ) : (
                    <Button
                      size="sm"
                      onClick={() => {
                        setReviewTarget(tx);
                        setIsReviewOpen(true);
                      }}
                      className="h-8.5 text-xs font-semibold cursor-pointer bg-primary text-primary-foreground hover:-translate-y-0.5 active:translate-y-0"
                    >
                      {t("reviewBtn")}
                    </Button>
                  )
                )}
              </div>
            </div>
          </div>
        ))}
      </div>
    );
  };

  return (
    <BuyerLayout>
      <div className="space-y-6">
        {/* Title Header */}
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">{t("title")}</h1>
          <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
        </div>

        {/* Filter Tabs & Search */}
        <div className="space-y-4">
          {/* Navigation Tabs (Tokopedia-style horizontal tabs) */}
          <div className="flex items-center gap-1.5 border-b border-border overflow-x-auto pb-1 scrollbar-none">
            {statusTabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => handleTabChange(tab.id)}
                className={cn(
                  "px-4 py-2 text-sm font-medium transition-all duration-200 whitespace-nowrap cursor-pointer hover:text-primary hover:-translate-y-0.5 active:translate-y-0",
                  activeTab === tab.id
                    ? "text-primary font-semibold"
                    : "text-muted-foreground"
                )}
              >
                {tab.label}
              </button>
            ))}
          </div>

          {/* Search Input */}
          <div className="relative">
            <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 h-4.5 w-4.5 text-muted-foreground" />
            <Input
              type="text"
              placeholder={t("searchPlaceholder")}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10 h-10.5 rounded-lg border-border focus-visible:ring-primary focus-visible:border-primary text-sm font-medium bg-card"
            />
          </div>
        </div>

        {/* Transactions List Content */}
        {renderContent()}

        {/* Pagination Controls */}
        {data && data.total > itemsPerPage && (
          <div className="flex items-center justify-between border-t border-border pt-4">
            <p className="text-xs text-muted-foreground font-semibold">
              {t("showing")} <span className="font-bold">{(currentPage - 1) * itemsPerPage + 1}</span> {t("to")}{" "}
              <span className="font-bold">
                {Math.min(currentPage * itemsPerPage, data.total)}
              </span>{" "}
              {t("of")} <span className="font-bold">{data.total}</span> {t("entries")}
            </p>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={currentPage === 1}
                onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                className="h-8 text-xs font-semibold cursor-pointer"
              >
                {t("previous")}
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={currentPage * itemsPerPage >= data.total}
                onClick={() => setCurrentPage((p) => p + 1)}
                className="h-8 text-xs font-semibold cursor-pointer"
              >
                {t("next")}
              </Button>
            </div>
          </div>
        )}
      </div>

      <BuyerReviewDialog
        open={isReviewOpen}
        onOpenChange={setIsReviewOpen}
        transaction={reviewTarget}
      />
    </BuyerLayout>
  );
}
